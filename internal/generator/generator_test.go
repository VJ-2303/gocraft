package generator

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vj/nomoreboiler/internal/spec"
)

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
	result, err := Generate(rsp, target, slog.Default())
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	wantFiles := []string{"go.mod", "Makefile", "README.md"}
	if len(result.Files) != len(wantFiles) {
		t.Fatalf("want files %v, got %v", wantFiles, result.Files)
	}
	for _, want := range wantFiles {
		if _, err := os.Stat(filepath.Join(target, want)); err != nil {
			t.Errorf("expected file %s: %v", want, err)
		}
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
	if _, err := Generate(rsp, target, slog.Default()); err != nil {
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
	result, err := Generate(rsp, target, slog.Default())
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
	result2, err := Generate(rsp, target2, slog.Default())
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
	if _, err := Generate(rsp, target, slog.Default()); err != nil {
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
	if _, err := Generate(rsp2, target2, slog.Default()); err != nil {
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
	if _, err := Generate(rsp, target, slog.Default()); err == nil {
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
	if _, err := Generate(rsp, target, slog.Default()); err != nil {
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
	if _, err := Generate(rsp, target, slog.Default()); err == nil {
		t.Fatal("expected generate error for invalid spec")
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Errorf("target directory should not exist, got err %v", err)
	}
}
