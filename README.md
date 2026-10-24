# gocraft

A clean, terminal-native generator for production-grade Go web applications and JSON APIs — no frameworks, no reflection magic, no lock-in.

`gocraft` scaffolds runnable Go codebases in the standard architecture popularized by Alex Edwards (*Let's Go* and *Let's Go Further*): explicit dependency injection via an `application` struct, standard library `net/http` conventions, structured `log/slog` logging, graceful shutdown, and a self-documenting Makefile. Every emitted line is standard Go, formatted with `go/format`, vetted, and 100% yours to keep under the MIT license.

Built for both **humans** (interactive Bubble Tea TUI with live dependency inspection) and **AI coding agents** (deterministic headless CLI runner with `--dry-run` and `--json`).

---

## Installation

### macOS & Linux (Shell Script)

```bash
curl -fsSL https://raw.githubusercontent.com/VJ-2303/gocraft/main/scripts/install.sh | sh
```

### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/VJ-2303/gocraft/main/scripts/install.ps1 | iex
```

### Go Toolchain

Requires Go 1.27 or later:

```bash
go install github.com/VJ-2303/gocraft/cmd/gocraft@latest
```

### Manual Download

Download pre-compiled binaries and SHA-256 checksums for Linux, macOS, and Windows from [GitHub Releases](https://github.com/VJ-2303/gocraft/releases).

---

## Usage

`gocraft` provides dual interfaces with automatic environment detection:

### 1. Interactive TUI (Human Mode)

Run `gocraft` in any interactive terminal without flags:

```bash
gocraft
```

- **In-the-Moment Conflict Resolution**: Prerequisites (e.g. `Full Accounts -> SQLite + Sessions + Emails`) are displayed with live badges and auto-activated on selection.
- **Destructive Downgrade Protection**: Switching a setting (like changing database to `None` with active migrations) triggers an inline confirmation prompt before resetting dependencies.
- **Direct Review Jumps**: In the final review screen, press `[1-9]` to jump straight to any section; pressing Enter returns directly back to review.
- **Warm Editorial Aesthetic**: Split-pane layout (terminal width ≥ 100 cols) with an active architecture blueprint and dependency inspector on the right.

### 2. Headless CLI (Agent & Script Mode)

Pass flags or pipe input from automated scripts, CI pipelines, or AI agents:

```bash
# Scaffold a full-featured web app in one command
gocraft \
  --app=web \
  --module=github.com/myorg/myapp \
  --db=sqlite \
  --router=chi \
  --config=flags \
  --log=plaintext \
  --migrations \
  --sessions \
  --accounts=full \
  --target=./myapp

# Dry-run inspection without disk writes
gocraft --app=api --db=postgres --dry-run

# Structured JSON output for agents
gocraft --app=api --db=sqlite --target=./api --json
```

#### Available Flags

| Flag | Default | Description |
|---|---|---|
| `--app` | `web` | Application type: `web` (HTML templates) or `api` (JSON) |
| `--module` | `example.com/myapp` | Go module path |
| `--db` | `none` | Database dialect: `none`, `sqlite`, `postgres`, `mysql` |
| `--router` | `servemux` | HTTP router: `servemux`, `flow`, `chi`, `gorilla`, `httprouter` |
| `--config` | `flags` | Config parser: `flags` (CLI flags) or `env` (environment variables) |
| `--log` | `plaintext` | Log output format: `plaintext`, `json`, `colorized` |
| `--migrations` | `false` | Embedded SQL migrations and Makefile management targets |
| `--sessions` | `false` | Server-side sessions via `alexedwards/scs` (web only) |
| `--accounts` | `none` | User accounts: `none`, `full` (web), or `jwt` (api) |
| `--request-logging` | `false` | HTTP request logging middleware (IP, method, URI, latency) |
| `--basic-auth` | `false` | HTTP Basic Auth middleware with bcrypt-hashed passwords |
| `--git-integration` | `true` | Production `.gitignore` and Git version metadata package |
| `--secure-cookies` | `false` | HMAC-SHA256 signing and AES-256-GCM cookie encryption |
| `--auto-https` | `false` | Automatic Let's Encrypt TLS certificates via `autocert` |
| `--custom-error-pages` | `false` | Custom 400, 401, 404, 500 error pages (web only) |
| `--emails` | `false` | SMTP mailer helpers and multipart email templates |
| `--error-notifications` | `false` | Email runtime panics and 500 errors to admin address |
| `--tests` | `true` | Comprehensive handler and route test suites |
| `--target` | `""` | Destination path (must be empty or not exist) |
| `--dry-run` | `false` | Compute planned file tree without writing to disk |
| `--json` | `false` | Output machine-readable JSON summary |
| `--quiet` | `false` | Suppress progress messages |
| `--auto-deps` | `true` | Automatically enable required prerequisites |

---

## Feature Matrix

Every combination of options produces valid, compiling, and fully vetted Go code:

| Component | Options & Implementations |
|---|---|
| **Architectures** | **Web Application** (server-side HTML, partials, static assets) · **JSON API** (structured JSON envelopes, error payloads) |
| **Databases** | **None** (stateless) · **SQLite** (`modernc.org/sqlite`, pure Go, CGO-free) · **PostgreSQL** (`jackc/pgx/v5` connection pool) · **MySQL** (`go-sql-driver/mysql` with native timestamps) |
| **Routers** | Standard Library `http.ServeMux` (Go 1.22+ method matching) · `alexedwards/flow` · `go-chi/chi/v5` · `gorilla/mux` · `julienschmidt/httprouter` |
| **Authentication** | **Full Web Auth** (signup, login, logout, password reset, `justinas/nosurf` CSRF protection) · **JWT Token Auth** (registration, activation emails, bearer tokens) · **HTTP Basic Auth** (bcrypt hash) |
| **Storage & Security** | Database-backed sessions (`alexedwards/scs`) · Signed and encrypted cookies (HMAC + AES-GCM) · HTTPS autocert with port 80 redirect |
| **Observability** | Structured logging via `log/slog` (Plaintext, JSON, or Colorized dev mode) · Detailed request metrics middleware · Automatic email error notifications |
| **Testing** | Complete handler test suites, in-memory session mocks, SQLite temp DB harness with automatic teardown |

---

## Layout of Generated Codebases

Generated projects follow the standard Go project layout:

```
├── assets/                  Embedded assets: HTML templates, static files, SQL migrations, emails
├── cmd/web|api/             Application entrypoint:
│   ├── main.go              Startup, flags/env config, logger, pool lifecycle, DI wiring
│   ├── server.go            HTTP server with graceful shutdown and timeouts
│   ├── routes.go            Route declarations and middleware stack
│   ├── handlers.go          HTTP handler methods
│   ├── helpers.go           Template rendering or JSON encoding helpers
│   ├── middleware.go        Panic recovery, request logging, security headers, auth guards
│   ├── errors.go            Centralized HTTP error response helpers
│   └── *_test.go            Standardized handler test suites
├── internal/                Private application packages (emitted only when selected):
│   ├── database/            sqlx connection pool and query models (users, tokens, resets)
│   ├── env/                 Type-safe environment variable parsing with validation
│   ├── mailer/              SMTP mail client with embedded template rendering
│   ├── auth/                JWT tokenizer and bearer token extraction
│   ├── cookies/             Cryptographically signed and encrypted cookie helpers
│   ├── validator/           Input validation and field error collectors
│   ├── response/            JSON and HTML response writers
│   └── version/             Build and Git revision metadata helpers
├── Makefile                 Self-documenting build, test, audit, and migration tasks
├── go.mod                   Module path with pinned, minimal dependencies
└── README.md                Tailored getting-started guide for the generated codebase
```

---

## Generated Makefile Commands

Every generated codebase includes a zero-dependency, self-documenting `Makefile`:

```bash
make help          # View all available targets and descriptions
make run           # Build and run the application
make run/live      # Run with live-reloading via air
make test          # Run test suite with race detector
make test/cover    # Run tests and generate HTML coverage report
make audit         # Run tests, go mod verify, gofmt, go vet, staticcheck, govulncheck
make tidy          # Run go mod tidy, go fix, and gofmt
make migrations/up # Apply all database migrations (when migrations enabled)
```

---

## Architecture of `gocraft`

`gocraft` is built with a strictly decoupled 4-layer architecture:

```
cmd/gocraft/main.go
       │
       ├──► Interactive TTY? ──► internal/ui (Bubble Tea TUI)
       └──► Flags or non-TTY? ─► internal/cli (POSIX Headless Runner)
                                        │
                                        ▼
                                  internal/spec
                        (ProjectSpec ──► Validate ──► Resolve)
                                        │
                                        ▼
                                internal/generator
                        (File Matrix ──► text/template ──► go/format)
```

- **`internal/spec`**: Source of truth for validation rules, dependency cascades, dialect drivers, and view-model derivation.
- **`internal/generator`**: File matrix mapping specs to templates. Automatically formats all Go output through `go/format` (malformed templates fail generation instantly).
- **`internal/ui`**: Bubble Tea interactive state machine with split-pane layout, real-time cascade observer, and destructive conflict prevention.
- **`internal/cli`**: Headless runner designed for AI agents, CI scripts, and automation with `--dry-run` and `--json`.

---

## Development & Verification

To verify the generator and all templates:

```bash
make help          # View repository development commands
make audit         # Audit gocraft tool: tests, vet, staticcheck, govulncheck
make gate          # Generate 13 representative codebases and verify: build, vet, test
make matrix        # Generate 29 permutations and run full audit:
                   # build + vet + test + gofmt + stray-file + comment-hygiene
```

Before committing any changes to templates, run `make gate` (fast verification) or `make matrix` (exhaustive 29-codebase permutation audit).

See [AGENTS.md](AGENTS.md) for architectural invariants and contributing guidelines.

### Publishing Releases

Releases are automated via GitHub Actions and GoReleaser. To publish a new release across all platforms:

```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

The release pipeline runs all verification gates, cross-compiles binaries for Linux, macOS, and Windows (`amd64` and `arm64`), generates checksums and changelogs, and publishes the GitHub release.

---

## License

[MIT](LICENSE) — free for personal, commercial, and open-source use, for both this tool and every codebase it generates.
