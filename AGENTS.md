# AGENTS.md

Small Go 1.26 app: Yahoo Fantasy waiver assistant. Server-rendered HTML (Go
`html/template`), stdlib only — `go.mod` has **no third-party modules**, and adding
one should be a deliberate decision, not a convenience.

## Commands

Run from the repo root: `.env` and `internal/home/templates/...` are resolved
relative to the working directory.

```bash
go run .                                  # serve on :8080 (hardcoded)
air                                       # hot reload via .air.toml -> ./tmp/main
go test ./...                             # all tests, no network
go test ./internal/yahoo -run TestLeague -v
go vet ./...
```

There is no Makefile, CI, linter config, or codegen. `tmp/` is build output.

## Layout

```
main.go                    server wiring, /login and /callback, :8080
internal/home              home page handler; template embedded with go:embed
internal/shared            package name is "config" (dir is "shared")
internal/auth              Yahoo OAuth: AuthCodeURL, Exchange, Refresh, Save
internal/yahoo             read-only Fantasy Sports API client (XML)
docs/yahoo-api-shapes.md   field-level reference for league/player resources
```

- `internal/shared` is **`package config`**; imports must alias it:
  `config "fantasy_waiver_assistant/internal/shared"`.
- Import paths use the module name `fantasy_waiver_assistant` (underscore).

## Hard-won constraints

- **Yahoo API is XML-only by design.** `internal/yahoo` sends
  `Accept: application/xml` and decodes `fantasy_content` envelopes. JSON quirks are
  documented in `docs/yahoo-api-shapes.md` for reference only; do not add a JSON
  decoder.
- **Read `docs/yahoo-api-shapes.md` before adding fields/types.** It records wire
  names, nesting, and quirks (string-valued stats, `-` for DNP, 25 players/page cap,
  lowercase `l` in `470.l.123` keys).
- Build keys and paths with the helpers in `internal/yahoo/path.go`. Position values
  contain `/`, so `escapeSegment` deliberately escapes only `% ? # space` — do not
  swap in full `url.PathEscape`.
- When Yahoo errors arrive, `client.go` returns `yahoo.ErrUnauthorized` on HTTP 401
  and unwraps `<error>` documents (including HTTP 200 error bodies) into
  `*yahoo.APIError`. Nothing refreshes and retries yet; that belongs in the caller.
- `auth.Client` keeps its token **in memory only**. `/callback` writes
  `.tokens.json` (0600) but startup never loads it, so a restart means logged out.
  There is also **no adapter** wiring `auth.Client` into `yahoo.Client`'s
  `TokenSource` — write one when adding authenticated Yahoo calls.
- `internal/home` embeds its template with `//go:embed` and parses it once, so the
  binary does not depend on the working directory. Keep that pattern for new
  handlers; do not reintroduce repo-root-relative `template.ParseFiles(...)`.
- Config: `config.Load()` reads `.env` from the cwd, never overwrites variables
  already set (real env wins), and reports all missing required vars in one error.
  Required: `YAHOO_CLIENT_ID`, `YAHOO_CLIENT_SECRET`, `YAHOO_REDIRECT_URI`.
- `YAHOO_REDIRECT_URI` must match the Yahoo app dashboard character for character;
  local dev uses an ngrok URL (see `.env.example`).

## Tests

- All tests are hermetic: `httptest` servers plus fixtures in
  `internal/yahoo/testdata/`, no live Yahoo calls. Pin new response shapes as
  fixtures instead of hitting the API.
- Fixtures are opened as `testdata/...`, which resolves because `go test` runs in the
  package directory. Keep that convention.
- `config.Load` tests use `t.Setenv`; the repo `.env` is not loaded during tests
  (package cwd), so do not assume it is.
- `.env` and `.tokens.json` are gitignored secrets. Never commit them.
