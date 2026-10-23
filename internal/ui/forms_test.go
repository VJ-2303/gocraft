package ui

import (
	"strings"
	"testing"

	"github.com/vj/nomoreboiler/internal/spec"
)

func TestReviewSummaryIncludesAllAnswers(t *testing.T) {
	s := spec.Defaults()
	s.ModulePath = "github.com/alice/example"
	s.Database = spec.DBSQLite
	s.Migrations = true
	s.Sessions = true
	s.AccountsFull = true
	s.ApplyDependencies()

	got := reviewSummary(&s)

	for _, want := range []string{
		"web", // application type
		"github.com/alice/example",
		"sqlite", // database
		"yes",    // migrations, sessions, emails enabled by dependencies
		"full (web)",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("review summary want substring %q, got:\n%s", want, got)
		}
	}

	// AccountsFull enables emails and sessions via ApplyDependencies.
	if !s.Emails || !s.Sessions {
		t.Fatalf("want emails and sessions enabled, got %+v", s)
	}
}

func TestContains(t *testing.T) {
	if !contains([]string{"a", "b"}, "b") {
		t.Error("want true for present item")
	}
	if contains([]string{"a", "b"}, "c") {
		t.Error("want false for absent item")
	}
	if contains(nil, "a") {
		t.Error("want false for nil list")
	}
}

func TestYesNo(t *testing.T) {
	if yesNo(true) != "yes" || yesNo(false) != "no" {
		t.Error("unexpected yesNo output")
	}
}
