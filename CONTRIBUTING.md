# Contributing to gocraft

Thanks for considering a contribution!

## Quick start

```bash
git clone https://github.com/VJ-2303/gocraft.git
cd gocraft
go test ./...          # fast — unit + headless wizard tests
make audit             # vet + staticcheck + govulncheck
make gate              # end-to-end: 13 generated codebases build / vet / test
make matrix            # exhaustive: 29 permutations (slow, ~2 min)
```

`make gate` is the fast loop. `make matrix` is required when you touch `templates/` or `internal/generator/files.go`.

If `make gate`/`make matrix` complains about a non-empty target after an interrupted run:

```bash
rm -rf /tmp/gocraft-gate /tmp/gocraft-matrix
```

## What to read first

- `README.md` — what the tool does for users.
- `DEVELOPMENT.md` — architecture, invariants, and template conventions.
- `AGENTS.md` — the exact checks that must pass before every commit.

## Guidelines

- Keep changes surgical — touch only what the task needs. Don't reformat unrelated files.
- Match the existing code style. Generated code follows Alex Edwards conventions (`application` struct, `run()` entrypoint, `errors.As` on `*BadRequestError`, `slog` groups).
- If you add a new emitted file, register it in `internal/generator/files.go` with the correct `Cond`.
- If you add a new third-party dependency for generated code, pin it in `internal/spec/resolve.go` (`requireVersions` + `requires()`).
- HTML/email/SQL assets must be registered with `Copy: true` so they are not double-rendered.

## Pull requests

- One branch per unit of work (`feat/...`, `fix/...`).
- Conventional commits (`feat:`, `fix:`, `docs:`, `chore:`).
- CI must be green — `go test ./...` and `gofmt -l` are the minimum gate.

## Questions?

Open an issue or start a discussion — happy to help.
