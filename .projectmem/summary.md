# projectmem - nomoreboiler

_Last updated: 2026-10-24_

## Project purpose
A terminal UI for generating production-ready Go web application and JSON API codebases — no boilerplate, no frameworks, no lock-in.

## Recent issues
- [DONE] #0003 PostgreSQL does not support driver LastInsertId; UserModel.Insert must use RETURNING id for postgres [templates/shared/internal/database/users.go.tmpl:33] -> Use RETURNING id for postgres in UserModel.Insert (fixed)
- [DONE] #0002 API with SQLite but no migrations fails test compilation because assets.Migrations is unguarded [templates/api/cmd/api/main_test.go.tmpl:104] -> Guard MigrateUp with {{if .Migrations}} and include assets import when Migrations or NeedsMailer (fixed)
- [DONE] #0001 Duplicate mux.GET("/") route registration in httprouter.go.tmpl panics at runtime [templates/web/cmd/web/routes/httprouter.go.tmpl:56] -> Removed duplicate mux.GET("/") route registration in httprouter.go.tmpl and cleaned redundant comments across templates (fixed)
- [DONE] #legacy_f8e1 Legacy issue: merge: close second-round parity gaps -> merge: close second-round parity gaps (fixed)
- [DONE] #legacy_c333 Legacy issue: style: gofmt resolve.go after version pin updates -> style: gofmt resolve.go after version pin updates (fixed)
- [DONE] #legacy_c29f Legacy issue: merge: close Autostrada parity gaps -> merge: close Autostrada parity gaps (fixed)
- [DONE] #legacy_8367 Legacy issue: fix: close all second-round Autostrada parity gaps -> fix: close all second-round Autostrada parity gaps (fixed)
- [DONE] #legacy_79b8 Legacy issue: fix: wizard contradictions, stuck navigation, and complete redesign -> fix: wizard contradictions, stuck navigation, and complete redesign (fixed)
- [DONE] #legacy_221e Legacy issue: fix: close all seven Autostrada parity gaps from the feature audit -> fix: close all seven Autostrada parity gaps from the feature audit (fixed)

## Decisions
- No decisions logged yet.

## Notes
- feat: full permutation matrix, hardened feature interactions
- merge: milestone 14 — permutation matrix and polish
- style: rewrite HTML and email templates in clean Let's Go fashion
- merge: wizard redesign and conditional correctness
- feat: generator formats every emitted Go file; matrix enforces code hygiene
- style: drop unused style vars flagged by staticcheck
- docs: detailed README and compact AGENTS.md reference for agents
- style: compact wizard — one-line question header, no brand chrome
- style: gofmt test files
- gotcha: GoReleaser v2 deprecates singular format in favor of formats list; PowerShell 5.1 requires explicit TLS 1.2 SecurityProtocol to query GitHub API

## Key files
- `resolve.go`
- `internal/spec/resolve.go`
- `internal/generator/files.go`
- `internal/generator/generator_test.go`
- `templates/api/cmd/api/context.go.tmpl`
- `templates/api/cmd/api/handlers-auth.go.tmpl`
- `templates/api/cmd/api/main.go.tmpl`
- `templates/api/cmd/api/middleware.go.tmpl`
- `templates/api/cmd/api/routes/chi.go.tmpl`
- `templates/api/cmd/api/routes/flow.go.tmpl`
- `templates/api/cmd/api/routes/gorilla.go.tmpl`
- `templates/api/cmd/api/routes/httprouter.go.tmpl`
- `templates/api/cmd/api/routes/servemux.go.tmpl`
- `templates/shared/assets/emails/account-activation.tmpl`
- `templates/shared/assets/migrations/000003_create_users_table.up.sql.tmpl`
- `templates/shared/assets/migrations/000004_create_activation_tokens_table.down.sql.tmpl`
- `templates/shared/assets/migrations/000004_create_activation_tokens_table.up.sql.tmpl`
- `templates/shared/internal/auth/jwt.go.tmpl`
- `templates/shared/internal/database/activation_tokens.go.tmpl`
- `templates/shared/internal/database/users.go.tmpl`

## Open questions
- None logged yet.
