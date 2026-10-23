package generator

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vj/nomoreboiler/internal/spec"
)

// TestGateGenerate writes one example codebase per application type to
// $NMB_GATE_DIR so that external tooling (make gate) can compile and run
// the generated applications as an end-to-end check. It is skipped when the
// environment variable is unset.
func TestGateGenerate(t *testing.T) {
	gateDir := os.Getenv("NMB_GATE_DIR")
	if gateDir == "" {
		t.Skip("NMB_GATE_DIR not set")
	}

	specs := map[string]spec.ProjectSpec{
		"webapp": testSpec(),
		"apiapp": func() spec.ProjectSpec {
			s := testSpec()
			s.AppType = spec.AppAPI
			return s
		}(),
		// Exercises the database package, embedded migrations and
		// external dependency resolution.
		"sqliteapp": func() spec.ProjectSpec {
			s := testSpec()
			s.Database = spec.DBSQLite
			s.Migrations = true
			return s
		}(),
		// Exercises the api assets package without the web-only embeds.
		"apimigrations": func() spec.ProjectSpec {
			s := testSpec()
			s.AppType = spec.AppAPI
			s.Database = spec.DBPostgres
			s.Migrations = true
			return s
		}(),
	}

	for name, s := range specs {
		t.Run(name, func(t *testing.T) {
			rsp := resolveFor(t, s)
			if _, err := Generate(rsp, filepath.Join(gateDir, name), slog.Default(), Options{Tidy: true}); err != nil {
				t.Fatalf("generate: %v", err)
			}
		})
	}
}

// contains reports whether list includes want.
func contains(list []string, want string) bool {
	for _, item := range list {
		if item == want {
			return true
		}
	}
	return false
}

// testSpec returns a minimal valid spec for tests.
func testSpec() spec.ProjectSpec {
	return spec.ProjectSpec{
		AppType:      spec.AppWeb,
		ModulePath:   "github.com/alice/example",
		Database:     spec.DBNone,
		Router:       spec.RouterServeMux,
		ConfigSource: spec.ConfigFlags,
		LogFormat:    spec.LogPlaintext,
	}
}

// resolveFor tests resolves a spec with deterministic port and cookie name.
func resolveFor(t *testing.T, s spec.ProjectSpec) spec.ResolvedSpec {
	t.Helper()

	rsp, err := spec.Resolve(s)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	rsp.Port = 6633
	rsp.SessionCookieName = "session_test1234"
	return rsp
}

func TestGenerateWritesRootFiles(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "example")

	rsp := resolveFor(t, testSpec())
	result, err := Generate(rsp, target, slog.Default(), Options{})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	for _, want := range []string{"go.mod", "Makefile", "README.md"} {
		if _, err := os.Stat(filepath.Join(target, want)); err != nil {
			t.Errorf("expected file %s: %v", want, err)
		}
	}
	if len(result.Files) < 4 {
		t.Errorf("want at least 4 files, got %d: %v", len(result.Files), result.Files)
	}
}

func TestGenerateWebApp(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "example")

	rsp := resolveFor(t, testSpec())
	result, err := Generate(rsp, target, slog.Default(), Options{})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	for _, want := range []string{
		"cmd/web/main.go",
		"cmd/web/routes.go",
		"cmd/web/handlers.go",
		"cmd/web/templates.go",
		"cmd/web/helpers.go",
		"internal/version/version.go",
		"internal/funcs/funcs.go",
		"internal/response/json.go",
		"internal/response/html.go",
		"assets/efs.go",
		"assets/static/css/main.css",
		"assets/html/base.tmpl",
		"assets/html/components/nav.tmpl",
		"assets/html/pages/home.tmpl",
	} {
		if !contains(result.Files, want) {
			t.Errorf("want generated file %s, got %v", want, result.Files)
		}
	}

	// The web main.go must reference the module path and random port.
	got, err := os.ReadFile(filepath.Join(target, "cmd/web/main.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`"github.com/alice/example/assets"`,
		`"http-port", 6633`,
		"response.NewHTMLRenderer",
		`"home.tmpl"`,
	} {
		if !strings.Contains(string(got), want) {
			t.Errorf("cmd/web/main.go want substring %q", want)
		}
	}
}

func TestGenerateAPIApp(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "example")

	s := testSpec()
	s.AppType = spec.AppAPI
	rsp := resolveFor(t, s)

	result, err := Generate(rsp, target, slog.Default(), Options{})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	for _, want := range []string{
		"cmd/api/main.go",
		"cmd/api/routes.go",
		"cmd/api/handlers.go",
		"cmd/api/helpers.go",
		"internal/version/version.go",
		"internal/response/json.go",
	} {
		if !contains(result.Files, want) {
			t.Errorf("want generated file %s, got %v", want, result.Files)
		}
	}

	// API apps must not include web-only files.
	for _, unwanted := range []string{"internal/response/html.go", "assets/html/base.tmpl"} {
		if contains(result.Files, unwanted) {
			t.Errorf("API app should not include %s", unwanted)
		}
	}

	got, err := os.ReadFile(filepath.Join(target, "cmd/api/routes.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "GET /v1/healthcheck") {
		t.Errorf("routes.go want healthcheck route, got:\n%s", got)
	}
}

func TestGenerateGoModContents(t *testing.T) {
	dir := t.TempDir()

	s := testSpec()
	s.Router = spec.RouterChi
	s.Database = spec.DBSQLite
	s.Migrations = true
	rsp := resolveFor(t, s)

	target := filepath.Join(dir, "example")
	if _, err := Generate(rsp, target, slog.Default(), Options{}); err != nil {
		t.Fatalf("generate: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(target, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{
		"module github.com/alice/example",
		"go 1.27",
		"github.com/go-chi/chi/v5",
		"github.com/jmoiron/sqlx",
		"modernc.org/sqlite",
	} {
		if !strings.Contains(string(got), want) {
			t.Errorf("go.mod want substring %q, got:\n%s", want, got)
		}
	}
}

func TestGenerateGitignoreOnlyWhenRequested(t *testing.T) {
	dir := t.TempDir()

	s := testSpec()
	s.Database = spec.DBSQLite
	rsp := resolveFor(t, s)
	rsp.GitIntegration = false

	target := filepath.Join(dir, "example")
	result, err := Generate(rsp, target, slog.Default(), Options{})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	for _, f := range result.Files {
		if f == ".gitignore" {
			t.Error(".gitignore generated without git integration")
		}
	}

	rsp.GitIntegration = true
	target2 := filepath.Join(dir, "example2")
	result2, err := Generate(rsp, target2, slog.Default(), Options{})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	found := false
	for _, f := range result2.Files {
		if f == ".gitignore" {
			found = true
		}
	}
	if !found {
		t.Error(".gitignore missing with git integration enabled")
	}
}

func TestGenerateMakefileMigrations(t *testing.T) {
	dir := t.TempDir()

	s := testSpec()
	s.Database = spec.DBPostgres
	s.Migrations = true
	rsp := resolveFor(t, s)

	target := filepath.Join(dir, "example")
	if _, err := Generate(rsp, target, slog.Default(), Options{}); err != nil {
		t.Fatalf("generate: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(target, "Makefile"))
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{
		"migrations/up",
		"migrations/new",
		"-tags 'postgres'",
		"DB_DSN ?= user:pass@localhost:5432/example?sslmode=disable",
	} {
		if !strings.Contains(string(got), want) {
			t.Errorf("Makefile want substring %q", want)
		}
	}

	// Without migrations the Makefile must not mention them.
	s.Migrations = false
	rsp2 := resolveFor(t, s)
	target2 := filepath.Join(dir, "example2")
	if _, err := Generate(rsp2, target2, slog.Default(), Options{}); err != nil {
		t.Fatalf("generate: %v", err)
	}
	got2, err := os.ReadFile(filepath.Join(target2, "Makefile"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(got2), "migrations") {
		t.Error("Makefile mentions migrations without the migrations option")
	}
}

func TestGenerateRejectsNonEmptyTarget(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "example")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "existing.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	rsp := resolveFor(t, testSpec())
	if _, err := Generate(rsp, target, slog.Default(), Options{}); err == nil {
		t.Fatal("expected an error for a non-empty target, got nil")
	}
}

func TestGenerateAllowsEmptyExistingTarget(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "example")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}

	rsp := resolveFor(t, testSpec())
	if _, err := Generate(rsp, target, slog.Default(), Options{}); err != nil {
		t.Fatalf("generate: %v", err)
	}
}

func TestGenerateRejectsInvalidSpec(t *testing.T) {
	dir := t.TempDir()

	// A spec that fails validation must be rejected before any file is
	// written or directory created.
	s := testSpec()
	s.ModulePath = "invalid module path"

	rsp := spec.ResolvedSpec{ProjectSpec: s}

	target := filepath.Join(dir, "example")
	if _, err := Generate(rsp, target, slog.Default(), Options{}); err == nil {
		t.Fatal("expected generate error for invalid spec")
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Errorf("target directory should not exist, got err %v", err)
	}
}

func TestGenerateSQLiteAppWithMigrations(t *testing.T) {
	dir := t.TempDir()

	s := testSpec()
	s.Database = spec.DBSQLite
	s.Migrations = true
	rsp := resolveFor(t, s)

	target := filepath.Join(dir, "example")
	result, err := Generate(rsp, target, slog.Default(), Options{})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	for _, want := range []string{
		"internal/database/database.go",
		"assets/migrations/000001_init.up.sql",
		"assets/migrations/000001_init.down.sql",
		"assets/efs.go",
	} {
		if !contains(result.Files, want) {
			t.Errorf("want generated file %s, got %v", want, result.Files)
		}
	}

	main, err := os.ReadFile(filepath.Join(target, "cmd/web/main.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`"db-dsn"`,
		`"db-automigrate"`,
		"database.New(",
		"database.MigrateUp(db, assets.MigrationFiles)",
		"db:        db,",
	} {
		if !strings.Contains(string(main), want) {
			t.Errorf("cmd/web/main.go want substring %q", want)
		}
	}

	gomod, err := os.ReadFile(filepath.Join(target, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"github.com/jmoiron/sqlx", "modernc.org/sqlite"} {
		if !strings.Contains(string(gomod), want) {
			t.Errorf("go.mod want %q", want)
		}
	}
}

func TestGenerateAPIMigrationsOmitWebAssets(t *testing.T) {
	dir := t.TempDir()

	s := testSpec()
	s.AppType = spec.AppAPI
	s.Database = spec.DBMySQL
	s.Migrations = true
	rsp := resolveFor(t, s)

	target := filepath.Join(dir, "example")
	_, err := Generate(rsp, target, slog.Default(), Options{})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	// The assets package must exist for migrations but must not embed the
	// web-only static and html directories.
	efs, err := os.ReadFile(filepath.Join(target, "assets/efs.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(efs), "MigrationFiles") {
		t.Error("api efs.go want MigrationFiles embed")
	}
	if strings.Contains(string(efs), "go:embed static") {
		t.Error("api efs.go must not embed static assets")
	}

	// The healthcheck must include the database check.
	handlers, err := os.ReadFile(filepath.Join(target, "cmd/api/handlers.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(handlers), "PingContext") {
		t.Error("api handlers.go want database ping in healthcheck")
	}
}

func TestGenerateNoDatabasePackageWithoutDB(t *testing.T) {
	dir := t.TempDir()

	rsp := resolveFor(t, testSpec())
	target := filepath.Join(dir, "example")
	result, err := Generate(rsp, target, slog.Default(), Options{})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	for _, unwanted := range []string{"internal/database/database.go", "assets/migrations/000001_init.up.sql"} {
		if contains(result.Files, unwanted) {
			t.Errorf("DB-less app should not include %s", unwanted)
		}
	}
}
