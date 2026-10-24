package ui

import (
	"strings"
	"testing"

	"github.com/VJ-2303/gocraft/internal/spec"
)

func TestApplyConvertsAndResolves(t *testing.T) {
	v := newWizardValues()
	v.appType = "web"
	v.modulePath = "github.com/alice/example"
	v.accounts = "full"

	rsp, err := apply(&v)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	// AccountsFull enables sessions and emails via ApplyDependencies.
	if !rsp.Sessions || !rsp.Emails {
		t.Errorf("want sessions and emails enabled, got %+v", rsp.ProjectSpec)
	}
	if rsp.ProjectName != "example" {
		t.Errorf("want project name example, got %q", rsp.ProjectName)
	}
}

func TestApplyRejectsInvalidValues(t *testing.T) {
	v := newWizardValues()
	v.modulePath = "not a module path"

	if _, err := apply(&v); err == nil {
		t.Fatal("expected an error for an invalid module path")
	}
}

func TestQuestionVisibility(t *testing.T) {
	v := newWizardValues()
	qs := questions(&v)

	count := func(pred func(question) bool) int {
		n := 0
		for _, q := range qs {
			if pred(q) {
				n++
			}
		}
		return n
	}

	// Without a database the migrations and sessions questions are hidden.
	dbQuestions := count(func(q question) bool { return q.title == "SQL migrations" || q.title == "Server-side sessions" })
	if dbQuestions != 2 {
		t.Fatalf("want 2 database questions defined, got %d", dbQuestions)
	}
	for _, q := range qs {
		if q.title == "SQL migrations" && q.visible != nil && q.visible(&v) {
			t.Error("migrations question visible without a database")
		}
		if q.title == "Server-side sessions" && q.visible != nil && q.visible(&v) {
			t.Error("sessions question visible without a database")
		}
	}

	// With web + database they become visible.
	v.database = string(spec.DBSQLite)
	for _, q := range qs {
		if (q.title == "SQL migrations" || q.title == "Server-side sessions") && !q.visible(&v) {
			t.Errorf("%s question hidden with web + database", q.title)
		}
	}

	// Error notifications only after emails.
	if q := findQuestion(qs, "Error notifications"); q.visible != nil && q.visible(&v) {
		t.Error("error notifications visible without emails")
	}
	v.emails = true
	if q := findQuestion(qs, "Error notifications"); !q.visible(&v) {
		t.Error("error notifications hidden with emails")
	}

	// Custom error pages only for web apps.
	v.appType = string(spec.AppAPI)
	if q := findQuestion(qs, "Custom error pages"); q.visible != nil && q.visible(&v) {
		t.Error("custom error pages visible for an API")
	}
}

func TestAccountOptionsFollowAppType(t *testing.T) {
	v := newWizardValues()
	qs := questions(&v)
	q := findQuestion(qs, "User accounts")

	v.appType = string(spec.AppWeb)
	opts := q.options(&v)
	if len(opts) != 2 || !strings.Contains(opts[1].label, "Full") {
		t.Errorf("web app want full variant option, got %v", opts)
	}

	v.appType = string(spec.AppAPI)
	opts = q.options(&v)
	if len(opts) != 2 || !strings.Contains(opts[1].label, "JWT") {
		t.Errorf("api want jwt variant option, got %v", opts)
	}
}

func TestReviewSummaryListsAnswers(t *testing.T) {
	v := newWizardValues()
	v.appType = "web"
	v.modulePath = "github.com/alice/example"
	v.database = "sqlite"
	v.migrations = true
	v.accounts = "full"
	v.targetDir = "./example"

	got := reviewSummary(&v)

	for _, want := range []string{
		"web",
		"github.com/alice/example",
		"sqlite",
		"yes",
		"Full (web)",
		"./example",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("review summary want substring %q, got:\n%s", want, got)
		}
	}
}

// findQuestion returns the question with the given title.
func findQuestion(qs []question, title string) question {
	for _, q := range qs {
		if q.title == title {
			return q
		}
	}
	panic("question not found: " + title)
}
