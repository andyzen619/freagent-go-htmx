# AGENTS.md

Small Go 1.26 app: Yahoo Fantasy waiver assistant. Server-rendered HTML (Go
`html/template`) with htmx for partial updates, stdlib only — `go.mod` has **no
third-party modules**, and adding one should be a deliberate decision, not a
convenience. htmx is a pinned CDN `<script>` in the shared layout; there is no JS
build step.

## Commands

Run from the repo root: `config.Load` resolves `.env` relative to the working
directory. Templates are embedded in the binary, not read from disk.

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
internal/web               shared layout + Parse/Render/Fragment (page vs fragment)
internal/home              home page handler
internal/players           free-agent page, /hx/players/rows fragment, Yahoo source
internal/middleware        RequireSession wrapper for session-gated routes
internal/shared            package name is "config" (dir is "shared")
internal/auth              Yahoo OAuth: AuthCodeURL, Exchange, Refresh, Save
internal/yahoo             read-only Fantasy Sports API client (XML)
docs/yahoo-api-shapes.md   field-level reference for league/player resources
```

- `internal/shared` is **`package config`**; imports must alias it:
  `config "fantasy_waiver_assistant/internal/shared"`.
- Import paths use the module name `fantasy_waiver_assistant` (underscore).

## Routing and htmx

- Pages use clean paths (`/`, `/players`); htmx fragments live under `/hx/...`.
  A component route never renders the layout, even on a direct visit.
- Routes registered on the `app` mux in `main.go` are gated by
  `middleware.RequireSession`; `/login` and `/callback` stay public on the outer mux.
- `web.Render` writes the layout for direct visits and the named fragment when
  `HX-Request: true`; `web.Fragment` always writes the fragment. Both set
  `Vary: HX-Request`, so a cached fragment is never served as a full page.
- Feature templates define `content` plus their fragments and are parsed once at
  startup with `web.Parse(featureFS, patterns...)`, which merges the shared layout.
- Filter links set `hx-push-url` to the page URL (`/players?pos=PG`), never to the
  fragment URL, so reloads and the back button land on a full page.

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
  `main.go` adapts it to `yahoo.Client`'s `TokenSource`; reuse that adapter for new
  authenticated Yahoo calls instead of adding a second one.
- Every feature embeds its templates with `//go:embed` and parses them once at
  startup via `web.Parse`, so the binary does not depend on the working directory.
  Do not reintroduce repo-root-relative `template.ParseFiles(...)`.
- Config: `config.Load()` reads `.env` from the cwd, never overwrites variables
  already set (real env wins), and reports all missing required vars in one error.
  Required: `YAHOO_CLIENT_ID`, `YAHOO_CLIENT_SECRET`, `YAHOO_REDIRECT_URI`.
  Optional: `YAHOO_LEAGUE_KEY` (e.g. `470.l.123`), used by the players page.
- `YAHOO_REDIRECT_URI` must match the Yahoo app dashboard character for character;
  local dev uses an ngrok URL (see `.env.example`).

## UI

The app is used on phones first. Every template must be usable at ~360px wide
before it looks right on desktop.

- Use fluid widths only; no fixed widths that force horizontal scrolling.
- Put wide content such as tables in a container that scrolls horizontally
  instead of pushing the page sideways.
- Tap targets must be thumb-sized, and controls must not depend on hover.
- Check each new page at a narrow viewport before calling it done.

## Tests

- All tests are hermetic: `httptest` servers plus fixtures in
  `internal/yahoo/testdata/`, no live Yahoo calls. Pin new response shapes as
  fixtures instead of hitting the API.
- Fixtures are opened as `testdata/...`, which resolves because `go test` runs in the
  package directory. Keep that convention.
- Handlers take dependencies as small interfaces (`home.LoginState`,
  `players.Source`), so tests pass stubs instead of network clients.
- `config.Load` tests use `t.Setenv`; the repo `.env` is not loaded during tests
  (package cwd), so do not assume it is.
- `.env` and `.tokens.json` are gitignored secrets. Never commit them.
