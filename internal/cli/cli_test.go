package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIDryRun(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := Run([]string{"--app=web", "--db=sqlite", "--dry-run", "--json"}, &stdout, &stderr, nil)
	if code != 0 {
		t.Fatalf("expected code 0, got %d, stderr: %s", code, stderr.String())
	}

	var out Output
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		t.Fatalf("decoding JSON output: %v", err)
	}

	if !out.Success || !out.DryRun {
		t.Errorf("expected success dry-run, got %+v", out)
	}
	if out.Count == 0 || len(out.Files) == 0 {
		t.Errorf("expected planned files in dry-run, got 0")
	}
}

func TestCLIValidationConflict(t *testing.T) {
	var stdout, stderr bytes.Buffer

	// Sessions without database and without auto-deps should fail validation
	code := Run([]string{"--app=web", "--db=none", "--sessions=true", "--auto-deps=false", "--dry-run"}, &stdout, &stderr, nil)
	if code == 0 {
		t.Fatalf("expected validation failure, got code 0")
	}

	if !strings.Contains(stderr.String(), "server-side sessions require a database") {
		t.Errorf("expected error message to mention database requirement, got: %s", stderr.String())
	}
}

func TestCLIGenerate(t *testing.T) {
	target := filepath.Join(t.TempDir(), "cli-app")
	var stdout, stderr bytes.Buffer

	code := Run([]string{"--app=web", "--db=none", "--target", target, "--json"}, &stdout, &stderr, nil)
	if code != 0 {
		t.Fatalf("expected code 0, got %d, stderr: %s", code, stderr.String())
	}

	var out Output
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		t.Fatalf("decoding JSON output: %v", err)
	}

	if !out.Success {
		t.Errorf("expected success, got %+v", out)
	}

	if _, err := os.Stat(filepath.Join(target, "go.mod")); err != nil {
		t.Errorf("expected go.mod to exist in target directory: %v", err)
	}
}
