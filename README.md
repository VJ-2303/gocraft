# gocraft

A clean, terminal-native generator for production-grade Go web applications and JSON APIs — no frameworks, no reflection magic, zero lock-in.

`gocraft` scaffolds runnable Go codebases in the standard architecture popularized by Alex Edwards (*Let's Go* and *Let's Go Further*): explicit dependency injection via an `application` struct, standard library `net/http` conventions, structured `log/slog` logging, graceful shutdown, and a self-documenting Makefile. Every emitted line is standard Go, formatted with `go/format`, vetted, and 100% yours to keep under the MIT license.

Built for both **humans** (interactive Bubble Tea TUI with live dependency resolution) and **scripts/agents** (deterministic headless CLI runner with `--dry-run` and `--json`).

---

## Features

- **Standard Library Core**: Zero heavy frameworks; built on standard library patterns, idiomatic middleware, and composable handlers.
- **Dual Architectures**: Choose between server-rendered HTML **Web Applications** (with partials and static assets) or **JSON APIs** (with structured envelopes and standard error responses).
- **Multiple SQL Dialects**: First-class support for **SQLite** (pure Go, CGO-free via `modernc.org/sqlite`), **PostgreSQL** (connection pooling via `jackc/pgx/v5`), and **MySQL** (`go-sql-driver/mysql`), or **None** for stateless services.
- **Router Flexibility**: Standard library `http.ServeMux` (Go 1.22+ routing), `alexedwards/flow`, `go-chi/chi/v5`, `gorilla/mux`, or `julienschmidt/httprouter`.
- **Production Authentication**:
  - Full web auth: user registration, activation, login/logout, password reset, and CSRF protection via `justinas/nosurf`.
  - API auth: user registration, activation emails, and signed JWT bearer tokens.
  - Basic auth: HTTP Basic Auth middleware backed by bcrypt-hashed credentials.
- **State & Storage**: Database-backed sessions (`alexedwards/scs`), cryptographically signed and encrypted cookies (HMAC-SHA256 + AES-GCM), and embedded SQL migrations.
- **Observability & Security**: Structured `log/slog` logging (plaintext, JSON, or colorized), request latency tracking, automatic admin email error alerts on runtime panics, and automatic Let's Encrypt TLS via `autocert`.
- **Comprehensive Test Suites**: Pre-configured test harnesses with in-memory SQLite instances and mock sessions.

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

### Go Toolchain (Go 1.27+)

```bash
go install github.com/VJ-2303/gocraft/cmd/gocraft@latest
```

### Manual Download

Download pre-compiled binaries and SHA-256 checksums for Linux, macOS, and Windows from [GitHub Releases](https://github.com/VJ-2303/gocraft/releases).

---

## Quick Start Tutorial

### Method 1: Interactive TUI

Launch `gocraft` with no arguments in any interactive terminal:

```bash
gocraft
```

1. **Choose Application Type**: Select between `Web Application` (HTML) or `JSON API`.
2. **Configure Core Settings**: Enter your Go module path (e.g., `github.com/myorg/myapp`), database engine, HTTP router, and config style.
3. **Select Features**: Toggle sessions, authentication, migrations, request logging, or mailers.
   - *Smart Prerequisite Resolution*: Selecting `Full Accounts` automatically activates SQLite, sessions, and emails.
   - *Conflict Warning*: Downgrading a setting (such as switching database to `None` when migrations are enabled) prompts for confirmation before clearing dependents.
4. **Review & Generate**: On the review screen, inspect the summary. Press `[1-9]` to jump directly to any question, or press `Enter` to select the target directory and scaffold your project.

### Method 2: Headless CLI (Scripts & Automation)

Scaffold a project in a single command using CLI flags:

```bash
# Example: Full-featured Web App with SQLite and User Accounts
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

# Example: Lightweight JSON API with PostgreSQL
gocraft \
  --app=api \
  --module=github.com/myorg/myapi \
  --db=postgres \
  --router=servemux \
  --config=env \
  --log=json \
  --migrations \
  --target=./myapi
```

#### CLI Flags Reference

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
| `--json` | `false` | Output machine-readable JSON summary for tools/agents |
| `--quiet` | `false` | Suppress progress messages |
| `--auto-deps` | `true` | Automatically enable required prerequisites |

---

## Working with Your Generated App

Every scaffolded codebase is completely self-contained and runnable out of the box.

### Project Layout

```
├── assets/                  Embedded assets (HTML templates, static assets, SQL migrations, emails)
├── cmd/web|api/             Application entrypoint:
│   ├── main.go              Bootstrap, config parsing, logger, database pool, DI container
│   ├── server.go            HTTP server with graceful shutdown and timeouts
│   ├── routes.go            Route declarations and middleware stack
│   ├── handlers.go          HTTP handler methods
│   ├── helpers.go           Template rendering or JSON encoding helpers
│   ├── middleware.go        Panic recovery, request logging, security headers, auth guards
│   ├── errors.go            Centralized HTTP error response helpers
│   └── *_test.go            End-to-end route and handler tests
├── internal/                Private application packages (only emitted when selected):
│   ├── database/            Connection pool lifecycle and query models (users, tokens)
│   ├── env/                 Type-safe environment variable parsing with validation
│   ├── mailer/              SMTP client with embedded template rendering
│   ├── auth/                JWT token issuance and validation helpers
│   ├── cookies/             HMAC-signed and AES-GCM encrypted cookie helpers
│   ├── validator/           Input validation and form/JSON error mapping
│   ├── response/            Standardized response writers
│   └── version/             Git revision and build metadata helpers
├── Makefile                 Self-documenting build, test, audit, and migration tasks
├── go.mod                   Clean module definition with pinned dependencies
└── README.md                Dedicated getting-started guide for your application
```

### Common Makefile Commands

Change into your generated project directory and use the built-in `Makefile`:

```bash
make help          # List all available targets and descriptions
make run           # Build and run the application
make run/live      # Run with live-reloading via air
make test          # Run test suite with race detector
make test/cover    # Run tests and generate an HTML coverage report
make audit         # Run full quality gate: test, verify, gofmt, vet, staticcheck, govulncheck
make tidy          # Run go mod tidy, go fix, and gofmt
make migrations/up # Apply all pending database migrations (when enabled)
```

---

## Contributing & Development

Interested in how `gocraft` works under the hood, adding new templates, or contributing code?

Read the [Developer Guide (DEVELOPMENT.md)](DEVELOPMENT.md) for internal architecture, testing matrices, code invariants, and release procedures.

---

## License

[MIT](LICENSE) — free for personal, commercial, and open-source use, for both `gocraft` and every codebase it generates.
