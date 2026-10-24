package ui

import (
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/VJ-2303/gocraft/internal/spec"
	tea "github.com/charmbracelet/bubbletea"
)

func keyRunes(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func key(t tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: t}
}

func answer(m *model, k tea.Msg) (*model, tea.Cmd) {
	updated, cmd := m.Update(k)
	return updated.(*model), cmd
}

func TestWizardAPIFlowDoesNotAskWebQuestions(t *testing.T) {
	var generated bool
	mockGen := func(rsp spec.ResolvedSpec, target string, logger *slog.Logger) (int, error) {
		generated = true
		return 35, nil
	}

	m := newModel(mockGen, nil)

	// Step 1: Application Type: pick 2 (API)
	m, _ = answer(m, keyRunes("2"))
	if m.values.appType != "api" {
		t.Fatalf("want api app type, got %q", m.values.appType)
	}

	// Step 2: Module Path
	for _, r := range "github.com/vj2303/testapi" {
		m, _ = answer(m, keyRunes(string(r)))
	}
	m, _ = answer(m, key(tea.KeyEnter))

	// Step 3: Database: pick 2 (SQLite)
	m, _ = answer(m, keyRunes("2"))
	// Step 4: Router: pick 3 (Chi)
	m, _ = answer(m, keyRunes("3"))
	// Step 5: Config: pick 2 (Env)
	m, _ = answer(m, keyRunes("2"))
	// Step 6: Log: pick 2 (JSON)
	m, _ = answer(m, keyRunes("2"))
	// Step 7: Migrations: n
	m, _ = answer(m, keyRunes("n"))
	// Step 8: Request logging: y
	m, _ = answer(m, keyRunes("y"))
	// Step 9: Basic auth: n
	m, _ = answer(m, keyRunes("n"))
	// Step 10: Git integration: y
	m, _ = answer(m, keyRunes("y"))
	// Step 11: Secure cookies: n
	m, _ = answer(m, keyRunes("n"))
	// Step 12: Auto HTTPS: n
	m, _ = answer(m, keyRunes("n"))

	// Next question must be Emails, NOT Custom error pages or Sessions
	if curr := m.current(); curr.id != "emails" {
		t.Fatalf("want the Emails question after Auto HTTPS for an API, got %q (%s)", curr.id, curr.title)
	}

	// Emails: y
	m, _ = answer(m, keyRunes("y"))
	// Error notifications: n
	if curr := m.current(); curr.id == "error_notifications" {
		m, _ = answer(m, keyRunes("n"))
	}
	// Tests: y
	m, _ = answer(m, keyRunes("y"))
	// Accounts: pick 2 (JWT)
	m, _ = answer(m, keyRunes("2"))
	if m.values.accounts != "jwt" {
		t.Fatalf("want jwt accounts, got %q", m.values.accounts)
	}

	// Target directory
	for _, r := range "./test-api" {
		m, _ = answer(m, keyRunes(string(r)))
	}
	m, _ = answer(m, key(tea.KeyEnter))

	// Should now be on Review
	if m.stepIdx != m.reviewIdx() {
		t.Fatalf("expected to be at review, got step %d (%s)", m.stepIdx, m.current().id)
	}

	// Press Enter on Review to trigger generation
	m, _ = answer(m, key(tea.KeyEnter))
	if m.state != stateGenerating {
		t.Fatalf("expected stateGenerating, got %v with errText: %q", m.state, m.errText)
	}

	// Deliver generation completion message
	m, _ = answer(m, genDoneMsg{count: 35, err: nil})
	if m.state != stateDone {
		t.Fatalf("expected stateDone, got %v", m.state)
	}

	if m.resolvedSpec.CustomErrorPages {
		t.Error("custom error pages must not be enabled for an API")
	}
	if m.resolvedSpec.Sessions {
		t.Error("sessions must not be enabled for an API")
	}
	if m.resolvedSpec.AppType != "api" || m.resolvedSpec.Database != "sqlite" || m.resolvedSpec.Router != "chi" {
		t.Errorf("unexpected resolved spec: %+v", m.resolvedSpec.ProjectSpec)
	}
	_ = generated
}

func TestBackNavigationResetsHiddenAnswers(t *testing.T) {
	m := newModel(nil, nil)

	// Set web-specific answers directly
	m.values.customErrorPages = true
	m.values.sessions = true
	m.values.accounts = "full"

	// Switch app type to API
	m.values.appType = "api"
	m.resetInvisible()

	if m.values.customErrorPages {
		t.Error("customErrorPages must be reset when app_type is API")
	}
	if m.values.sessions {
		t.Error("sessions must be reset when app_type is API")
	}
	if m.values.accounts != "none" {
		t.Errorf("accounts must be reset when switching to API, got %q", m.values.accounts)
	}
}

func TestEnforceDependenciesAndCascade(t *testing.T) {
	m := newModel(nil, nil)

	// Test badge detection
	badge := m.values.detectCascadeBadges("accounts", "full")
	if !strings.Contains(badge, "SQLite") || !strings.Contains(badge, "Sessions") || !strings.Contains(badge, "Emails") {
		t.Errorf("expected full accounts badge to mention SQLite, Sessions, Emails, got %q", badge)
	}

	// Apply cascade
	m.values.applyCascade("accounts", "full")
	if !m.values.sessions || !m.values.emails || m.values.database != "sqlite" {
		t.Errorf("expected sessions, emails, and sqlite to be auto-enabled, got db=%s, sess=%t, emails=%t",
			m.values.database, m.values.sessions, m.values.emails)
	}

	// Reset db to none and verify sessions auto-enables SQLite badge
	m.values.database = "none"
	sessBadge := m.values.detectCascadeBadges("sessions", "true")
	if !strings.Contains(sessBadge, "SQLite") {
		t.Errorf("expected sessions badge to mention SQLite, got %q", sessBadge)
	}
	m.values.applyCascade("sessions", true)
	if m.values.database != "sqlite" {
		t.Errorf("expected database to become sqlite, got %q", m.values.database)
	}
}

func TestDestructiveConflictPrompt(t *testing.T) {
	m := newModel(nil, nil)

	// Step 3 is database
	m.stepIdx = 2
	m.values.database = "sqlite"
	m.values.migrations = true
	m.values.sessions = true
	m.syncInputs()

	// Select option 1 (None)
	m, _ = answer(m, keyRunes("1"))

	// Must transition to stateConflictConfirm
	if m.state != stateConflictConfirm {
		t.Fatalf("expected stateConflictConfirm, got %v", m.state)
	}
	if !strings.Contains(m.conflictMsg, "SQL Migrations") {
		t.Errorf("expected conflict message to mention SQL Migrations, got %q", m.conflictMsg)
	}

	// Cancel with 'n'
	m, _ = answer(m, keyRunes("n"))
	if m.state != stateQuestion {
		t.Fatalf("expected stateQuestion after cancel, got %v", m.state)
	}
	if m.values.database != "sqlite" || !m.values.migrations {
		t.Errorf("expected database to remain sqlite with migrations, got db=%s, migrations=%t",
			m.values.database, m.values.migrations)
	}

	// Try again and confirm with 'y'
	m, _ = answer(m, keyRunes("1"))
	if m.state != stateConflictConfirm {
		t.Fatalf("expected stateConflictConfirm, got %v", m.state)
	}
	m, _ = answer(m, keyRunes("y"))
	if m.state != stateQuestion {
		t.Fatalf("expected stateQuestion after confirm, got %v", m.state)
	}
	if m.values.database != "none" || m.values.migrations || m.values.sessions {
		t.Errorf("expected database none and dependent features disabled, got db=%s, migrations=%t, sessions=%t",
			m.values.database, m.values.migrations, m.values.sessions)
	}
}

func TestReviewDirectJump(t *testing.T) {
	m := newModel(nil, nil)
	m.stepIdx = m.reviewIdx()
	m.syncInputs()

	// Jump to step 1 (App Type)
	m, _ = answer(m, keyRunes("1"))
	if m.stepIdx != 0 {
		t.Fatalf("expected step 0 (app_type), got %d (%s)", m.stepIdx, m.current().id)
	}
	if !m.editingFromReview {
		t.Fatal("expected editingFromReview to be true")
	}

	// Confirming with Enter should return directly to Review
	m, _ = answer(m, key(tea.KeyEnter))
	if m.stepIdx != m.reviewIdx() {
		t.Fatalf("expected to return to review (%d), got %d", m.reviewIdx(), m.stepIdx)
	}
	if m.editingFromReview {
		t.Fatal("expected editingFromReview to be false")
	}

	// Jump to step 2 (Module)
	m, _ = answer(m, keyRunes("2"))
	if m.current().id != "module" {
		t.Fatalf("expected module step, got %s", m.current().id)
	}

	// Esc should cancel edit and return to Review
	m, _ = answer(m, key(tea.KeyEsc))
	if m.stepIdx != m.reviewIdx() {
		t.Fatalf("expected to return to review on Esc, got %d", m.stepIdx)
	}
}

func TestResponsiveRendering(t *testing.T) {
	m := newModel(nil, nil)

	// Wide terminal (split pane)
	m.width = 120
	m.height = 30
	viewWide := m.View()
	if !strings.Contains(viewWide, "ARCHITECTURE BLUEPRINT") {
		t.Error("wide view must contain ARCHITECTURE BLUEPRINT right pane")
	}

	// Narrow terminal (stacked)
	m.width = 80
	m.height = 24
	viewNarrow := m.View()
	if !strings.Contains(viewNarrow, "GOCRAFT") {
		t.Error("narrow view must contain GOCRAFT header")
	}

	// Done view
	m.state = stateDone
	m.genCount = 30
	viewDone := m.View()
	if !strings.Contains(viewDone, "Project Ready") {
		t.Error("done view must contain 'Project Ready'")
	}

	// Failed view
	m.state = stateFailed
	m.genErr = errors.New("mock generation error")
	viewFailed := m.View()
	if !strings.Contains(viewFailed, "GENERATION FAILED") {
		t.Error("failed view must contain 'GENERATION FAILED'")
	}
}
