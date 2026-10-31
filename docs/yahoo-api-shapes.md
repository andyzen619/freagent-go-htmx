# Yahoo Fantasy API shapes: League & Players

Field-level shape reference for the two resources the waiver assistant needs first.
Compiled 2026-10-31 from:

- Yahoo Sports Developer Portal docs capture (2026-05-14, via `spiflicate/yfs-api` archive)
- Live unauthenticated probes of Yahoo's public read host
  `https://pub-api-ro.fantasysports.yahoo.com/fantasy/v2/...?format=json` (2026-10-31, NFL game 470)
- Cross-checked against `famendola1/yfantasy` (Go XML schema) and `n-ae/yahoo-fantasy-sports-api-go`
  (Go JSON schema) field definitions

Field names below are the **XML wire names** (snake_case). The JSON wire format uses the
same names but has structural quirks — see [JSON wire format](#json-wire-format-quirks).

## Endpoint conventions

| | |
|---|---|
| Base URL (OAuth) | `https://fantasysports.yahooapis.com/fantasy/v2` |
| Format | XML by default; `?format=json` for JSON (same data, quirks below) |
| League key | `{game_key}.l.{league_id}` — e.g. `461.l.1000` (lowercase `l`, not `1`) |
| Player key | `{game_key}.p.{player_id}` — e.g. `461.p.30121` |
| Team key | `{game_key}.l.{league_id}.t.{team_id}` |
| Game key | `{game_id}` or game code (`nfl`); codes resolve to the current season id |

Requests are read-only (`GET`). Sub-resources are path segments or `;out=a,b` params:

```
/fantasy/v2/league/{league_key}                       # metadata (default)
/fantasy/v2/league/{league_key}/settings              # or /standings /scoreboard /teams /players ...
/fantasy/v2/league/{league_key};out=settings,teams    # several at once
/fantasy/v2/league/{league_key}/players;player_keys={key1},{key2}/ownership
```

---

## League resource

`GET /fantasy/v2/league/{league_key}` — default sub-resource: `metadata`.

### Sub-resources

| Name | Description | URI |
|---|---|---|
| metadata | key, id, name, url, draft status, team count, current week | `/league/{key}` |
| settings | draft type, scoring, roster positions, stat categories, waivers | `/league/{key}/settings` |
| standings | team rankings (includes `team_standings` under teams) | `/league/{key}/standings` |
| scoreboard | matchups (optional `;week={n}`), includes `team_stats` | `/league/{key}/scoreboard` |
| teams | all teams in the league | `/league/{key}/teams` |
| players | the league's eligible players | `/league/{key}/players` |
| draftresults | draft results for all teams | `/league/{key}/draftresults` |
| transactions | adds, drops, trades | `/league/{key}/transactions` |

Leagues collection: `/fantasy/v2/leagues;league_keys={k1},{k2}` — same shape per league.

### League metadata fields

| Field | Type | Notes |
|---|---|---|
| `league_key` | string | `461.l.1000` |
| `league_id` | number | |
| `name` | string | |
| `url` | string | |
| `logo_url` | string | optional |
| `draft_status` | string | `predraft` / `drafting` / `postdraft` |
| `num_teams` | number | |
| `edit_key` | string | changes on league edit; looks numeric/datey but keep string |
| `weekly_deadline` | string | e.g. `intraday`, often empty |
| `league_update_timestamp` | number | unix seconds |
| `scoring_type` | string | `head` / `roto` / `points` |
| `league_type` | string | `public` / `private` |
| `renew`, `renewed` | string | league migration keys, often empty |
| `iris_group_chat_id` | string | often empty |
| `allow_add_to_dl_extra_pos` | bool 0/1 | |
| `is_pro_league`, `is_cash_league` | bool 0/1 | |
| `current_week` | number | |
| `start_week`, `end_week` | number | |
| `start_date`, `end_date` | string | `YYYY-MM-DD` |
| `is_finished` | bool 0/1 | |
| `game_code` | string | `nfl` etc. |
| `season` | number | e.g. `2026` |

Also seen in real responses (include if useful, all optional):
`roster_type` (`date`/`week`), `current_date` (`YYYY-MM-DD`), `matchup_week`,
`felo_tier`, `is_plus_league`, `is_highscore`, `league_chat_id`, `short_invitation_url`.

### Settings shape (`/league/{key}/settings`)

Scalar settings fields:

| Field | Type | Notes |
|---|---|---|
| `draft_type` | string | `live`/`offline`/`autodraft`... |
| `is_auction_draft` | bool | |
| `scoring_type` | string | |
| `uses_playoff` | bool | |
| `has_playoff_consolation_games` | bool | |
| `playoff_start_week` | number | |
| `uses_playoff_reseeding`, `uses_lock_eliminated_teams` | bool | |
| `num_playoff_teams`, `num_playoff_consolation_teams` | number | |
| `has_multiweek_championship` | bool | |
| `waiver_type` | string | **waiver behavior** — needed for waiver logic |
| `waiver_rule` | string | e.g. `all`, `gametime`... |
| `uses_faab` | bool | FAAB budget instead of waiver priority |
| `waiver_time` | string | when waivers clear |
| `draft_time`, `draft_pick_time` | string | |
| `post_draft_players` | string | `waivers` / `freeagents` |
| `max_teams` | number | |
| `trade_end_date` | string | `YYYY-MM-DD` |
| `trade_ratify_type` | string | `yahoo`/`commish`... |
| `trade_reject_time` | number | |
| `player_pool` | string | e.g. `all` |
| `cant_cut_list` | string | |
| `sendbird_channel_url` | string | |
| `persistent_url`, `uses_roster_import`, `roster_import_deadline`, `max_weekly_adds`, `can_trade_draft_picks`, `draft_together` | mixed | optional, sport/league variants |

Nested:

```
settings
├── roster_positions            (list)
│   └── roster_position         { position, position_type, count }
│                               e.g. QB/O/1, W/R/T/O/1; count = roster slots
├── stat_categories
│   └── stats                   (list)
│       └── stat                { stat_id, enabled, name, display_name,
│                                 sort_order, position_type?,
│                                 stat_position_types: [stat_position_type:{position_type}] }
├── stat_modifiers
│   └── stats                   (list)  { stat_id, value }   # scoring weights
└── divisions                   (list)  { division_id, name }
```

`stat_id` values map player stats to categories — e.g. NFL: 4 = pass yds, 5 = pass tds,
9 = rush yds, 10 = rush tds, 12 = rec yds, 13 = rec tds (full table via
`/game/{game_key}/stat_categories`).

---

## Players

### URIs (players collection)

| URI | Meaning |
|---|---|
| `/league/{league_key}/players` | all eligible players in a league |
| `/leagues;league_keys={k1},{k2}/players` | players across leagues |
| `/team/{team_key}/players` | players on a team |
| `/teams;team_keys={k1},{k2}/players` | players across teams |
| `/players;player_keys={p1},{p2}` | specific players |
| `/game/{game_key}/players;player_keys={...}` | game context (no league info) |

Any player sub-resource works under the collection: `/players/{sub}` or `/players;out=a,b`.

### Player sub-resources

| Name | Description | URI |
|---|---|---|
| metadata | key, id, name, editorial info, image, eligible positions | `/player/{key}` (default) |
| stats | season/week/date stats, and points in league context | `/player/{key}/stats` — `;type=week;week={n}` or `;type=date;date={yyyy-mm-dd}` |
| ownership | owned / waivers / free agent **within a league** | `/league/{league_key}/players;player_keys={key}/ownership` |
| percent_owned | % of Yahoo leagues owning the player | `/player/{key}/percent_owned` |
| draft_analysis | avg pick / round / cost, % drafted | `/player/{key}/draft_analysis` |

For a waiver assistant: fetch via the league collection with
`;status=FA` / `;status=W` and `;out=ownership,percent_owned`.

### Players collection filters (league context)

| Filter | Values | Notes |
|---|---|---|
| `position` | valid position (`QB`, `RB`, `W/R/T`, ...) | league context only |
| `status` | `A` all available, `FA` free agents, `W` waivers, `T` taken, `K` keepers | league context only |
| `search` | player name | league context only |
| `sort` | `{stat_id}`, `NAME`, `OR` overall rank, `AR` actual rank, `PTS` | league context only |
| `sort_type` | `season`, `week` (football), `date`/`lastweek`/`lastmonth` (other sports) | |
| `sort_season` | year | with `sort_type=season` |
| `sort_week` | week | football |
| `sort_date` | `YYYY-MM-DD` | non-football |
| `start` | int ≥ 0 | pagination offset |
| `count` | int > 0 | **page size; community reports a 25/page cap** — paginate `start` |

Example waiver wire query:

```
GET /fantasy/v2/league/{league_key}/players;status=FA;position=WR;sort=PTS;sort_type=season;count=25;out=ownership,percent_owned
```

### Player metadata fields

| Field | Type | Notes |
|---|---|---|
| `player_key` | string | `461.p.30121` |
| `player_id` | number | |
| `name` | object | `full`, `first`, `last`, `ascii_first`, `ascii_last` |
| `url` | string | |
| `status` | string | injury code: `Q`, `O`, `IR`, `P`, ... (empty = healthy) |
| `status_full` | string | `Questionable`, ... |
| `injury_note` | string | e.g. `Hamstring` |
| `editorial_player_key` | string | `nfl.p.30121` (game-code form) |
| `editorial_team_key` | string | `nfl.t.25` (real NFL team) |
| `editorial_team_full_name` | string | `San Francisco 49ers` |
| `editorial_team_abbr` | string | `SF` |
| `editorial_team_url` | string | |
| `bye_weeks` | object | `week` (single) or list of weeks |
| `uniform_number` | number | |
| `display_position` | string | `QB`, `RB`, sometimes `1B,OF` |
| `headshot` | object | `url`, `size` (`small`) |
| `image_url` | string | legacy field |
| `is_undroppable` | bool 0/1 | |
| `position_type` | string | `O`, `P`, `K`, `B`, `DP`... |
| `primary_position` | string | |
| `eligible_positions` | object/list | `position` leaf(s): `RB`, `W/R/T`, ... |
| `eligible_positions_to_add` | | optional |
| `is_keeper` | object | `status`, `cost` (often empty) |
| `has_player_notes`, `has_recent_player_notes` | bool | |
| `player_notes_last_timestamp` | number | unix seconds |

### Stats & points (`;out=stats` or `/stats`)

```
player_stats            { coverage_type: season|week|date,
                          season|week|date,          # coverage key
                          stats: [stat:{stat_id, value}] }
player_points           { coverage_type, season|week|date, total }   # league context only
```

`value` is a **string** and can be `-` for DNP. In roster/matchup contexts there is also
`selected_position` = `{position, is_flex?}`.

### Ownership, percent_owned, draft_analysis

```
ownership               { ownership_type: team|waivers|freeagent,
                          owner_team_key,            # "" when unowned
                          owner_team_name,           # "" when unowned
                          team? }                    # full team resource sometimes attached

percent_owned           { coverage_type: season|week,
                          week, value, delta }       # value = % of leagues (0-100)

draft_analysis          { average_pick, average_round,
                          average_cost, percent_drafted }   # all strings, e.g. "30.8"
```

`ownership_type` is the field that answers "is this player addable?" —
`freeagent` = add now, `waivers` = claim, `team` = rostered (`owner_team_key` says where).

---

## JSON wire format quirks

`?format=json` output is an XML→JSON transliteration with structural traps.
Observed live on 2026-10-31 (`game/nfl/...` responses):

### 1. Envelope and chunk arrays

A resource is an **array of chunks**: index 0 = its metadata fields, later indices = one
sub-resource container each.

```json
{ "fantasy_content": {
    "xml:lang": "en-US",
    "yahoo:uri": "/fantasy/v2/game/nfl/game_weeks",
    "game": [
      { "game_key": "470", "game_id": "470", "name": "Football", "...": "metadata fields" },
      { "game_weeks": { "...": "sub-resource" } }
    ],
    "time": "8.5ms", "copyright": "...", "refresh_rate": "30"
} }
```

### 2. Collections: numeric-keyed maps **or** wrapper arrays

```json
"players": { "0": { "player": [...] }, "1": { "player": [...] }, "count": 2 }
"game_weeks": { "0": { "game_week": { "week": "1", "start": "2026-09-09" } }, "1": { ... } }
"roster_positions": [ { "roster_position": { "position": "QB", "...": "..." } }, { ... } ]
"stat_categories": { "stats": [ { "stat": { "stat_id": 0, "...": "..." } }, { ... } ] }
```

Same logical shape, two encodings: `players`/`game_weeks` are objects keyed `"0"`,`"1"`,…
(plus `count`), while `roster_positions`/`stats` are arrays of one-key wrappers.
A normalizer must accept both.

### 3. Player = positional one-key field array

Each `player` entry is `[ [ {"player_key": ...}, {"player_id": ...}, {"name": {...}}, ... ] ]`
— one single-key object per XML child element, order-preserving — while simpler elements
(`game_week`, `stat`, `roster_position`) come back as plain merged objects. **Do not assume
flat objects for players.**

### 4. Single vs. many and type chaos

- Repeated leaves collapse to a bare value when there is one:
  `"bye_weeks": {"week": "8"}`, `"eligible_positions": {"position": "RB"}` — but they become
  lists when there are several (`{"position": ["RB", "W/R/T"]}` or positional variants).
- Most numbers arrive as **strings** (`"player_id": "30121"`), but `count` and some
  newer fields arrive as JSON numbers. Empty values are `""`, not `null`.
- Error envelope: `{"error": {"description": "...", "detail": "..."}}` (also used for
  auth failures).

### Normalization recipe (if staying on JSON)

1. Unwrap one-key wrapper objects into their value (keep the key when the wrapper *is* the
   field, e.g. `{"name": {...}}`).
2. Merge positional single-key field arrays into one object per resource.
3. Convert numeric-keyed containers (`{"0":..., "1":..., "count":N}`) to ordered lists.
4. Wrap bare scalar-vs-list leaves as `[]T` uniformly.
5. Decode ids/keys/timestamps as strings or `json.Number`, parse lazily.

**Decision:** this project decodes **XML**. It is Yahoo's canonical format (the default
response and every official sample) and maps cleanly onto Go `encoding/xml` structs with
no normalization. The types and client live in `internal/yahoo/`; the JSON quirks above are
kept for reference only.

Once OAuth lands, capture real responses from
`league/{key}`, `league/{key}/settings`, `league/{key}/players;status=FA;out=ownership,percent_owned`
into `internal/yahoo/testdata/` and pin them with golden tests — the frontend host probed
above is not guaranteed identical to the OAuth API host.

---

## Go implementation

Implemented in `internal/yahoo/` (XML decoding, no third-party dependencies):

| File | Contents |
|---|---|
| `types.go` | `League`, `LeagueSettings`, `Player`, `Players`, `Ownership`, `PercentOwned`, `PlayerStats`, ... |
| `decode.go` | `Envelope` (`<fantasy_content>` root), `APIError`, `decodeResponse` |
| `path.go` | key builders, `PlayerFilter` (status/position/sort/pagination/out), path encoding |
| `client.go` | `Client`, `TokenSource`, resource methods, `ErrUnauthorized` |

```go
c := yahoo.NewClient(func(ctx context.Context) (string, error) {
	// wire this to auth.Client once tokens persist
	return token, nil
})

league, err := c.League(ctx, yahoo.LeagueKey("470", 1000), "settings")
players, err := c.LeaguePlayers(ctx, league.LeagueKey, yahoo.PlayerFilter{
	Status: yahoo.StatusFreeAgents,
	Sort:   "PTS",
	Count:  25, // Yahoo caps pages at 25
	Out:    []string{"ownership", "percent_owned"},
})
```

Golden fixtures in `internal/yahoo/testdata/` come from the official sample responses
(`league.xml`, `league_settings.xml`, `player_league.xml`, `player_stats.xml`) plus
synthetic players/ownership documents. On a 401 the client returns `yahoo.ErrUnauthorized`
so the caller can refresh the token and retry. The JSON decoder is deliberately not
implemented.
