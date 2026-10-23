# Plan: `nomoreboiler` — Autostrada feature-parity Go project generator (TUI)

Build a Bubble Tea + Huh TUI that generates complete, runnable Go web-app / JSON-API codebases in Alex Edwards' *Let's Go* style, replicating **every** Autostrada option (free + paid). Templates are written from scratch by us following public conventions — features are replicated, not Autostrada's proprietary template code.

Your decisions: Bubble Tea + Huh · TUI-only (no headless mode) · output to a directory · current Autostrada feature set only.

---

## 1. Git workflow (used throughout development)

- **M0:** `git init` on `main`; commit `.gitignore`, `LICENSE` (MIT), stub `README.md`, `Makefile` → `chore: initialize repository`.
- **One branch per milestone:** `feat/02-spec-model`, `feat/04-vertical-slice`, `feat/06-databases`, …
- **Conventional commits** at every coherent unit (`feat(spec): enforce dependency rules`, `test(generator): golden files for sqlite web app`, `fix(templates): chi wildcard syntax`). Commit when a unit compiles + tests pass; never commit red.
- **Merge `--no-ff` into `main` only after the milestone's verification gate passes.** Main stays always-buildable.
- **Tags:** `v0.1.0` after vertical slice, `v0.2.0` after all Step-2 options, `v1.0.0` at full parity.
- Local repo only; a remote can be added later on request.

## 2. Tool repository layout (every file)

```
nomoreboiler/
├── cmd/nomoreboiler/main.go          # run() pattern; parses --version; starts UI; returns error
├── internal/
│   ├── spec/
│   │   ├── spec.go                   # ProjectSpec struct (wizard answers)
│   │   ├── validate.go               # Validate() error: module-path regexp, per-app-type legality, dependency rules
│   │   ├── resolve.go                # Resolve() ResolvedSpec — template view-model w/ derived fields
│   │   └── spec_test.go              # table-driven tests incl. every illegal combination
│   ├── ui/
│   │   ├── model.go                  # Bubble Tea model; states: Wizard → Review → PathInput → Generating → Done/Error
│   │   ├── form.go                   # 3 Huh forms; conditional fields per dependency rules
│   │   └── render.go                 # Lipgloss styles; review summary; done screen w/ tailored run instructions
│   ├── generator/
│   │   ├── generator.go              # Generate(rsp ResolvedSpec, target string) error
│   │   ├── files.go                  # file-selection matrix: []FileSpec{Template, OutPath, Cond}
│   │   ├── viewmodel.go              # helpers: random port, random session-cookie name, go.mod deps per options
│   │   └── generator_test.go         # golden-file tests over representative specs
│   └── logging/logging.go            # slog setup for the tool itself
├── templates/  (go:embed)
│   ├── common/{makefile,readme,gitignore,gomod}/
│   ├── shared/internal/{env,version,database,request,response,validator,password,mailer,funcs}/
│   ├── web/{cmd,assets}/…
│   └── api/{cmd,assets}/…
├── testdata/matrix/                  # spec fixtures driving `make matrix`
├── Makefile                          # help/audit/test/test/cover/upgradeable/tidy/build/run + `matrix`
├── go.mod  (bubbletea, huh, lipgloss)  └── README.md
```

### ResolvedSpec view-model (drives all templates)
Direct fields (spec answers) plus derived: `CmdDir` (`web`/`api`), `UseDB`, `Dialect` (`sqlite|postgres|mysql|none`), `DriverImport`/`DriverDSNComment`, `NeedsDB` (DB or sessions or accounts), `NeedsMailer` (emails, error-notifications, either accounts variant), `NeedsSessions`, `HasAccountsFull`, `HasAccountsJWT`, `NeedsAuthMiddleware`, `RouterKey`, `ConfigSource`, `LogFormat`, `HasMigrations`, `HasTests`, `Port` (random per codebase), `SessionCookieName` (random 8-char), `GoModRequires` (per-option dependency list), `TitleName`, `AssetsEmbedList`.

### Dependency rules (enforced in `Validate()` + wizard UI)
- sessions → web app **and** database
- accounts-full → web app, database, sessions, emails (CSRF/nosurf auto-included)
- accounts-JWT → API app, database, emails (activation mail auto-included)
- error-notifications → emails
- custom error pages → web only; accounts variants restricted by app type
- migrations recommended with sessions/accounts (wizard auto-ticks & locks)

## 3. Wizard UX flow

1. **Step 1** (Huh select): Traditional web app · JSON API.
2. **Step 2**: database, migrations (toggle), router, config source, log format, request logging, basic auth, git integration, secure cookies, sessions (only shown for web+DB), module path (Huh text input, regexp-validated live: domain/path form, no scheme).
3. **Step 3** (Plus): auto HTTPS, custom error pages (web), emails, error notifications (shown only if emails), tests, accounts variant (per app type). Invalid combos auto-disabled with inline explanation.
4. **Review screen**: Lipgloss table of every choice + dependency consequences; edit / confirm.
5. **Path input**: absolute or relative dir; must not exist or be empty; created with parents.
6. **Generating**: spinner w/ file count, then **Done screen**: tree summary + tailored run instructions (`cd`, `make run`, migration commands if applicable) — mirroring Autostrada's confirmation page. `ctrl+c` aborts cleanly at any state.

## 4. Generated-code template inventory

### Shared `internal/` packages (emit when used)
- `env/` — `GetString/GetInt/GetBool` with fallbacks (panics on bad values).
- `version/` — `Get()` from VCS build info, `GetRevision()` commit hash.
- `database/database.go` — sqlx `New(ctx, dsn)` w/ connect timeout, pool settings; `MigrateUp()` runs embedded `assets/migrations/*.up.sql` in order (idempotent).
- `database/users.go`, `password_resets.go`, `tokens.go` — per accounts variant; single-row SELECTs return `(T, bool)`.
- `request/` — `DecodePostForm` (1MB `MaxBytesReader`), `DecodeQueryString`, strict `DecodeJSONObject/DecodeJSONArray` using `encoding/json/v2`, returning `response.BadRequestError`.
- `response/` — `WriteJSON` (json/v2), `HTMLRenderer` (parse-once at startup w/ `funcs` + `{{template}}` composition), `BadRequestError`.
- `validator/` — `Validator` w/ `CheckField/AddFieldError`, stdlib `cmp` helpers.
- `password/` — bcrypt `Hash()` (cost 12) / `Matches()`; `common.go` common-password list.
- `mailer/` — `wneessen/go-mail` SMTP client, `Send(to, data, template)` rendering subject/plain/optional-html blocks.
- `funcs/` — template func map (`year`, `safeHTML`, etc.).

### Web app (`cmd/web/` + `assets/`)
- `main.go` — config struct (flags **or** env variant), slog handler (plaintext/JSON/colorized variant), DB/session/mailer/HTTPS init variants, `application` struct, `run()`, graceful shutdown (SIGINT/SIGTERM, `srv.Shutdown`, standardized logs).
- `routes.go` — **five router variants** (ServeMux `method pattern`, Flow, Chi, Gorilla, HttpRouter) × middleware chain (logRequest → recoverPanic → secureHeaders → sessionManager.LoadAndSave → nosurf → authenticate; basicAuth where ticked).
- `handlers.go` — home; DB example handler when DB selected.
- `handlers-auth.go` — signup/login/logout, forgotten-password (+confirmation), password-reset/{token} (+confirmation); redirect-after-login.
- `helpers.go` — `render`, `serverError`/`clientError` (variants: custom error pages, email notifications), `backgroundTask` (recovers panics, `sync.WaitGroup`), `defaultHTMLData` (`BaseURL`, `CSRFToken`, `AuthenticatedUser`), `defaultEmailData`.
- `middleware.go` — logRequest (status/body-size/latency), recoverPanic, secureHeaders, basicAuth, `authenticate`, `requireAuthenticatedUser`, `requireAnonymousUser`, `preventCSRF` (nosurf).
- `templates.go` (HTMLRenderer), `context.go` (`contextGetAuthenticatedUser` → `(User, bool)`).
- `*_test.go` + `testsupport` (Tests option): `newTestApplication`, httptest helpers, per-handler coverage.
- `assets/`: `efs.go` (embed); `static/css,js`; `html/base.tmpl`, `pages/{home,signup,login,forgotten-password,forgotten-password-confirmation,password-reset,password-reset-confirmation, error pages}.tmpl`, `components/nav.tmpl` (logout form when accounts); `emails/`; `migrations/000001_init` + `000002_create_sessions_table` + `000003_create_users_table` + `000004_create_password_resets_table` (up/down pairs; SQLite FK pragma, case-insensitive unique email, users↔password_resets FK).

### JSON API (`cmd/api/` + `assets/`)
- `main.go`/`routes.go` (same variants)/`handlers.go` (`GET /v1/healthcheck` returning version+status+db check).
- Accounts-JWT: `POST /v1/users` (register+email activation token), `PUT /v1/tokens/activation`, `POST /v1/tokens/authentication` (issue JWT access+refresh), refresh endpoint, password-reset endpoints; `requireAuthenticatedUser` middleware validating `Authorization: Bearer` JWT; users + activation_tokens tables.
- Tests option: API test suite (`httptest` server, auth'd request helpers).

### Root files per codebase
`Makefile` (help, audit, test, test/cover, upgradeable, tidy, build, run, run/live w/ embedded `.air.toml`, `migrations/new|up|down|goto|force|version`), `README.md` (tailored run instructions), `go.mod` (chosen module path), `.gitignore` (git integration), all files mode 0644.

### Config settings emitted (flag / env var pairs)
base: `--base-url`/`BASE_URL`, `--http-port`/`HTTP_PORT` (random default) · db: `--db-dsn`/`DB_DSN`, `--db-max-conns`, `--db-min-conns`, `--db-max-idle-time`, `--db-automigrate`/`DB_AUTOMIGRATE` · sessions: `--session-cookie-name`/`SESSION_COOKIE_NAME` · smtp: `--smtp-{host,port,username,password, from}`/`SMTP_*` · error notifications: `--admin-email`/`ADMIN_EMAIL` (sensible defaults we choose where Autostrada doesn't document them).

## 5. Milestones, gates, commits

| # | Branch | Work | Verification gate (then merge `--no-ff`) |
|---|---|---|---|
| 0 | — | git init, .gitignore, LICENSE, Makefile stub, README | initial commit |
| 1 | `feat/tool-skeleton` | go.mod, `cmd/nomoreboiler` run(), slog, Makefile targets | `make audit` green |
| 2 | `feat/spec-model` | spec/validate/resolve + full table tests | `go test ./...` green |
| 3 | `feat/generator-engine` | embed FS, file matrix, writer (0644), golden harness, common templates | golden tests green |
| 4 | `feat/vertical-slice` | minimal web+api base templates, working TUI end-to-end | generated apps `go build` + `go run` + curl 200 → **tag v0.1.0** |
| 5 | `feat/tui-wizard` | full 3-step forms, conditionals, review, path input, done screen | manual matrix of UI states |
| 6 | `feat/databases` | 4 dialects + migrations + pool config + example handler | each dialect builds; sqlite+postgres run & migrate |
| 7 | `feat/routers` | 5 routes.go variants × chain composition | build each router×app-type |
| 8 | `feat/config-logging-auth` | flags/env ×2, log formats ×3, request logging, basic auth | build matrix green |
| 9 | `feat/git-cookies-sessions` | git integration, secure cookies, scs sessions (+cookie name, migrations interplay) | sessions flow smoke-tested → **tag v0.2.0** |
| 10 | `feat/plus-platform` | auto HTTPS (autocert + redirect), custom error pages, emails, error notifications | HTTPS config unit-tested; emails via Mailpit-style local SMTP |
| 11 | `feat/accounts-web` | accounts-full: schema, handlers, middleware, CSRF, templates | full signup→reset flow tested |
| 12 | `feat/accounts-api` | accounts-JWT: schema, endpoints, middleware, activation emails | auth flow integration test |
| 13 | `feat/tests-option` | web+api generated test suites + helpers | generated `make test` green in all variants |
| 14 | `feat/matrix-polish` | `make matrix` (curated ~36 permutations: build+vet each), README, help text | matrix green → **tag v1.0.0** |

## 6. Verification & standards

- **Compile matrix:** `testdata/matrix/*.json` specs → generate → `go build ./... && go vet ./...` in each; kept green on main.
- **Golden tests** pin rendered output; updated deliberately, never blindly.
- Tool's own code follows the same Edwards conventions (`run()` entry, `internal/`, DI via structs, `errors.Is/As`, no panics in library code, slog logging, table-driven tests, `make audit` = test + vet + gofmt + staticcheck + govulncheck).
- Day 1: verify local Go toolchain; generated-code baseline (json/v2, `rand.Text`, `url.Clone`, `sync.WaitGroup.Go`) matches installed version, with fallbacks if older.

## Risks
- **Combination explosion** (2×4×5×2×3×~10 booleans): mitigated by view-model conditionals, per-variant template files, and the curated matrix.
- **Router/DB syntax drift**: pinned dependency versions in generated `go.mod`; upgradeable target mirrors Autostrada.
