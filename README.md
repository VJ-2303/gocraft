# nomoreboiler

A terminal UI for generating production-ready Go web application and JSON
API codebases — no boilerplate, no frameworks, no lock-in.

`nomoreboiler` scaffolds a complete, runnable Go project in the idiomatic
style popularized by Alex Edwards' *Let's Go* books: a `cmd/` +
`internal/` layout, an `application` struct with dependency injection,
`log/slog` logging, graceful shutdown, and a Makefile of admin commands.
Every line of the generated code is yours to keep and edit — MIT
licensed, no attribution required.

This tool replicates the full feature set of Autostrada, including its
paid "Plus" tier, free of charge.

## Installation

```
go install github.com/vj/nomoreboiler/cmd/nomoreboiler@latest
```

## Usage

```
nomoreboiler
```

Follow the three-step wizard:

1. **Application type** — traditional web application or JSON API.
2. **Customize** — database (none, SQLite, PostgreSQL, MySQL), SQL
   migration helpers, router, configuration source, log format, request
   logging, basic auth, Git integration, secure cookies and (for web
   apps with a database) server-side sessions.
3. **Plus features** — automatic HTTPS, custom error pages, emails,
   error notifications, full test coverage, and user accounts with
   authentication (session-based for web apps, JWT for APIs).

Then point it at an empty directory. The generated codebase is
immediately runnable with `make run`.

## What the generated codebase contains

- `cmd/web/` or `cmd/api/` — `main.go` with the `run()` startup pattern
  and graceful shutdown, `routes.go`, `handlers.go`, `helpers.go`
  (including `backgroundTask()`), `middleware.go`, and HTML template
  rendering for web apps.
- `internal/` — ancillary packages, emitted only when used: `database`
  (sqlx pool + embedded migration runner), `env`, `version`,
  `request`, `response`, `validator`, `password`, `mailer`, `cookies`,
  `log`, `auth` and `funcs`.
- `assets/` — embedded static files, HTML templates, email templates
  and SQL migrations.
- `Makefile` — `help`, `audit`, `test`, `test/cover`, `upgradeable`,
  `tidy`, `build`, `run`, `run/live` (Air live reloading) and, when
  migrations are selected, the `migrations/*` targets backed by
  golang-migrate.

Configuration is available through command-line flags or environment
variables (`--help` lists every setting), with a randomized default
HTTP port per codebase.

## Development

Requires Go 1.27+.

```
make audit    # tests, vet, staticcheck, govulncheck
make gate     # generate example codebases, build + vet + test each
make matrix   # generate ~34 permutations (routers × dialects × features), build + vet + test each
```

The template library lives in `templates/` and is embedded into the
binary via `go:embed`; the file-selection matrix is
`internal/generator/files.go`.

## License

MIT — for both this tool and every codebase it generates.
