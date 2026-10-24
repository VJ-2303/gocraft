package ui

import (
	"github.com/VJ-2303/gocraft/internal/spec"
	"testing"

	"github.com/VJ-2303/gocraft/internal/generator"
	tea "github.com/charmbracelet/bubbletea"
)

// keyRunes builds a runes key message, the way a terminal delivers
// typed characters.
func keyRunes(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

// key builds a single special-key message.
func key(t tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: t}
}

// answer drives the model with one key message.
func answer(m wizardModel, k tea.KeyMsg) (wizardModel, tea.Cmd) {
	model, cmd := m.Update(k)
	return model.(wizardModel), cmd
}

// assertWizard converts a Bubble Tea model back to the wizard model.
func assertWizard(model tea.Model, cmd tea.Cmd) wizardModel {
	_ = cmd
	return model.(wizardModel)
}

// TestWizardAPIFlowDoesNotAskWebQuestions reproduces the reported bug:
// choosing API must never ask the Custom error pages question, and the
// generated spec must not contain web-only answers.
func TestWizardAPIFlowDoesNotAskWebQuestions(t *testing.T) {
	var generated bool
	generate := func(rsp spec.ResolvedSpec, target string) (generator.Result, error) {
		generated = true
		return generator.Result{Dir: target}, nil
	}

	m := newWizardModel(generate)

	// Application type: JSON API.
	m, _ = answer(m, keyRunes("2"))
	if m.v.appType != "api" {
		t.Fatalf("want api app type, got %q", m.v.appType)
	}

	// Module path.
	for _, r := range "github.com/vj2303/main" {
		m, _ = answer(m, keyRunes(string(r)))
	}
	m, _ = answer(m, key(tea.KeyEnter))

	// Database: SQLite.
	m, _ = answer(m, keyRunes("2"))
	// Router: Chi.
	m, _ = answer(m, keyRunes("3"))
	// Config: env.
	m, _ = answer(m, keyRunes("2"))
	// Log format: JSON.
	m, _ = answer(m, keyRunes("2"))
	// Migrations: no.
	m, _ = answer(m, keyRunes("n"))
	// Request logging: yes.
	m, _ = answer(m, keyRunes("y"))
	// Basic auth: no.
	m, _ = answer(m, keyRunes("n"))
	// Git integration: yes.
	m, _ = answer(m, keyRunes("y"))
	// Secure cookies: no.
	m, _ = answer(m, keyRunes("n"))
	// Automatic HTTPS: no.
	m, _ = answer(m, keyRunes("n"))

	// The next question must be Emails, NOT Custom error pages.
	if q := m.current(); q.title != "Emails" {
		t.Fatalf("want the Emails question after Automatic HTTPS for an API, got %q", q.title)
	}

	// Emails: yes.
	m, _ = answer(m, keyRunes("y"))
	// Error notifications: no (asked because emails is now enabled).
	if q := m.current(); q.title == "Error notifications" {
		m, _ = answer(m, keyRunes("n"))
	}
	// Tests: no.
	m, _ = answer(m, keyRunes("n"))
	// Accounts: JWT (default none, pick 2).
	m, _ = answer(m, keyRunes("2"))
	// Target directory.
	for _, r := range "./hello" {
		m, _ = answer(m, keyRunes(string(r)))
	}
	m, _ = answer(m, key(tea.KeyEnter))

	// Review: generate.
	m, _ = answer(m, key(tea.KeyEnter))
	if !m.generating {
		t.Fatalf("expected the wizard to start generating after the review; review error: %v, on question %q", m.reviewError, m.current().title)
	}

	// Deliver the generation result, as the tea.Cmd would.
	model, _ := m.Update(taskDoneMsg{result: generator.Result{Dir: "./hello"}})
	m = model.(wizardModel)

	if generated {
		t.Error("the injected generate func must not touch disk in tests")
	}
	if m.outcome == outcomeFailed {
		t.Fatalf("generation failed: %v", m.genErr)
	}
	if m.rsp.CustomErrorPages {
		t.Error("custom error pages must not be enabled for an API")
	}
	if m.rsp.AppType != "api" || m.rsp.Database != "sqlite" || m.rsp.Router != "chi" {
		t.Errorf("unexpected spec: %+v", m.rsp.ProjectSpec)
	}
}

// TestBackNavigationResetsHiddenAnswers ensures that changing an earlier
// answer resets the answers of questions it makes invisible.
func TestBackNavigationResetsHiddenAnswers(t *testing.T) {
	m := newWizardModel(nil)

	// Walk to the custom error pages question as a web app and answer yes.
	m.idx = indexOfQuestion(m.qs, "Custom error pages")
	*m.current().boolean = true

	// Go back to the application type and switch to an API; the next
	// commit must reset the now-invisible custom error pages answer.
	m.idx = 0
	m.v.appType = "api"
	m = assertWizard(m.Update(keyRunes("2"))) // re-answer: api
	m = assertWizard(m.Update(key(tea.KeyEnter)))

	if m.v.customErrorPages {
		t.Fatal("custom error pages answer survived an app-type change to api")
	}
}

// TestEnforceDependencies keeps the review consistent with earlier
// answers: user accounts turn on their dependencies.
func TestEnforceDependencies(t *testing.T) {
	v := newWizardValues()
	v.accounts = "full"
	enforceDependencies(&v)
	if !v.sessions || !v.emails {
		t.Error("full accounts must enable sessions and emails")
	}

	v = newWizardValues()
	v.accounts = "jwt"
	enforceDependencies(&v)
	if !v.emails {
		t.Error("jwt accounts must enable emails")
	}

	v = newWizardValues()
	v.errorNotifications = true
	enforceDependencies(&v)
	if !v.emails {
		t.Error("error notifications must enable emails")
	}
}

// TestApplyRejectsContradictions is the last line of defense: even if a
// contradictory answer sneaks through, apply refuses it.
func TestApplyRejectsContradictions(t *testing.T) {
	v := newWizardValues()
	v.appType = "api"
	v.modulePath = "github.com/alice/demo"
	v.customErrorPages = true

	if _, err := apply(&v); err == nil {
		t.Fatal("apply must reject custom error pages for an API")
	}
}

// indexOfQuestion returns the index of the named question.
func indexOfQuestion(qs []question, title string) int {
	for i, q := range qs {
		if q.title == title {
			return i
		}
	}
	panic("question not found: " + title)
}
