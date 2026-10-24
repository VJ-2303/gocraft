# AGENTS.md — working reference for agents

Operational facts for AI agents and contributors working on this repo.
The README explains what the tool is; this explains how to change it
without breaking invariants. No section repeats the README.

## Verify before every commit

```
go test ./...     # fast; unit + headless wizard flow tests
make audit        # tests + vet + staticcheck + govulncheck (tool itself)
make gate         # 13 representative codebases: build + vet + test
make matrix       # 29 permutations: build + vet + test + gofmt + stray-file
                  # + comment-hygiene audit — REQUIRED for template changes
```

Never commit red. `make matrix` is slow (~2 min, downloads deps per
codebase); `make gate` is the fast loop. Regenerate with `rm -rf
/tmp/nmb-gate` first if the directory check fails — generation refuses
non-empty targets.

## Repo layout

| Path | Purpose |
|---|---|
| `cmd/gocraft/main.go` | Tool entrypoint: CLI vs TUI detection, flags, slog dispatcher |
| `internal/cli/cli.go` | Headless CLI runner: flag parsing, `--dry-run`, `--json`, validation |
| `internal/spec/spec.go` | `ProjectSpec` — choice representation and constants |
| `internal/spec/validate.go` | Option legality + dependency rules (source of truth) |
| `internal/spec/resolve.go` | `ResolvedSpec` view-model, `ApplyDependencies()`, drivers, `requireVersions` pins, `BasicAuthHashedPassword` |
| `internal/generator/generator.go` | Render + write loop; gofmt via `go/format`; `go mod tidy` |
| `internal/generator/files.go` | **The file matrix** — every emitted file, with `Cond` and `Copy` |
| `internal/ui/ui.go` | Interactive TUI entrypoint (`ui.Run`) |
| `internal/ui/questions.go` | Question definitions, options, validators, and visibility rules |
| `internal/ui/model.go` | Bubble Tea state machine, navigation, cascade resolution, rendering |
| `internal/ui/values.go` | `wizardValues` state store and `toSpec()` conversion |
| `internal/ui/theme.go` | Lipgloss styles, color palette, and layout helpers |
| `scripts/` | Cross-platform installation scripts (`install.sh`, `install.ps1`) |
| `.goreleaser.yaml` | Multi-platform build and archive packaging configuration |
| `.github/workflows/` | GitHub Actions automation (`release.yml` on tag push) |
| `templates/common/` | Root files: go.mod, Makefile, README (rendered) |
| `templates/shared/` | Packages/assets emitted for both app types (rendered + verbatim) |
| `templates/web/`, `templates/api/` | App-type-specific files |
| `testdata`-free testing | Tests generate into `t.TempDir()` or `/tmp` via env-var gates |

## Template conventions

- Files ending `.tmpl` are rendered with `text/template` against
  `spec.ResolvedSpec`. `missingkey=error` is on — any field typo fails
  generation.
- HTML/email/SQL assets are registered with `Copy: true` and written
  verbatim; they contain their own Go template syntax and must never be
  rendered twice.
- Every `.go` output is run through `go/format` by the generator. A file
  that doesn't parse fails generation with the file named — this is the
  syntax gate for template edits, don't defeat it.
- **Every new emitted file must be registered in `files.go`** with the
  right `Cond`. Forgetting this silently omits the file.
- Import blocks in templates: use `{{- if}}` trimming carefully; the
  generator's gofmt pass fixes alignment but cannot fix missing/extra
  imports — vet in the matrix catches those.
- Literal `{{ }}` inside a rendered template (e.g. the Makefile's
  `go list` format string) must be escaped as
  `{{"{{if ...}}{{end}}"}}`.
- New third-party deps for generated code: pin the version in
  `requireVersions` (internal/spec/resolve.go) AND add the `add(...)`
  call in `requires()`. `go mod tidy` in generated projects resolves
  store sub-packages (e.g. `github.com/alexedwards/scs/sqlite3store`,
  no `/v2`, pseudo-versions) automatically.

## Invariants (break these and the gates should fail)

1. Generated `.go` files are always gofmt-clean and parse — enforced in
   the generator write path and audited by `make matrix`.
2. Every option combination in the matrix produces a valid spec;
   `spec.Validate()` runs before generation in both UI and CLI.
3. Wizard & CLI: Prerequisites auto-enable via cascade logic in UI and
   `spec.ApplyDependencies()` in CLI (full accounts ⇒ sqlite+sessions+emails;
   jwt ⇒ sqlite+emails; notifications ⇒ emails); `resetInvisible` resets
   answers of questions current answers make invisible — a stale web-only
   answer can never reach an API spec.
4. Question visibility funcs depend only on other answers, never on the
   question's own value.
5. Generated code style: Alex Edwards conventions — `run()` entrypoint,
   `application` struct DI, `errors.As` on `*response.BadRequestError`,
   `(User, bool)` found-style returns, slog with grouped fields, Warn
   ErrorLog, no panics outside `newLogger`-style startup helpers.
6. Generated codebases never contain: TODO/FIXME, stray artifacts
   (*.sqlite, logs, .air.toml, tmp/), plaintext passwords (basic auth
   default is a bcrypt hash), or references to gocraft beyond the
   README attribution line.

## Testing patterns

- **TUI flow testing**: No TTY needed — drive `model` headlessly with
  `m := newModel(mockGen, nil)` and helper `answer(m, keyRunes("2"))` /
  `answer(m, key(tea.KeyEnter))` (helpers in `internal/ui/ui_test.go`),
  then inspect `m.values`, `m.state`, and `m.resolvedSpec`.
- **Headless CLI testing**: Drive runner directly via
  `cli.Run(args, stdout, stderr, logger, mockGen)` in
  `internal/cli/cli_test.go`. Test flags, `--dry-run`, `--json`, and
  validation rejections.

## Known traps

- `make gate`/`make matrix` fail with "target directory is not empty"
  after a partial run: `rm -rf /tmp/nmb-gate /tmp/nmb-matrix` first.
- scs database stores are separate nested modules
  (`github.com/alexedwards/scs/{sqlite3store,postgresstore,mysqlstore}`);
  the core is `/v2`. Imports must not carry `/v2` on the store path.
- pgx needs `default_query_exec_mode=simple_protocol` in the example DSN
  for multi-statement migrations.
- `url.Clone` does not exist in Go 1.27 — `funcs.urlSetParam` copies via
  `*url.URL` manually.
- env config uses `NOTIFICATIONS_EMAIL`/`BASIC_AUTH_HASHED_PASSWORD` (not
  the admin-email/plaintext names from earlier iterations).
- Config struct fields are nested per feature (`cfg.db.dsn`,
  `cfg.basicAuth.hashedPassword`, `cfg.smtp.host`) — keep flags, env
  reads and struct fields in sync across flags/env variants of
  `cmd/*/main.go.tmpl`.

## Git and release conventions

Conventional commits (`feat:`, `fix:`, `style:`, `chore:`), one branch
per unit of work (`feat/...`, `fix/...`), `--no-ff` merges to `main`.
Releases trigger automatically by pushing semver tags
(`git tag -a v1.0.0 -m "Release v1.0.0"` and `git push origin v1.0.0`),
which runs `.github/workflows/release.yml` with GoReleaser cross-compilation
and publishing. Normal commits do not trigger CI runners.
