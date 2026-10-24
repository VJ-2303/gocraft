package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/vj/nomoreboiler/internal/generator"
	"github.com/vj/nomoreboiler/internal/spec"
)

// questionKind is the interaction style of a question.
type questionKind int

const (
	kindSelect questionKind = iota
	kindBool
	kindText
	kindReview
)

// option is one selectable choice in a select question.
type option struct {
	label string
	value string
}

// question is a single wizard step.
type question struct {
	title string
	desc  string
	kind  questionKind

	// options returns the choices for select questions; it may depend
	// on earlier answers.
	options func(v *wizardValues) []option

	// str and boolean bind the answer to the wizard values.
	str     *string
	boolean *bool

	// validate checks a text answer before it is accepted.
	validate func(string) error

	// visible reports whether the question applies to the current
	// answers; invisible questions are skipped and their answers reset.
	visible func(v *wizardValues) bool

	// placeholder is shown for empty text answers.
	placeholder string
}

// wizardValues holds every answer across the wizard.
type wizardValues struct {
	appType      string
	modulePath   string
	database     string
	router       string
	configSource string
	logFormat    string
	migrations   bool
	sessions     bool

	requestLogging bool
	basicAuth      bool
	gitIntegration bool
	secureCookies  bool

	autoHTTPS          bool
	customErrorPages   bool
	emails             bool
	errorNotifications bool
	tests              bool
	accounts           string

	targetDir string
}

// newWizardValues returns the default wizard answers.
func newWizardValues() wizardValues {
	return wizardValues{
		appType:      string(spec.AppWeb),
		database:     string(spec.DBNone),
		router:       string(spec.RouterServeMux),
		configSource: string(spec.ConfigFlags),
		logFormat:    string(spec.LogPlaintext),
		accounts:     "none",
	}
}

// questions builds the wizard's question list. Visibility functions keep
// every question consistent with the answers given so far; the wizard
// resets the answers of questions that become invisible.
func questions(v *wizardValues) []question {
	return []question{
		{
			title: "Application type",
			desc:  "Determines the structure of the codebase.",
			kind:  kindSelect,
			str:   &v.appType,
			options: func(*wizardValues) []option {
				return []option{
					{"Traditional web application — server-side HTML rendering", string(spec.AppWeb)},
					{"JSON API — processes JSON requests and returns JSON responses", string(spec.AppAPI)},
				}
			},
		},
		{
			title:       "Module path",
			desc:        "Used in every import statement.",
			kind:        kindText,
			str:         &v.modulePath,
			placeholder: "github.com/alice/example",
			validate: func(value string) error {
				if value == "" {
					return fmt.Errorf("the module path is required")
				}
				if err := checkModulePath(value); err != nil {
					return fmt.Errorf("not a valid Go module path, e.g. github.com/alice/example")
				}
				return nil
			},
		},
		{
			title: "Database",
			desc:  "The SQL database the codebase connects to.",
			kind:  kindSelect,
			str:   &v.database,
			options: func(*wizardValues) []option {
				return []option{
					{"None", string(spec.DBNone)},
					{"SQLite", string(spec.DBSQLite)},
					{"PostgreSQL", string(spec.DBPostgres)},
					{"MySQL", string(spec.DBMySQL)},
				}
			},
		},
		{
			title: "Router",
			kind:  kindSelect,
			str:   &v.router,
			options: func(*wizardValues) []option {
				return []option{
					{"http.ServeMux (standard library)", string(spec.RouterServeMux)},
					{"Flow", string(spec.RouterFlow)},
					{"Chi", string(spec.RouterChi)},
					{"Gorilla Mux", string(spec.RouterGorilla)},
					{"HttpRouter", string(spec.RouterHttpRouter)},
				}
			},
		},
		{
			title: "Configuration source",
			kind:  kindSelect,
			str:   &v.configSource,
			options: func(*wizardValues) []option {
				return []option{
					{"Command-line flags", string(spec.ConfigFlags)},
					{"Environment variables", string(spec.ConfigEnv)},
				}
			},
		},
		{
			title: "Log format",
			kind:  kindSelect,
			str:   &v.logFormat,
			options: func(*wizardValues) []option {
				return []option{
					{"Plaintext", string(spec.LogPlaintext)},
					{"JSON", string(spec.LogJSON)},
					{"Colorized plaintext", string(spec.LogColorized)},
				}
			},
		},
		{
			title:   "SQL migrations",
			desc:    "Embedded migrations with Makefile targets and auto-apply on startup.",
			kind:    kindBool,
			boolean: &v.migrations,
			visible: func(v *wizardValues) bool { return v.database != string(spec.DBNone) },
		},
		{
			title:   "Server-side sessions",
			desc:    "Stored in your database; the cookie holds only a random ID.",
			kind:    kindBool,
			boolean: &v.sessions,
			visible: func(v *wizardValues) bool {
				return v.appType == string(spec.AppWeb) && v.database != string(spec.DBNone)
			},
		},
		{
			title:   "Request logging",
			desc:    "Logs client IP, method, URL, status code and body size.",
			kind:    kindBool,
			boolean: &v.requestLogging,
		},
		{
			title:   "Basic auth",
			desc:    "HTTP basic auth with a bcrypt-hashed password.",
			kind:    kindBool,
			boolean: &v.basicAuth,
		},
		{
			title:   "Git integration",
			desc:    "A .gitignore and a version derived from Git metadata.",
			kind:    kindBool,
			boolean: &v.gitIntegration,
		},
		{
			title:   "Secure cookies",
			desc:    "Helpers for signed and encrypted cookies.",
			kind:    kindBool,
			boolean: &v.secureCookies,
		},
		{
			title:   "Automatic HTTPS",
			desc:    "Let's Encrypt certificates with HTTP→HTTPS redirects.",
			kind:    kindBool,
			boolean: &v.autoHTTPS,
		},
		{
			title:   "Custom error pages",
			desc:    "Custom 400, 404 and 500 error pages.",
			kind:    kindBool,
			boolean: &v.customErrorPages,
			visible: func(v *wizardValues) bool { return v.appType == string(spec.AppWeb) },
		},
		{
			title:   "Emails",
			desc:    "Email templates and SMTP sending helpers.",
			kind:    kindBool,
			boolean: &v.emails,
		},
		{
			title:   "Error notifications",
			desc:    "Email an admin address when a runtime error occurs.",
			kind:    kindBool,
			boolean: &v.errorNotifications,
			visible: func(v *wizardValues) bool { return v.emails },
		},
		{
			title:   "Tests",
			desc:    "Full test coverage and standardized test helpers.",
			kind:    kindBool,
			boolean: &v.tests,
		},
		{
			title: "User accounts",
			desc:  "Email/password authentication.",
			kind:  kindSelect,
			str:   &v.accounts,
			options: func(v *wizardValues) []option {
				options := []option{{"None", "none"}}
				if v.appType == string(spec.AppWeb) {
					options = append(options, option{"Full — signup, login, logout, password reset, CSRF protection", "full"})
				} else {
					options = append(options, option{"JWT — JSON Web Token authentication", "jwt"})
				}
				return options
			},
		},
		{
			title:       "Target directory",
			desc:        "Must not exist yet, or must be empty.",
			kind:        kindText,
			str:         &v.targetDir,
			placeholder: "./my-project",
			validate: func(value string) error {
				if value == "" {
					return fmt.Errorf("the target directory is required")
				}
				return checkTargetDir(value)
			},
		},
		{
			title: "Review",
			kind:  kindReview,
		},
	}
}

// generateFunc runs generation for an applied spec into the target
// directory.
type generateFunc func(spec.ResolvedSpec, string) (generator.Result, error)

// outcome describes how the wizard ended.
type outcome int

const (
	outcomeDone outcome = iota
	outcomeAborted
	outcomeCanceled
	outcomeFailed
)

// taskDoneMsg carries the outcome of the generation task.
type taskDoneMsg struct {
	result generator.Result
	err    error
}

// wizardModel is the Bubble Tea model driving the whole wizard.
type wizardModel struct {
	v   *wizardValues
	qs  []question
	idx int // index into qs of the active question

	cursor     int // select and bool cursor
	input      string
	inputError error

	generate    generateFunc
	spinner     spinnerModel
	generating  bool
	canceled    bool
	outcome     outcome
	reviewError error

	rsp    spec.ResolvedSpec
	result generator.Result
	genErr error
}

// newWizardModel returns the model at the first question.
func newWizardModel(generate generateFunc) wizardModel {
	v := newWizardValues()
	return wizardModel{
		v:        &v,
		qs:       questions(&v),
		generate: generate,
		outcome:  outcomeDone,
		spinner:  newSpinnerModel(),
	}
}

// current returns the active question.
func (m wizardModel) current() question {
	return m.qs[m.idx]
}

// syncQuestionState resets the per-question interaction state from the
// bound values.
func (m *wizardModel) syncQuestionState() {
	q := m.current()
	m.inputError = nil

	switch q.kind {
	case kindSelect:
		m.cursor = 0
		for i, opt := range q.options(m.v) {
			if opt.value == *q.str {
				m.cursor = i
				break
			}
		}
	case kindBool:
		m.cursor = 0
		if *q.boolean {
			m.cursor = 1
		}
	case kindText:
		m.input = *q.str
	}
}

// resetQuestion restores a question's answer to its default. It is used
// when a question becomes invisible, so stale answers never reach the
// generated codebase.
func resetQuestion(q question) {
	switch q.kind {
	case kindBool:
		*q.boolean = false
	case kindSelect:
		if opts := q.options(nil); len(opts) > 0 {
			*q.str = opts[0].value
		}
	}
}

// enforceDependencies keeps answers consistent at all times: features
// that depend on other features turn those dependencies on, so the
// review never contradicts an earlier answer.
func enforceDependencies(v *wizardValues) {
	if v.accounts == "full" {
		v.sessions = true
		v.emails = true
	}
	if v.accounts == "jwt" {
		v.emails = true
	}
	if v.errorNotifications {
		v.emails = true
	}
}

// nextQuestion moves to the next visible question after index from.
func (m *wizardModel) nextQuestion(from int) {
	for i := from + 1; i < len(m.qs); i++ {
		q := m.qs[i]
		if q.visible == nil || q.visible(m.v) {
			m.idx = i
			m.syncQuestionState()
			return
		}
	}

	// No further question: stay on the review.
	m.idx = len(m.qs) - 1
	m.syncQuestionState()
}

// resetInvisible restores the default answers of every question the
// current answers make invisible, so stale values never reach the
// generated codebase — whichever way the user navigated.
func (m *wizardModel) resetInvisible() {
	for i, q := range m.qs {
		if i == m.idx {
			continue
		}
		if q.visible != nil && !q.visible(m.v) {
			resetQuestion(q)
		}
	}
}

// previousQuestion moves to the closest visible question before index
// from.
func (m *wizardModel) previousQuestion(from int) {
	for i := from - 1; i >= 0; i-- {
		q := m.qs[i]
		if q.visible == nil || q.visible(m.v) {
			m.idx = i
			m.syncQuestionState()
			return
		}
	}
}

// next commits the active question and advances. Dependencies are
// enforced first, then every question the new answers make invisible is
// reset, so the review can never contradict an earlier answer.
func (m *wizardModel) next() tea.Cmd {
	enforceDependencies(m.v)
	m.resetInvisible()
	m.nextQuestion(m.idx)
	return nil
}

// back returns to the previous visible question, or aborts at the first.
func (m *wizardModel) back() tea.Cmd {
	if m.idx == 0 {
		m.outcome = outcomeAborted
		return tea.Quit
	}
	m.previousQuestion(m.idx)
	return nil
}

// progress returns the 1-based position of the active question and the
// total number of visible questions, computed from the current answers.
func (m wizardModel) progress() (int, int) {
	total, position := 0, 0
	for i, q := range m.qs {
		if q.visible != nil && !q.visible(m.v) {
			continue
		}
		total++
		if i <= m.idx {
			position++
		}
	}
	return position, total
}

// Init prepares the model.
func (m wizardModel) Init() tea.Cmd {
	return nil
}

// Update handles key input for the active question.
func (m wizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			if m.generating {
				m.canceled = true
				m.outcome = outcomeCanceled
			} else {
				m.outcome = outcomeAborted
			}
			return m, tea.Quit
		}

		if m.generating {
			return m, nil
		}

		if msg.String() == "esc" {
			cmd := m.back()
			return m, cmd
		}

		switch m.current().kind {
		case kindSelect:
			return m.updateSelect(msg)
		case kindBool:
			return m.updateBool(msg)
		case kindText:
			return m.updateText(msg)
		case kindReview:
			return m.updateReview(msg)
		}

	case taskDoneMsg:
		m.result = msg.result
		m.genErr = msg.err
		if msg.err != nil {
			m.outcome = outcomeFailed
		} else {
			m.outcome = outcomeDone
		}
		return m, tea.Quit
	}

	if m.generating {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

// updateSelect handles a select question.
func (m wizardModel) updateSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	q := m.current()
	opts := q.options(m.v)

	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(opts)-1 {
			m.cursor++
		}
	case "enter":
		*q.str = opts[m.cursor].value
		return m, m.next()
	default:
		if n, ok := keyNumber(msg.String()); ok && n >= 1 && n <= len(opts) {
			m.cursor = n - 1
			*q.str = opts[m.cursor].value
			return m, m.next()
		}
	}
	return m, nil
}

// updateBool handles a boolean question: y or n answers instantly, or
// arrows and enter commit the cursor.
func (m wizardModel) updateBool(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		*m.current().boolean = true
		return m, m.next()
	case "n", "N":
		*m.current().boolean = false
		return m, m.next()
	case "up", "k", "left", "h":
		m.cursor = 0
	case "down", "j", "right", "l":
		m.cursor = 1
	case "enter":
		*m.current().boolean = m.cursor == 1
		return m, m.next()
	case "1":
		*m.current().boolean = false
		return m, m.next()
	case "2":
		*m.current().boolean = true
		return m, m.next()
	}
	return m, nil
}

// updateText handles a text question: typing appends, enter validates
// and advances. Invalid answers keep the user on the question with an
// inline error.
func (m wizardModel) updateText(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	q := m.current()

	switch msg.Type {
	case tea.KeyEnter:
		if err := validateTextAnswer(q, m.input); err != nil {
			m.inputError = err
			return m, nil
		}
		*q.str = m.input
		m.inputError = nil
		return m, m.next()
	case tea.KeyBackspace, tea.KeyCtrlH:
		if len(m.input) > 0 {
			runes := []rune(m.input)
			m.input = string(runes[:len(runes)-1])
		}
		m.inputError = nil
		return m, nil
	case tea.KeyCtrlU:
		m.input = ""
		m.inputError = nil
		return m, nil
	case tea.KeySpace:
		m.input += " "
		m.inputError = nil
		return m, nil
	case tea.KeyRunes:
		m.input += string(msg.Runes)
		m.inputError = nil
		return m, nil
	}
	return m, nil
}

// validateTextAnswer runs a text question's validation.
func validateTextAnswer(q question, value string) error {
	if q.validate != nil {
		return q.validate(value)
	}
	return nil
}

// updateReview handles the review: enter applies and generates. Apply
// errors stay on the review screen as an inline message instead of
// leaving the wizard.
func (m wizardModel) updateReview(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type != tea.KeyEnter {
		return m, nil
	}

	s, err := apply(m.v)
	if err != nil {
		m.reviewError = err
		return m, nil
	}
	m.reviewError = nil
	m.rsp = s
	m.generating = true

	generate := m.generate
	rsp := s
	target := m.v.targetDir
	return m, tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			result, err := generate(rsp, target)
			return taskDoneMsg{result: result, err: err}
		},
	)
}

// keyNumber reports whether the key is a plain digit and returns its
// numeric value.
func keyNumber(key string) (int, bool) {
	if len(key) == 1 && key[0] >= '1' && key[0] <= '9' {
		return int(key[0] - '0'), true
	}
	return 0, false
}

// View renders the active question and the progress line.
func (m wizardModel) View() string {
	position, total := m.progress()

	if m.generating {
		return "  " + progress(position, total) + "\n\n" +
			"  " + m.spinner.View() + " Generating codebase…  " +
			hintStyle.Render("(ctrl+c to cancel)") + "\n"
	}

	q := m.current()

	var b strings.Builder
	b.WriteString("  " + questionLine(q.title, q.desc) + "\n")
	b.WriteString("  " + progress(position, total) + "\n\n")

	switch q.kind {
	case kindSelect:
		opts := q.options(m.v)
		for i, opt := range opts {
			cursor, mark := "  ", fmt.Sprintf("%d", i+1)
			label := opt.label
			if i == m.cursor {
				cursor = selectorMark + " "
				mark = selectedValueStyle.Render(fmt.Sprintf("%d", i+1))
				label = selectedValueStyle.Render(opt.label)
			}
			b.WriteString(fmt.Sprintf("  %s %s  %s\n", cursor, mark, label))
		}
		b.WriteString("\n")
		b.WriteString("  " + help("↑/↓ move", "1-"+fmt.Sprint(len(opts))+" select", "enter confirm", "esc back", "ctrl+c quit"))
	case kindBool:
		options := []string{"No", "Yes"}
		for i, label := range options {
			cursor, mark := "  ", fmt.Sprintf("%d", i+1)
			labelView := label
			if i == m.cursor {
				cursor = selectorMark + " "
				mark = selectedValueStyle.Render(fmt.Sprintf("%d", i+1))
				labelView = selectedValueStyle.Render(label)
			}
			b.WriteString(fmt.Sprintf("  %s %s  %s\n", cursor, mark, labelView))
		}
		b.WriteString("\n")
		b.WriteString("  " + help("y/n quick answer", "↑/↓ move", "enter confirm", "esc back"))
	case kindText:
		b.WriteString("  " + m.input + cursorStyle.Render("▏"))
		if m.input == "" && q.placeholder != "" {
			b.WriteString(hintStyle.Render(q.placeholder + "  ⏎"))
		}
		b.WriteString("\n")
		if m.inputError != nil {
			b.WriteString("\n  " + errorStyle.Render("✗ "+m.inputError.Error()) + "\n")
		}
		b.WriteString("\n")
		b.WriteString("  " + help("enter confirm", "esc back", "ctrl+c quit"))
	case kindReview:
		b.WriteString(reviewSummary(m.v))
		if m.reviewError != nil {
			b.WriteString("\n  " + errorStyle.Render("✗ "+m.reviewError.Error()) + "\n")
			b.WriteString("  " + hintStyle.Render("Go back with esc, answer the conflicting questions, and return."))
			b.WriteString("\n")
		}
		b.WriteString("\n  " + help("enter generate", "esc back", "ctrl+c quit"))
	}

	return b.String()
}

// reviewSummary renders the review table body.
func reviewSummary(v *wizardValues) string {
	row := func(label, value string) string {
		return fmt.Sprintf("  %-20s %s\n", label, value)
	}
	rowBool := func(label string, b bool) string {
		return row(label, yesNo(b))
	}

	accounts := "None"
	switch v.accounts {
	case "full":
		accounts = "Full (web)"
	case "jwt":
		accounts = "JWT (API)"
	}

	var out strings.Builder
	out.WriteString(row("Application type", v.appType))
	out.WriteString(row("Module path", v.modulePath))
	out.WriteString(row("Database", v.database))
	out.WriteString(rowBool("SQL migrations", v.migrations))
	out.WriteString(row("Router", v.router))
	out.WriteString(row("Configuration", v.configSource))
	out.WriteString(row("Log format", v.logFormat))
	out.WriteString(rowBool("Request logging", v.requestLogging))
	out.WriteString(rowBool("Basic auth", v.basicAuth))
	out.WriteString(rowBool("Git integration", v.gitIntegration))
	out.WriteString(rowBool("Secure cookies", v.secureCookies))
	out.WriteString(rowBool("Sessions", v.sessions))
	out.WriteString(rowBool("Automatic HTTPS", v.autoHTTPS))
	out.WriteString(rowBool("Custom error pages", v.customErrorPages))
	out.WriteString(rowBool("Emails", v.emails))
	out.WriteString(rowBool("Error notifications", v.errorNotifications))
	out.WriteString(rowBool("Tests", v.tests))
	out.WriteString(row("User accounts", accounts))
	out.WriteString(row("Target directory", v.targetDir))
	return out.String()
}
