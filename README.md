# gocraft

[![CI](https://github.com/VJ-2303/gocraft/actions/workflows/ci.yml/badge.svg)](https://github.com/VJ-2303/gocraft/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/go-1.27-00ADD8?logo=go)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Stop writing the same Go boilerplate. Start building.

`gocraft` is an open-source, terminal-native project generator that scaffolds production-ready Go web applications and JSON APIs in seconds. You answer a few questions (or pass some flags), and it hands you a clean, runnable codebase — structured the way experienced Go developers actually write Go.

The generated code follows the architecture from Alex Edwards' *Let's Go* and *Let's Go Further*: an `application` struct for dependency injection, standard library `net/http` routing, structured `log/slog` logging, graceful shutdown, and a Makefile that documents every command you need. No frameworks to learn. No magic to debug. Every line is yours.

It works two ways: an **interactive TUI** that walks you through choices with live dependency previews, and a **headless CLI** for scripts, CI pipelines, and AI agents.

---

## What You Get

Pick and choose. Every combination produces valid, compiling, tested Go code.

**Application types** — Server-rendered HTML web apps (with template partials and static assets) or JSON APIs (with structured response envelopes and typed error payloads).

**Databases** — SQLite (pure Go, no CGO), PostgreSQL (connection pooling via pgx), MySQL, or none at all for stateless services.

**Routers** — The standard library `http.ServeMux` with Go 1.22+ method routing, or drop in `chi`, `flow`, `gorilla/mux`, or `httprouter`.

**Authentication** — Full web auth with signup, login, logout, email activation, password reset, and CSRF protection. Or JWT bearer tokens for APIs. Or HTTP Basic Auth with bcrypt. Or nothing — your call.

**The rest** — Database-backed sessions, encrypted cookies, embedded SQL migrations, SMTP email templates, request logging middleware, error notification emails, automatic Let's Encrypt TLS, custom error pages, and comprehensive test suites with in-memory database harnesses.

---

## Install

Pick whichever method suits you:

**macOS & Linux**
```bash
curl -fsSL https://raw.githubusercontent.com/VJ-2303/gocraft/main/scripts/install.sh | sh
```

**Windows (PowerShell)**
```powershell
irm https://raw.githubusercontent.com/VJ-2303/gocraft/main/scripts/install.ps1 | iex
```

**Go toolchain** (requires Go 1.27+)
```bash
go install github.com/VJ-2303/gocraft/cmd/gocraft@latest
```

Or grab a binary directly from [GitHub Releases](https://github.com/VJ-2303/gocraft/releases) — every release includes SHA-256 checksums.

---

## Getting Started

### The interactive way

Just run it:

```bash
gocraft
```

The TUI walks you through every decision:

1. **Pick your app type** — Web or API.
2. **Set your module path** — e.g. `github.com/yourname/yourapp`.
3. **Choose your stack** — database, router, config style, log format.
4. **Toggle features** — sessions, auth, migrations, mailer, whatever you need.

The wizard handles dependencies for you. Turn on "Full Accounts" and it knows you need a database, sessions, and an email sender — those light up automatically. Try to remove the database while migrations are still on and it'll ask you to confirm before clearing things out.

When you're happy with the summary, hit Enter, pick a directory, and your project is ready to `make run`.

### The scripted way

One command, no prompts:

```bash
gocraft \
  --app=web \
  --module=github.com/myorg/myapp \
  --db=sqlite \
  --router=chi \
  --migrations \
  --sessions \
  --accounts=full \
  --target=./myapp
```

Or a leaner API project:

```bash
gocraft \
  --app=api \
  --module=github.com/myorg/myapi \
  --db=postgres \
  --config=env \
  --log=json \
  --migrations \
  --target=./myapi
```

Use `--dry-run` to preview the file tree without writing anything. Use `--json` for machine-readable output.

### All flags

| Flag | Default | What it does |
|---|---|---|
| `--app` | `web` | `web` for HTML templates, `api` for JSON |
| `--module` | `example.com/myapp` | Your Go module path |
| `--db` | `none` | `none`, `sqlite`, `postgres`, or `mysql` |
| `--router` | `servemux` | `servemux`, `flow`, `chi`, `gorilla`, or `httprouter` |
| `--config` | `flags` | `flags` for CLI flags, `env` for environment variables |
| `--log` | `plaintext` | `plaintext`, `json`, or `colorized` |
| `--migrations` | `false` | Embedded SQL migrations with Makefile targets |
| `--sessions` | `false` | Server-side sessions via scs (web only) |
| `--accounts` | `none` | `none`, `full` (web), or `jwt` (api) |
| `--request-logging` | `false` | Log every request with method, path, and latency |
| `--basic-auth` | `false` | HTTP Basic Auth with bcrypt-hashed passwords |
| `--git-integration` | `true` | `.gitignore` and build version metadata |
| `--secure-cookies` | `false` | HMAC-signed and AES-encrypted cookies |
| `--auto-https` | `false` | Automatic Let's Encrypt TLS |
| `--custom-error-pages` | `false` | Styled 400/401/404/500 pages (web only) |
| `--emails` | `false` | SMTP mailer and multipart email templates |
| `--error-notifications` | `false` | Email alerts on panics and 500 errors |
| `--tests` | `true` | Handler and route test suites |
| `--target` | `""` | Output directory (must be empty or not exist) |
| `--dry-run` | `false` | Preview file tree without writing |
| `--json` | `false` | Machine-readable JSON output |
| `--quiet` | `false` | Suppress progress messages |
| `--auto-deps` | `true` | Auto-enable prerequisites |

---

## Your Generated Project

The output is a self-contained Go project. No post-generation steps, no hidden dependencies — just `cd` in and start working.

### Structure

```
myapp/
├── cmd/web/                 (or cmd/api/)
│   ├── main.go              Config, logger, DB pool, dependency wiring
│   ├── server.go            HTTP server with graceful shutdown
│   ├── routes.go            All routes and middleware in one place
│   ├── handlers.go          Your HTTP handlers
│   ├── helpers.go           Template rendering or JSON helpers
│   ├── middleware.go        Recovery, logging, security headers, auth
│   ├── errors.go            Centralized error responses
│   └── *_test.go            Tests for every handler
├── assets/                  HTML templates, static files, migrations, emails
├── internal/
│   ├── database/            Connection pool and query models
│   ├── mailer/              SMTP client with template rendering
│   ├── auth/                JWT helpers (API projects)
│   ├── cookies/             Signed & encrypted cookie helpers
│   ├── validator/           Input validation
│   ├── response/            Response writers
│   └── version/             Git revision metadata
├── Makefile                 Every command you need
├── go.mod
└── README.md                Getting-started guide for your app
```

Only the packages you selected are generated. No dead code.

### Day-one commands

```bash
cd myapp
make run           # Build and start the server
make run/live      # Live-reload with air
make test          # Tests with race detector
make audit         # Full quality gate: test + vet + staticcheck + govulncheck
make migrations/up # Run pending migrations (if enabled)
make help          # See everything else
```

---

## Contributing

Want to add a feature, fix a template, or understand how the generator works?

See the [Developer Guide](DEVELOPMENT.md) for architecture, testing, invariants, and release procedures.

---

## License

[MIT](LICENSE) — use it however you want, for any purpose. The tool and every codebase it generates are yours.
