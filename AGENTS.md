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
go run ./cmd/server                       # serve on :8080 (hardcoded)
air                                       # hot reload via .air.toml -> ./tmp/main
go test ./...                             # all tests, no network
go test ./internal/infra/yahoo -run TestLeague -v
go vet ./...
```

There is no Makefile, CI, linter config, or codegen. `tmp/` is build output.

## Layout

```
cmd/server/main.go           entrypoint: log.Fatal(server.Run())
internal/server              composition root: wires deps, registers routes, :8080
internal/pages/home          home page handler + templates
internal/pages/players       free-agent page handler + templates + sort/heatmap view
internal/web                 shared layout + Parse/Render + RequireSession middleware
internal/domain/player       Player, Stats, MissingStat (pure domain types)
internal/service/freeagents  Source + YahooSource/MockSource (yahoo -> domain stats)
internal/service/session     Yahoo token verification and status propagation
internal/infra/auth          Yahoo OAuth client + /login and /callback handlers
internal/infra/yahoo         read-only Fantasy Sports API client (XML)
internal/infra/config        package name is "config" (dir is "infra/config")
docs/yahoo-api-shapes.md     field-level reference for league/player resources
```

- `internal/infra/config` is **`package config`**; imports must alias it:
  `config "fantasy_waiver_assistant/internal/infra/config"`.
- Import paths use the module name `fantasy_waiver_assistant` (underscore).
- Layers point inward: `pages`/`web` (transport) -> `service`/`domain` ->
  `infra`. `domain/player` imports nothing internal; `infra` never imports pages.

## Routing and htmx

- Pages use clean paths (`/`, `/players`). One route serves both shapes: the full
  layout on a direct visit, the named fragment when `HX-Request: true`. There are
  no separate `/hx/...` component routes.
- Routes registered on the `app` mux in `internal/server` are gated by
  `web.RequireSession`; `/login` and `/callback` stay public on the outer mux.
- `web.Render` writes the layout for direct visits and the named fragment when
  `HX-Request: true`, and sets `Vary: HX-Request`, so a cached fragment is never
  served as a full page.
- Feature templates define `content` plus their fragments and are parsed once at
  startup with `web.Parse(featureFS, patterns...)`, which merges the shared layout.
- Filter links set `hx-push-url` to the page URL (`/players?sort=PTS&dir=desc`),
  never to a fragment URL, so reloads and the back button land on a full page.
- The layout marks the active tab from a `Nav` value in the page data (`"home"`,
  `"players"`). Set it in every page handler; fragments ignore it.

## Hard-won constraints

- **Yahoo API is XML-only by design.** `internal/infra/yahoo` sends
  `Accept: application/xml` and decodes `fantasy_content` envelopes. JSON quirks are
  documented in `docs/yahoo-api-shapes.md` for reference only; do not add a JSON
  decoder.
- **Read `docs/yahoo-api-shapes.md` before adding fields/types.** It records wire
  names, nesting, and quirks (string-valued stats, `-` for DNP, 25 players/page cap,
  lowercase `l` in `470.l.123` keys).
- Build keys and paths with the helpers in `internal/infra/yahoo/path.go`. Position values
  contain `/`, so `escapeSegment` deliberately escapes only `% ? # space` — do not
  swap in full `url.PathEscape`.
- When Yahoo errors arrive, `client.go` returns `yahoo.ErrUnauthorized` on HTTP 401
  and unwraps `<error>` documents (including HTTP 200 error bodies) into
  `*yahoo.APIError`. Nothing refreshes and retries yet; that belongs in the caller.
- `auth.Client` (in `internal/infra/auth`) keeps its token **in memory only**.
  `/callback` writes `.tokens.json` (0600) but startup never loads it, so a restart
  means logged out. `internal/server` adapts it to `yahoo.Client`'s `TokenSource`;
  reuse that adapter for new authenticated Yahoo calls instead of adding a second one.
- `internal/service/session` verifies a token against `/users;use_login=1`; token
  existence alone is not proof of a working login. Verified status is passed
  through the request context so middleware and handlers do not validate twice.
  A Yahoo 403 is `AccessDenied`, not a logout: it means the Yahoo app lacks
  Fantasy Sports permission and the home page shows how to fix it.
- Every feature embeds its templates with `//go:embed` and parses them once at
  startup via `web.Parse`, so the binary does not depend on the working directory.
  Do not reintroduce repo-root-relative `template.ParseFiles(...)`.
- Config: `config.Load()` reads `.env` from the cwd, never overwrites variables
  already set (real env wins), and reports all missing required vars in one error.
  Required: `YAHOO_CLIENT_ID`, `YAHOO_CLIENT_SECRET`, `YAHOO_REDIRECT_URI`.
  Optional: `YAHOO_LEAGUE_KEY` (e.g. `470.l.123`), used by the players page.
  `PLAYERS_SOURCE=mock` serves built-in sample players from a public `/players`;
  the default `yahoo` keeps players behind a verified session.
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
  `internal/infra/yahoo/testdata/`, no live Yahoo calls. Pin new response shapes as
  fixtures instead of hitting the API.
- Fixtures are opened as `testdata/...`, which resolves because `go test` runs in the
  package directory. Keep that convention.
- Handlers take dependencies as small interfaces (`home.SessionValidator`,
  `freeagents.Source`), so tests pass stubs instead of network clients.
- `config.Load` tests use `t.Setenv`; the repo `.env` is not loaded during tests
  (package cwd), so do not assume it is.
- `.env` and `.tokens.json` are gitignored secrets. Never commit them.
