# gocraft Developer Guide

Operational and technical documentation for developing, testing, and contributing to `gocraft`.

---

## Architecture Overview

`gocraft` uses a strictly decoupled 4-layer architecture:

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

### Layer Responsibilities

- **`cmd/gocraft/`**: Application entrypoint. Inspects flags, initializes logging, detects interactive terminal vs. headless invocation, and routes execution to `internal/ui` or `internal/cli`.
- **`internal/ui/`**: Bubble Tea interactive TUI. Manages multi-step wizard state, real-time dependency cascade visualization, conflict resolution, and the review screen.
- **`internal/cli/`**: Headless runner for scripts and AI agents. Parses CLI flags, validates arguments, and outputs plain text or structured JSON.
- **`internal/spec/`**: Core domain logic and single source of truth:
  - `spec.go`: `ProjectSpec` definition capturing user choices.
  - `validate.go`: Invariant and dependency validation rules.
  - `resolve.go`: Derives `ResolvedSpec` view-model, third-party dependency versions, and connection strings.
- **`internal/generator/`**: Code generation engine:
  - `files.go`: The file matrix mapping every template to output path with evaluation conditions (`Cond`) and verbatim copy flags (`Copy`).
  - `generator.go`: Template renderer and disk writer. Runs all generated Go files through `go/format` before writing.
- **`templates/`**: Embedded template assets (`embed.FS`):
  - `templates/common/`: Root files (Makefile, README, .gitignore, go.mod).
  - `templates/shared/`: Shared packages and assets (database, mailer, validator, env, auth).
  - `templates/web/`: Web-specific templates, handlers, and static assets.
  - `templates/api/`: API-specific handlers, middleware, and response helpers.

---

## Repository Layout

| Path | Purpose |
|---|---|
| `cmd/gocraft/main.go` | Application entrypoint |
| `internal/cli/` | Headless CLI parser and runner |
| `internal/generator/` | File matrix, rendering engine, and gofmt validator |
| `internal/spec/` | Spec structs, validation rules, dependency resolution |
| `internal/ui/` | Bubble Tea wizard model, views, validation, and theme |
| `scripts/` | Cross-platform installation scripts (`install.sh`, `install.ps1`) |
| `templates/` | Embedded Go template files and verbatim assets |
| `.github/workflows/` | GitHub Actions release automation |

---

## Local Development & Testing

### Prerequisites

- Go 1.27 or later
- Make

### Verification Commands

Before submitting pull requests or committing changes, run the test suites:

```bash
# Run unit tests and headless wizard flow tests (fast)
go test ./...

# Run full tool audit (tests + vet + staticcheck + govulncheck)
make audit

# Generate 13 representative codebases and verify: build, vet, test
make gate

# Generate 29 permutations and run full audit:
# build + vet + test + gofmt + stray-file + comment-hygiene audit
make matrix
```

> [!TIP]
> `make gate` is the recommended fast verification loop for local iteration. `make matrix` runs exhaustive 29-combination audits and is required when editing template files.

If `make gate` or `make matrix` fails with "target directory is not empty" after an interrupted run, clean the temporary directories first:

```bash
rm -rf /tmp/nmb-gate /tmp/nmb-matrix
```

---

## Invariants & Design Rules

When making changes to `gocraft`, ensure these invariants are preserved:

1. **Gofmt-Clean Output**: All generated `.go` files must parse and be gofmt-clean. The generator executes `go/format` on every Go file; template syntax errors fail generation immediately.
2. **Deterministic Validation**: Every option combination tested in `make matrix` must pass `spec.Validate()` and produce a working codebase.
3. **Dependency Cascades**:
   - `AccountsFull` requires database (`sqlite` default if none), sessions, and emails.
   - `AccountsJWT` requires database (`sqlite` default if none) and emails.
   - `Notifications` requires emails.
   - Web-only settings (`sessions`, `custom_error_pages`, `accounts=full`) must never appear in API projects.
4. **No Code Smells in Generated Output**: Generated codebases must never contain `TODO`, `FIXME`, stray artifacts (`.sqlite`, `.air.toml`, logs), or plaintext passwords.
5. **No Third-Party Frameworks**: Generated projects follow standard library `net/http` conventions and Alex Edwards application structure (explicit `application` struct dependency injection, `run()` entrypoint, structured logging).

---

## Template Development Rules

- **Registration**: Every new emitted file must be registered in `internal/generator/files.go` with its corresponding condition (`Cond`).
- **Verbatim Assets**: Files with `Copy: true` (HTML templates, emails, SQL migrations) are copied verbatim to avoid double-evaluation of template directives.
- **Dependency Pinning**: New third-party packages must have their version pinned in `requireVersions` (`internal/spec/resolve.go`) and registered in `requires()`.
- **Template Escaping**: Literal `{{ }}` inside rendered templates (e.g. format strings in Makefiles) must be escaped as `{{ "{{" }}...{{ "}}" }}`.

---

## Release Process

Releases are fully automated via GitHub Actions and GoReleaser.

1. Ensure all local tests and gates pass:
   ```bash
   make audit
   make gate
   ```
2. Create and push a semver tag:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```
3. GitHub Actions triggers `.github/workflows/release.yml`, runs `make audit`, builds multi-platform archives (Linux, macOS, Windows for amd64 and arm64), generates checksums, and publishes the release.
