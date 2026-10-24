# nomoreboiler

A terminal UI for generating production-ready Go web application and JSON
API codebases — no boilerplate, no frameworks, no lock-in.

`nomoreboiler` scaffolds a complete, runnable Go project in the idiomatic
style popularized by Alex Edwards' *Let's Go* books: a `cmd/` + `internal/`
layout, an `application` struct with dependency injection, `log/slog`
logging, graceful shutdown, and a Makefile of admin commands. Every line of
the generated code is yours to keep and edit — MIT licensed, no attribution
required.

It replicates the full feature set of [autostrada.dev](https://autostrada.dev)
— including its paid Plus tier — free of charge, with its own hand-written
template library.

## Installation

Requires Go 1.27 or later.

```
go install github.com/vj/nomoreboiler/cmd/nomoreboiler@latest
```

## Usage

```
nomoreboiler
```

The wizard asks one question per screen and answers accept a single
keypress. Use `esc` to go back at any point and `ctrl+c` to abort.

### Step 1 — Application type

- **Traditional web application** — server-side HTML rendering
- **JSON API** — processes JSON requests and sends JSON responses

The choice drives the whole codebase: web apps get HTML templates,
partials, static file serving and session-based accounts; APIs get JSON
envelopes, bearer-token authentication and `Error`/`FieldErrors` payloads.

### Step 2 — Customize

| Setting | Choices |
|---|---|
| Database | None · SQLite · PostgreSQL · MySQL |
| Router | http.ServeMux · Flow · Chi · Gorilla Mux · HttpRouter |
| Configuration | Command-line flags · Environment variables |
| Log format | Plaintext · JSON · Colorized plaintext |

With a database selected, SQL migration helpers are available: an embedded
migrations directory, Makefile targets backed by golang-migrate, and
automatic migration on startup. Web apps with a database can also enable
server-side sessions backed by alexedwards/scs.

Optional features: request logging (client IP, method, URL, status, body
size), HTTP basic auth with a bcrypt-hashed password, Git integration
(`.gitignore` plus version derived from Git metadata), and signed and
encrypted cookie helpers.

### Step 3 — Plus features

All of Autostrada's paid tier, included free:

- **Automatic HTTPS** — Let's Encrypt certificates via autocert, with
  `--auto-https-domain/email/staging` settings, an HTTP-to-HTTPS redirect
  server and a `certs/` cache.
- **Custom error pages** — per-status templates in `pages/errors/`.
- **Emails** — templates with `subject`, `plainBody` and `htmlBody`
  blocks, sent via SMTP through `internal/mailer`.
- **Error notifications** — runtime errors emailed to
  `--notifications-email`; disabled unless an address is configured.
- **Tests** — full handler test suites with standardized helpers.
- **User accounts** — web: signup, login, logout, password reset, CSRF
  protection; API: registration with email activation and JWT
  authentication.

### Step 4 — Generate

The review screen summarizes every choice; confirming generates the
codebase into the directory you pick (it must be empty), runs `go mod
tidy`, and prints tailored getting-started instructions.

## What a generated codebase contains

```
├── assets/            Embedded static files, HTML templates, emails, migrations
├── cmd/web|api/       main.go, server.go, routes.go, handlers.go,
│                      helpers.go, middleware.go, errors.go (+ auth handlers, context)
├── internal/          Emitted only when used:
│                      database (sqlx pool + embedded migration runner), env,
│                      version, request, response, validator, password,
│                      mailer, cookies, log, auth, funcs
├── go.mod             Your module path; pinned, known-good dependency versions
├── Makefile           help · audit · test · test/cover · upgradeable · tidy ·
│                      build · run · run/live (+ migrations/* when selected)
└── README.md          Tailored getting-started instructions
```

Every generated file is gofmt-clean by construction, `go vet`-clean, and
ships with tests when the Tests option is selected. Configuration is read
from flags or environment variables (`--help` lists every setting), with a
randomized default HTTP port per codebase to avoid clashes.

## How the tool works

Three pieces, wired in one direction:

```
internal/spec        ProjectSpec (wizard answers) → validation → ResolvedSpec
                     (view-model: dialects, drivers, random port, go.mod
                     requirements, dependency enforcement)
internal/generator   File-selection matrix + embedded text/template rendering;
                     .go output is formatted with go/format; anything that
                     fails to parse fails generation with the file named
internal/ui          A single Bubble Tea model: one question per screen,
                     dynamic question visibility with answer resets,
                     dependency enforcement, review, generation, outcome
templates/           The template library, embedded via go:embed
```

The file-selection matrix in `internal/generator/files.go` decides which
templates are emitted for which answers. Rendered templates receive the
`ResolvedSpec`; HTML, email and SQL templates are copied verbatim (they
contain their own Go template syntax).

Question visibility and dependencies live in `internal/ui/wizard.go`:
invisible questions have their answers reset after every commit, and
`enforceDependencies` turns on anything a selection implies, so the review
screen can never contradict an earlier answer.

## Development

```
make help          # all targets
make audit         # tool: tests, vet, staticcheck, govulncheck
make gate          # generate 13 representative codebases; build, vet, test each
make matrix        # generate 29 permutations; build, vet, test, gofmt,
                   # stray-file and comment-hygiene audit for each
```

The matrix is the safety net for template changes: it covers every router
for both application types, every database dialect with and without
migrations, both configuration sources, all log formats, feature bundles,
and everything-on combinations. Run it before merging anything that
touches `templates/` or `internal/generator/`.

Contributions and agents should read [AGENTS.md](AGENTS.md) — a compact
reference covering the template conventions, invariants and verification
loop this repository expects.

## License

MIT — for both this tool and every codebase it generates.
