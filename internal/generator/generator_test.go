package generator

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/VJ-2303/gocraft/internal/spec"
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
		// Exercises sessions, secure cookies and git integration.
		"sessionapp": func() spec.ProjectSpec {
			s := testSpec()
			s.Database = spec.DBSQLite
			s.Migrations = true
			s.Sessions = true
			s.SecureCookies = true
			s.GitIntegration = true
			s.Router = spec.RouterGorilla
			return s
		}(),
		// Exercises the generated test suites for both app types.
		"testwebapp": func() spec.ProjectSpec {
			s := testSpec()
			s.Database = spec.DBSQLite
			s.Migrations = true
			s.Tests = true
			s.CustomErrorPages = true
			return s
		}(),
		"testjwtapp": func() spec.ProjectSpec {
			s := testSpec()
			s.AppType = spec.AppAPI
			s.Database = spec.DBSQLite
			s.Migrations = true
			s.AccountsJWT = true
			s.Tests = true
			s.ApplyDependencies()
			return s
		}(),
		"testaccountsapp": func() spec.ProjectSpec {
			s := testSpec()
			s.Database = spec.DBSQLite
			s.Migrations = true
			s.AccountsFull = true
			s.Tests = true
			s.ApplyDependencies()
			return s
		}(),
		// Exercises the JWT accounts variant: registration, activation,
		// token issuance and protected routes.
		"jwtapp": func() spec.ProjectSpec {
			s := testSpec()
			s.AppType = spec.AppAPI
			s.Database = spec.DBSQLite
			s.Migrations = true
			s.AccountsJWT = true
			s.ApplyDependencies()
			return s
		}(),
		// Exercises the full user accounts feature: signup, login,
		// password reset, CSRF and session authentication.
		"accountsapp": func() spec.ProjectSpec {
			s := testSpec()
			s.Database = spec.DBSQLite
			s.Migrations = true
			s.AccountsFull = true
			s.ApplyDependencies()
			return s
		}(),
		// Exercises every Plus platform feature together.
		"plusapp": func() spec.ProjectSpec {
			s := testSpec()
			s.Database = spec.DBSQLite
			s.Migrations = true
			s.Sessions = true
			s.SecureCookies = true
			s.CustomErrorPages = true
			s.Emails = true
			s.ErrorNotifications = true
			s.AutoHTTPS = true
			s.Router = spec.RouterFlow
			return s
		}(),
		// Exercises a third-party router end to end.
		"chiapp": func() spec.ProjectSpec {
			s := testSpec()
			s.Router = spec.RouterChi
			s.Database = spec.DBSQLite
			return s
		}(),
		// Exercises the env config source, colorized logs, request
		// logging and basic auth.
		"envapi": func() spec.ProjectSpec {
			s := testSpec()
			s.AppType = spec.AppAPI
			s.ConfigSource = spec.ConfigEnv
			s.LogFormat = spec.LogColorized
			s.RequestLogging = true
			s.BasicAuth = true
			s.Router = spec.RouterHttpRouter
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
		"response.NewHTMLRenderer(assets.HTMLFiles, funcs.All())",
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
		"database.MigrateUp(db, assets.Migrations)",
		`"db-max-lifetime"`,
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

func TestGenerateAllRouters(t *testing.T) {
	tests := []struct {
		router     spec.Router
		wantImport string // quoted import in routes.go
		wantMod    string // unquoted module path in go.mod
		wantRoute  string
	}{
		{spec.RouterServeMux, "", "", `"GET /v1/healthcheck"`},
		{spec.RouterFlow, `"github.com/alexedwards/flow"`, "github.com/alexedwards/flow", `mux.HandleFunc("/v1/healthcheck", app.healthcheck, http.MethodGet)`},
		{spec.RouterChi, `"github.com/go-chi/chi/v5"`, "github.com/go-chi/chi/v5", `mux.Get("/v1/healthcheck"`},
		{spec.RouterGorilla, `"github.com/gorilla/mux"`, "github.com/gorilla/mux", `Methods(http.MethodGet)`},
		{spec.RouterHttpRouter, `"github.com/julienschmidt/httprouter"`, "github.com/julienschmidt/httprouter", `mux.GET("/v1/healthcheck"`},
	}

	for _, tt := range tests {
		t.Run(string(tt.router), func(t *testing.T) {
			s := testSpec()
			s.AppType = spec.AppAPI
			s.Router = tt.router
			rsp := resolveFor(t, s)

			target := filepath.Join(t.TempDir(), "example")
			if _, err := Generate(rsp, target, slog.Default(), Options{}); err != nil {
				t.Fatalf("generate: %v", err)
			}

			got, err := os.ReadFile(filepath.Join(target, "cmd/api/routes.go"))
			if err != nil {
				t.Fatal(err)
			}

			if tt.wantImport != "" && !strings.Contains(string(got), tt.wantImport) {
				t.Errorf("routes.go want import %q, got:\n%s", tt.wantImport, got)
			}
			if !strings.Contains(string(got), tt.wantRoute) {
				t.Errorf("routes.go want route %q, got:\n%s", tt.wantRoute, got)
			}

			// The router dependency must appear in go.mod for third-party
			// routers.
			gomod, err := os.ReadFile(filepath.Join(target, "go.mod"))
			if err != nil {
				t.Fatal(err)
			}
			if tt.wantMod != "" && !strings.Contains(string(gomod), tt.wantMod) {
				t.Errorf("go.mod want %q", tt.wantMod)
			}
			if tt.router == spec.RouterServeMux && strings.Contains(string(gomod), "gorilla") {
				t.Error("servemux app must not require external router")
			}
		})
	}
}

func TestGenerateWebRoutersServeStatic(t *testing.T) {
	tests := []struct {
		router     spec.Router
		wantStatic string
	}{
		{spec.RouterServeMux, `"GET /static/"`},
		{spec.RouterFlow, `"/static/"`},
		{spec.RouterChi, `"/static/*"`},
		{spec.RouterGorilla, `PathPrefix("/static/")`},
		{spec.RouterHttpRouter, `"/static/*filepath"`},
	}

	for _, tt := range tests {
		t.Run(string(tt.router), func(t *testing.T) {
			s := testSpec()
			s.Router = tt.router
			rsp := resolveFor(t, s)

			target := filepath.Join(t.TempDir(), "example")
			if _, err := Generate(rsp, target, slog.Default(), Options{}); err != nil {
				t.Fatalf("generate: %v", err)
			}

			got, err := os.ReadFile(filepath.Join(target, "cmd/web/routes.go"))
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(got), tt.wantStatic) {
				t.Errorf("routes.go want static %q, got:\n%s", tt.wantStatic, got)
			}
			if !strings.Contains(string(got), "app.home") {
				t.Errorf("routes.go want home handler, got:\n%s", got)
			}
		})
	}
}

func TestGenerateEnvConfigAndLogVariants(t *testing.T) {
	tests := []struct {
		name     string
		mutate   func(*spec.ProjectSpec)
		wantFile string
		checks   map[string]string // file → expected substring
	}{
		{
			name:     "env config source",
			mutate:   func(s *spec.ProjectSpec) { s.ConfigSource = spec.ConfigEnv },
			wantFile: "internal/env/env.go",
			checks: map[string]string{
				"cmd/web/main.go": `env.GetString("BASE_URL"`,
			},
		},
		{
			name:     "json logs",
			mutate:   func(s *spec.ProjectSpec) { s.LogFormat = spec.LogJSON },
			wantFile: "",
			checks: map[string]string{
				"cmd/web/main.go": "slog.NewJSONHandler",
			},
		},
		{
			name:     "colorized logs",
			mutate:   func(s *spec.ProjectSpec) { s.LogFormat = spec.LogColorized },
			wantFile: "internal/log/colorized.go",
			checks: map[string]string{
				"cmd/web/main.go": "log.NewColorizedHandler",
			},
		},
		{
			name: "request logging and basic auth",
			mutate: func(s *spec.ProjectSpec) {
				s.RequestLogging = true
				s.BasicAuth = true
			},
			wantFile: "cmd/web/middleware.go",
			checks: map[string]string{
				"cmd/web/middleware.go": "func (app *application) basicAuth(",
				"cmd/web/routes.go":     "app.logRequest(handler)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testSpec()
			tt.mutate(&s)
			rsp := resolveFor(t, s)

			target := filepath.Join(t.TempDir(), "example")
			result, err := Generate(rsp, target, slog.Default(), Options{})
			if err != nil {
				t.Fatalf("generate: %v", err)
			}

			if tt.wantFile != "" && !contains(result.Files, tt.wantFile) {
				t.Errorf("want generated file %s, got %v", tt.wantFile, result.Files)
			}

			for file, want := range tt.checks {
				got, err := os.ReadFile(filepath.Join(target, file))
				if err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(string(got), want) {
					t.Errorf("%s want substring %q", file, want)
				}
			}
		})
	}
}

func TestGenerateSessionsAndCookies(t *testing.T) {
	s := testSpec()
	s.Database = spec.DBSQLite
	s.Migrations = true
	s.Sessions = true
	s.SecureCookies = true
	rsp := resolveFor(t, s)

	target := filepath.Join(t.TempDir(), "example")
	result, err := Generate(rsp, target, slog.Default(), Options{})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	for _, want := range []string{
		"assets/migrations/000002_create_sessions_table.up.sql",
		"internal/cookies/cookies.go",
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
		`"github.com/alexedwards/scs/sqlite3store"`,
		"sessionManager.Store = sqlite3store.New(db.DB)",
		"sessionManager.Cookie.Secure = true",
		`"session-cookie-name"`,
		`"cookie-secret-key"`,
	} {
		if !strings.Contains(string(main), want) {
			t.Errorf("cmd/web/main.go want substring %q", want)
		}
	}

	routes, err := os.ReadFile(filepath.Join(target, "cmd/web/routes.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(routes), "app.sessionManager.LoadAndSave(handler)") {
		t.Error("routes.go want session LoadAndSave middleware")
	}

	// The sessions migration must contain the SQLite session table.
	mig, err := os.ReadFile(filepath.Join(target, "assets/migrations/000002_create_sessions_table.up.sql"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(mig), "CREATE TABLE IF NOT EXISTS sessions") {
		t.Errorf("sessions migration want CREATE TABLE, got:\n%s", mig)
	}

	// Secure cookies helpers must not appear in apps without the option.
	s2 := testSpec()
	rsp2 := resolveFor(t, s2)
	result2, err := Generate(rsp2, filepath.Join(t.TempDir(), "example2"), slog.Default(), Options{})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if contains(result2.Files, "internal/cookies/cookies.go") {
		t.Error("cookies helpers generated without secure cookies option")
	}
}

func TestGeneratePlusFeatures(t *testing.T) {
	s := testSpec()
	s.CustomErrorPages = true
	s.Emails = true
	s.ErrorNotifications = true
	s.AutoHTTPS = true
	rsp := resolveFor(t, s)

	target := filepath.Join(t.TempDir(), "example")
	result, err := Generate(rsp, target, slog.Default(), Options{})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	for _, want := range []string{
		"internal/mailer/mailer.go",
		"assets/emails/example.tmpl",
		"assets/emails/error-notification.tmpl",
		"assets/html/pages/errors/500.tmpl",
		"assets/html/pages/errors/404.tmpl",
		"assets/html/pages/errors/400.tmpl",
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
		`"smtp-host"`,
		`"notifications-email"`,
		`"auto-https-domain"`,
		`"auto-https-email"`,
		`"auto-https-staging"`,
		"mailer.New(",
	} {
		if !strings.Contains(string(main), want) {
			t.Errorf("cmd/web/main.go want substring %q", want)
		}
	}

	serverFile, err := os.ReadFile(filepath.Join(target, "cmd/web/server.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"autocert.Manager",
		"slog.LevelWarn",
		`HostWhitelist(app.config.autoHTTPS.domain)`,
	} {
		if !strings.Contains(string(serverFile), want) {
			t.Errorf("cmd/web/server.go want substring %q", want)
		}
	}

	errorsFile, err := os.ReadFile(filepath.Join(target, "cmd/web/errors.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"reportServerError",
		`app.mailer.Send(app.config.notifications.email, data, "error-notification.tmpl")`,
		`app.render(w, r, http.StatusInternalServerError, "pages/errors/500.tmpl", data)`,
		`errors.As(err, &badRequestErr)`,
	} {
		if !strings.Contains(string(errorsFile), want) {
			t.Errorf("cmd/web/errors.go want substring %q", want)
		}
	}

	helpers, err := os.ReadFile(filepath.Join(target, "cmd/web/helpers.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(helpers), "defaultEmailData") {
		t.Error("cmd/web/helpers.go want defaultEmailData")
	}
}

// matrixSpecs returns the curated permutation set for the compile
// matrix: every router for both app types, every database dialect, both
// configuration sources, every log format, and the feature combinations
// most likely to interact badly with each other.
func matrixSpecs() map[string]spec.ProjectSpec {
	base := func() spec.ProjectSpec {
		s := testSpec()
		return s
	}
	api := func() spec.ProjectSpec {
		s := base()
		s.AppType = spec.AppAPI
		return s
	}

	specs := map[string]spec.ProjectSpec{}
	add := func(name string, mutate func(*spec.ProjectSpec)) {
		s := base()
		if strings.HasPrefix(name, "api-") {
			s = api()
		}
		mutate(&s)
		s.ApplyDependencies()
		specs[name] = s
	}

	// Every router, both app types, no database.
	for _, r := range []spec.Router{spec.RouterServeMux, spec.RouterFlow, spec.RouterChi, spec.RouterGorilla, spec.RouterHttpRouter} {
		router := r
		add(fmt.Sprintf("web-router-%s", router), func(s *spec.ProjectSpec) { s.Router = router })
		add(fmt.Sprintf("api-router-%s", router), func(s *spec.ProjectSpec) { s.Router = router })
	}

	// Every database dialect, with migrations, both app types.
	for _, d := range []spec.Database{spec.DBSQLite, spec.DBPostgres, spec.DBMySQL} {
		dialect := d
		add(fmt.Sprintf("web-db-%s", dialect), func(s *spec.ProjectSpec) {
			s.Database = dialect
			s.Migrations = true
		})
		add(fmt.Sprintf("api-db-%s", dialect), func(s *spec.ProjectSpec) {
			s.Database = dialect
			s.Migrations = true
		})
	}

	// Database without migration helpers.
	add("web-db-nomigrate", func(s *spec.ProjectSpec) {
		s.Database = spec.DBSQLite
		s.Migrations = false
	})

	// Configuration sources and log formats.
	add("web-env-json", func(s *spec.ProjectSpec) {
		s.ConfigSource = spec.ConfigEnv
		s.LogFormat = spec.LogJSON
	})
	add("api-env-colorized", func(s *spec.ProjectSpec) {
		s.ConfigSource = spec.ConfigEnv
		s.LogFormat = spec.LogColorized
	})
	add("web-flags-json", func(s *spec.ProjectSpec) {
		s.ConfigSource = spec.ConfigFlags
		s.LogFormat = spec.LogJSON
	})

	// Free optional features together.
	add("web-features", func(s *spec.ProjectSpec) {
		s.Database = spec.DBSQLite
		s.Migrations = true
		s.RequestLogging = true
		s.BasicAuth = true
		s.GitIntegration = true
		s.SecureCookies = true
	})
	add("api-features", func(s *spec.ProjectSpec) {
		s.RequestLogging = true
		s.BasicAuth = true
		s.GitIntegration = true
		s.SecureCookies = true
	})

	// Sessions across dialects.
	add("web-sessions-pg", func(s *spec.ProjectSpec) {
		s.Database = spec.DBPostgres
		s.Migrations = true
		s.Sessions = true
	})
	add("web-sessions-mysql", func(s *spec.ProjectSpec) {
		s.Database = spec.DBMySQL
		s.Migrations = true
		s.Sessions = true
	})

	// Accounts variants across dialects.
	add("web-accounts-pg", func(s *spec.ProjectSpec) {
		s.Database = spec.DBPostgres
		s.Migrations = true
		s.AccountsFull = true
	})
	add("api-jwt-mysql", func(s *spec.ProjectSpec) {
		s.Database = spec.DBMySQL
		s.Migrations = true
		s.AccountsJWT = true
	})

	// Plus platform features together.
	add("web-plus", func(s *spec.ProjectSpec) {
		s.AutoHTTPS = true
		s.CustomErrorPages = true
		s.Emails = true
		s.ErrorNotifications = true
	})

	// Everything on, both app types.
	add("web-everything", func(s *spec.ProjectSpec) {
		s.Database = spec.DBSQLite
		s.Migrations = true
		s.RequestLogging = true
		s.BasicAuth = true
		s.GitIntegration = true
		s.SecureCookies = true
		s.Sessions = true
		s.AutoHTTPS = true
		s.CustomErrorPages = true
		s.Emails = true
		s.ErrorNotifications = true
		s.Tests = true
		s.AccountsFull = true
	})
	add("api-everything", func(s *spec.ProjectSpec) {
		s.Database = spec.DBSQLite
		s.Migrations = true
		s.RequestLogging = true
		s.BasicAuth = true
		s.GitIntegration = true
		s.SecureCookies = true
		s.AutoHTTPS = true
		s.Emails = true
		s.ErrorNotifications = true
		s.Tests = true
		s.AccountsJWT = true
	})

	return specs
}

// TestMatrixGenerate writes one codebase per matrix spec to
// $NMB_MATRIX_DIR for the make matrix target to compile, vet and test.
// It is skipped when the environment variable is unset.
func TestMatrixGenerate(t *testing.T) {
	matrixDir := os.Getenv("NMB_MATRIX_DIR")
	if matrixDir == "" {
		t.Skip("NMB_MATRIX_DIR not set")
	}

	for name, s := range matrixSpecs() {
		t.Run(name, func(t *testing.T) {
			rsp, err := spec.Resolve(s)
			if err != nil {
				t.Fatalf("resolve: %v", err)
			}
			rsp.Port = 6633
			rsp.SessionCookieName = "session_matrix0001"

			if _, err := Generate(rsp, filepath.Join(matrixDir, name), slog.Default(), Options{Tidy: true}); err != nil {
				t.Fatalf("generate: %v", err)
			}
		})
	}
}
