package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/vj/nomoreboiler/internal/generator"
	"github.com/vj/nomoreboiler/internal/spec"
)

// scenario is a full set of desired answers, keyed by question title.
// Bools are "yes"/"no"; selects are the option value; texts are the
// literal value. Questions not listed answer "no"/default.
type scenario struct {
	name     string
	answers  map[string]string
	appType  string
	database string
	accounts string // "none", "full", "jwt"
}

// drive walks the real wizard model through a scenario exactly as a user
// would, one key message at a time, and returns the model after the
// generation result has been delivered, along with every question title
// that was asked.
func drive(t *testing.T, s scenario) (wizardModel, []string) {
	t.Helper()

	m := newWizardModel("test", func(rsp spec.ResolvedSpec, target string) (generator.Result, error) {
		return generator.Result{Dir: target}, nil
	})
	answers := func(title string) string {
		if a, ok := s.answers[title]; ok {
			return a
		}
		return "no"
	}

	var asked []string
	seen := map[string]bool{}
	const maxSteps = 64

	for step := 0; step < maxSteps; step++ {
		if m.generating {
			break
		}
		q := m.current()
		if !seen[q.title] {
			asked = append(asked, q.title)
			seen[q.title] = true
		}

		answer := answers(q.title)

		var msg tea.KeyMsg
		switch q.kind {
		case kindSelect:
			if answer == "no" {
				// Unlisted: confirm the currently bound value.
				msg = key(tea.KeyEnter)
				break
			}
			opts := q.options(m.v)
			idx := -1
			for i, opt := range opts {
				if opt.value == answer {
					idx = i
					break
				}
			}
			if idx < 0 {
				t.Fatalf("scenario %q: no option %q for question %q (options %v)", s.name, answer, q.title, opts)
			}
			msg = keyRunes(fmtInt(idx + 1))
		case kindBool:
			if answer == "yes" {
				msg = keyRunes("y")
			} else {
				msg = keyRunes("n")
			}
		case kindText:
			value := answer
			if value == "no" || value == "" {
				// Defaults per text question.
				switch q.title {
				case "Module path":
					value = "github.com/alice/demo"
				case "Target directory":
					value = "./demo-output"
				default:
					t.Fatalf("scenario %q: missing text answer for %q", s.name, q.title)
				}
			}
			for _, r := range value {
				msg = keyRunes(string(r))
				model, _ := m.Update(msg)
				m = model.(wizardModel)
			}
			msg = key(tea.KeyEnter)
		case kindReview:
			msg = key(tea.KeyEnter)
		}

		model, _ := m.Update(msg)
		m = model.(wizardModel)

		if m.outcome != outcomeDone {
			t.Fatalf("scenario %q: wizard ended unexpectedly (%v) on %q: %v",
				s.name, m.outcome, q.title, m.genErr)
		}
	}

	// Deliver the generation result.
	if !m.generating {
		t.Fatalf("scenario %q: generation never started", s.name)
	}
	model, _ := m.Update(taskDoneMsg{result: generator.Result{Dir: m.v.targetDir}})
	m = model.(wizardModel)

	if m.outcome != outcomeDone {
		t.Fatalf("scenario %q: generation failed: %v", s.name, m.genErr)
	}

	return m, asked
}

func fmtInt(n int) string {
	digits := "0123456789"
	if n < 10 {
		return string(digits[n])
	}
	return string(digits[n/10]) + string(digits[n%10])
}

// TestConditionalScenarios walks every meaningful conditional path and
// asserts that the wizard never produces an invalid spec and never asks
// a question that contradicts earlier answers.
func TestConditionalScenarios(t *testing.T) {
	scenarios := []scenario{
		{
			name:     "web, no database, everything off",
			appType:  "web",
			database: "none",
			accounts: "none",
		},
		{
			name:     "web, sqlite, everything on, full accounts",
			appType:  "web",
			database: "sqlite",
			accounts: "full",
			answers: map[string]string{
				"SQL migrations":       "yes",
				"Server-side sessions": "yes",
				"Request logging":      "yes",
				"Basic auth":           "yes",
				"Git integration":      "yes",
				"Secure cookies":       "yes",
				"Automatic HTTPS":      "yes",
				"Custom error pages":   "yes",
				"Emails":               "yes",
				"Error notifications":  "yes",
				"Tests":                "yes",
			},
		},
		{
			name:     "api, postgres, everything on, jwt accounts",
			appType:  "api",
			database: "postgres",
			accounts: "jwt",
			answers: map[string]string{
				"SQL migrations":      "yes",
				"Request logging":     "yes",
				"Basic auth":          "yes",
				"Git integration":     "yes",
				"Secure cookies":      "yes",
				"Automatic HTTPS":     "yes",
				"Emails":              "yes",
				"Error notifications": "yes",
				"Tests":               "yes",
			},
		},
		{
			name:     "web, mysql, emails without notifications, no accounts",
			appType:  "web",
			database: "mysql",
			accounts: "none",
			answers: map[string]string{
				"SQL migrations":       "yes",
				"Server-side sessions": "no",
				"Emails":               "yes",
			},
		},
		{
			name:     "api, no database, jwt accounts attempted",
			appType:  "api",
			database: "none",
			accounts: "jwt",
			answers: map[string]string{
				"Emails": "yes",
			},
		},
		{
			name:     "web, sqlite, sessions declined but full accounts force them on",
			appType:  "web",
			database: "sqlite",
			accounts: "full",
			answers: map[string]string{
				"SQL migrations":       "no",
				"Server-side sessions": "no",
			},
		},
	}

	for _, s := range scenarios {
		s := s
		t.Run(s.name, func(t *testing.T) {
			m, asked := drive(t, s)
			rsp := m.rsp

			// The applied spec must always be valid.
			if err := rsp.Validate(); err != nil {
				t.Errorf("applied spec is invalid: %v", err)
			}

			// Web-only features must never be enabled for an API.
			if rsp.AppType == spec.AppAPI {
				if rsp.CustomErrorPages {
					t.Error("api spec enabled custom error pages")
				}
				if rsp.Sessions {
					t.Error("api spec enabled sessions")
				}
			}

			// Custom error pages must never be asked for an API.
			if rsp.AppType == spec.AppAPI {
				for _, title := range asked {
					if title == "Custom error pages" {
						t.Error("custom error pages question was asked for an API")
					}
				}
			}

			// Sessions must never be asked without a database.
			if rsp.Database == spec.DBNone {
				for _, title := range asked {
					if title == "Server-side sessions" || title == "SQL migrations" {
						t.Errorf("%q question was asked without a database", title)
					}
				}
			}

			// Error notifications must never be asked without emails.
			if !rsp.Emails {
				for _, title := range asked {
					if title == "Error notifications" {
						t.Error("error notifications question was asked without emails")
					}
				}
			}

			// Account dependencies: full accounts force sessions and
			// emails on; jwt accounts force emails on.
			if m.v.accounts == "full" {
				if !rsp.Sessions || !rsp.Emails {
					t.Error("full accounts must enable sessions and emails")
				}
			}
			if m.v.accounts == "jwt" && !rsp.Emails {
				t.Error("jwt accounts must enable emails")
			}

			// Error notifications imply emails.
			if rsp.ErrorNotifications && !rsp.Emails {
				t.Error("error notifications enabled without emails")
			}

			// Sessions require a database and a web app.
			if rsp.Sessions && (rsp.Database == spec.DBNone || rsp.AppType != spec.AppWeb) {
				t.Error("sessions enabled without a web app and database")
			}
		})
	}
}

// TestConditionalBackNavigationChangesVisibility verifies that answering
// questions, going back, changing an earlier answer, and moving forward
// again keeps the asked questions and the final spec consistent.
func TestConditionalBackNavigationChangesVisibility(t *testing.T) {
	v := newWizardValues()
	v.modulePath = "github.com/alice/demo"
	v.targetDir = "./demo"

	// Answer as a web app first: the custom error pages answer is yes.
	v.appType = "web"
	v.customErrorPages = true

	// Go back and switch to an API. The next commit must reset the
	// custom error pages answer and never ask for it again.
	v.appType = "api"

	m := newWizardModel("test", nil)
	m.v = &v
	m.qs = questions(&v)

	// Walk forward from the application type to the review using next()
	// semantics; the reset must keep the spec valid.
	enforceDependencies(m.v)
	m.resetInvisible()

	if v.customErrorPages {
		t.Fatal("custom error pages survived the switch to an API")
	}

	rsp, err := apply(m.v)
	if err != nil {
		t.Fatalf("apply after visibility change: %v", err)
	}
	if rsp.CustomErrorPages {
		t.Error("applied spec enabled custom error pages for an API")
	}
}

// TestCancelDuringGeneration exercises ctrl+c while generating.
func TestCancelDuringGeneration(t *testing.T) {
	m := newWizardModel("test", func(rsp spec.ResolvedSpec, target string) (generator.Result, error) {
		return generator.Result{Dir: target}, nil
	})

	// Force the generating state.
	m.generating = true
	m.spinner = newSpinnerModel()

	model, _ := m.Update(key(tea.KeyCtrlC))
	m = model.(wizardModel)

	if m.outcome != outcomeCanceled {
		t.Errorf("want outcomeCanceled, got %v", m.outcome)
	}
}
