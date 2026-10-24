# #0002 API with SQLite but no migrations fails test compilation because assets.Migrations is unguarded

- 2026-10-24T10:30:32Z `issue`: API with SQLite but no migrations fails test compilation because assets.Migrations is unguarded [templates/api/cmd/api/main_test.go.tmpl:104]
- 2026-10-24T10:30:58Z `attempt`: Guard MigrateUp with {{if .Migrations}} in main_test.go.tmpl and ensure assets import when Migrations or NeedsMailer (worked)
- 2026-10-24T10:32:03Z `fix`: Guard MigrateUp with {{if .Migrations}} and include assets import when Migrations or NeedsMailer
