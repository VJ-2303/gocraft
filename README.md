# nomoreboiler

A terminal UI for generating production-ready Go web application and JSON API
codebases — no boilerplate, no frameworks, no lock-in.

`nomoreboiler` scaffolds a complete, runnable Go project in the idiomatic
style popularized by Alex Edwards' *Let's Go* books: a `cmd/` +
`internal/` layout, an `application` struct with dependency injection,
`log/slog` logging, graceful shutdown, and a Makefile of admin commands.
Every line of the generated code is yours to keep and edit.

Generated projects support (optionally): SQLite/PostgreSQL/MySQL with SQL
migrations, five router choices, flag- or environment-based configuration,
plaintext/JSON/colorized logging, request logging, basic auth, Git
integration, secure cookies, server-side sessions, automatic HTTPS, custom
error pages, email sending, error notifications, user accounts with
authentication, and full test coverage.

## Installation

```
go install github.com/vj/nomoreboiler/cmd/nomoreboiler@latest
```

## Usage

```
nomoreboiler
```

Then follow the three-step wizard and point it at an empty directory.

## Status

Under active development. See the git history for progress.
