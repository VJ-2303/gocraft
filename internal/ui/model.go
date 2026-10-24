package ui

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/VJ-2303/gocraft/internal/generator"
	"github.com/VJ-2303/gocraft/internal/spec"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type uiState int

const (
	stateQuestion uiState = iota
	stateConflictConfirm
	stateGenerating
	stateDone
	stateFailed
)

type generateFunc func(rsp spec.ResolvedSpec, target string, logger *slog.Logger) (int, error)

type genDoneMsg struct {
	count int
	err   error
}

type model struct {
	state     uiState
	values    *wizardValues
	questions []question
	stepIdx   int
	cursor    int

	textInput textinput.Model
	spinner   spinner.Model

	width  int
	height int

	// Conflict confirmation state
	conflictMsg          string
	conflictPendingApply func()

	// Review jump state
	editingFromReview bool

	// Inline error & feedback
	errText    string
	autoNotice string

	// Generation
	generateFn   generateFunc
	resolvedSpec spec.ResolvedSpec
	genCount     int
	genErr       error
	logger       *slog.Logger
}

func defaultGenerate(rsp spec.ResolvedSpec, target string, logger *slog.Logger) (int, error) {
	res, err := generator.Generate(rsp, target, logger, generator.Options{Tidy: true})
	if err != nil {
		return 0, err
	}
	return len(res.Files), nil
}

func newModel(genFn generateFunc, logger *slog.Logger) *model {
	if genFn == nil {
		genFn = defaultGenerate
	}
	if logger == nil {
		logger = slog.Default()
	}

	ti := textinput.New()
	ti.Prompt = ""
	ti.TextStyle = styleInput
	ti.PlaceholderStyle = styleDesc
	ti.Cursor.Style = styleCursor
	ti.Focus()

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = styleCursor

	v := defaultWizardValues()
	qs := getQuestions(v)

	m := &model{
		state:      stateQuestion,
		values:     v,
		questions:  qs,
		stepIdx:    0,
		cursor:     0,
		textInput:  ti,
		spinner:    sp,
		generateFn: genFn,
		logger:     logger,
		width:      100,
		height:     28,
	}

	m.syncInputs()
	return m
}

func (m *model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *model) current() question {
	return m.questions[m.stepIdx]
}

func (m *model) reviewIdx() int {
	return len(m.questions) - 1
}

func (m *model) nextStep(from int) int {
	for i := from + 1; i < len(m.questions); i++ {
		q := m.questions[i]
		if q.visible == nil || q.visible(m.values) {
			return i
		}
	}
	return m.reviewIdx()
}

func (m *model) prevStep(from int) int {
	for i := from - 1; i >= 0; i-- {
		q := m.questions[i]
		if q.visible == nil || q.visible(m.values) {
			return i
		}
	}
	return 0
}

func (m *model) syncInputs() {
	m.errText = ""
	m.autoNotice = ""
	if m.stepIdx >= len(m.questions) {
		return
	}
	q := m.questions[m.stepIdx]
	switch q.kind {
	case kindSelect:
		opts := q.options(m.values)
		m.cursor = 0
		if q.str != nil {
			for i, o := range opts {
				if o.value == *q.str {
					m.cursor = i
					break
				}
			}
		}
	case kindBool:
		m.cursor = 1
		if q.boolean != nil && *q.boolean {
			m.cursor = 0
		}
	case kindText:
		if q.str != nil {
			m.textInput.SetValue(*q.str)
		} else {
			m.textInput.SetValue("")
		}
		m.textInput.Placeholder = q.placeholder
		m.textInput.Focus()
	}
}

func (m *model) resetInvisible() {
	if m.values.appType == string(spec.AppAPI) {
		m.values.sessions = false
		m.values.customErrorPages = false
		if m.values.accounts == "full" {
			m.values.accounts = "none"
		}
	}
	if m.values.database == string(spec.DBNone) {
		m.values.migrations = false
		m.values.sessions = false
		if m.values.accounts != "none" {
			m.values.accounts = "none"
		}
	}
	if !m.values.emails {
		m.values.errorNotifs = false
	}
}

func (m *model) jumpToKey(key string) {
	var targetID string
	switch key {
	case "1":
		targetID = "app_type"
	case "2":
		targetID = "module"
	case "3":
		targetID = "database"
	case "4":
		targetID = "router"
	case "5":
		targetID = "config"
	case "6":
		if m.values.appType == string(spec.AppWeb) {
			targetID = "sessions"
		} else {
			targetID = "basic_auth"
		}
	case "7":
		targetID = "accounts"
	case "8":
		targetID = "emails"
	case "9":
		targetID = "target"
	default:
		return
	}

	for i, q := range m.questions {
		if q.id == targetID {
			if q.visible == nil || q.visible(m.values) {
				m.stepIdx = i
				m.editingFromReview = true
				m.syncInputs()
				return
			}
		}
	}
}

func (m *model) commitCurrentAndAdvance() tea.Cmd {
	m.resetInvisible()
	if m.editingFromReview {
		m.stepIdx = m.reviewIdx()
		m.editingFromReview = false
		m.syncInputs()
		return nil
	}
	m.stepIdx = m.nextStep(m.stepIdx)
	m.syncInputs()
	return nil
}

func (m *model) triggerGeneration() tea.Cmd {
	s := m.values.toSpec()
	if err := s.Validate(); err != nil {
		m.errText = err.Error()
		return nil
	}
	rsp, err := spec.Resolve(s)
	if err != nil {
		m.errText = err.Error()
		return nil
	}
	for _, q := range m.questions {
		if q.id == "target" && q.validate != nil {
			if err := q.validate(m.values.targetDir); err != nil {
				m.errText = err.Error()
				return nil
			}
		}
	}

	m.resolvedSpec = rsp
	m.state = stateGenerating
	return tea.Batch(m.spinner.Tick, func() tea.Msg {
		count, genErr := m.generateFn(rsp, m.values.targetDir, m.logger)
		return genDoneMsg{count: count, err: genErr}
	})
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case spinner.TickMsg:
		if m.state == stateGenerating {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil

	case genDoneMsg:
		if msg.err != nil {
			m.state = stateFailed
			m.genErr = msg.err
			return m, nil
		}
		m.state = stateDone
		m.genCount = msg.count
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}

		switch m.state {
		case stateDone, stateFailed:
			switch msg.String() {
			case "q", "Q", "enter", "esc":
				return m, tea.Quit
			}
			return m, nil

		case stateGenerating:
			return m, nil

		case stateConflictConfirm:
			switch msg.String() {
			case "y", "Y", "enter":
				if m.conflictPendingApply != nil {
					m.conflictPendingApply()
				}
				m.conflictMsg = ""
				m.conflictPendingApply = nil
				m.state = stateQuestion
				return m, m.commitCurrentAndAdvance()
			case "n", "N", "esc":
				m.conflictMsg = ""
				m.conflictPendingApply = nil
				m.state = stateQuestion
				m.syncInputs()
				return m, nil
			}
			return m, nil

		case stateQuestion:
			return m.updateQuestion(msg)
		}
	}

	return m, nil
}

func (m *model) updateQuestion(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	q := m.questions[m.stepIdx]

	if msg.Type == tea.KeyEsc {
		if m.editingFromReview {
			m.stepIdx = m.reviewIdx()
			m.editingFromReview = false
			m.syncInputs()
			return m, nil
		}
		prev := m.prevStep(m.stepIdx)
		if prev != m.stepIdx {
			m.stepIdx = prev
			m.syncInputs()
		}
		return m, nil
	}

	switch q.kind {
	case kindReview:
		switch msg.Type {
		case tea.KeyEnter:
			return m, m.triggerGeneration()
		case tea.KeyRunes:
			ch := msg.String()
			if len(ch) == 1 && ch >= "1" && ch <= "9" {
				m.jumpToKey(ch)
				return m, nil
			}
		}

	case kindSelect:
		opts := q.options(m.values)
		switch msg.Type {
		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
			} else {
				m.cursor = len(opts) - 1
			}
			return m, nil
		case tea.KeyDown:
			if m.cursor < len(opts)-1 {
				m.cursor++
			} else {
				m.cursor = 0
			}
			return m, nil
		case tea.KeyRunes:
			ch := msg.String()
			if ch == "k" {
				if m.cursor > 0 {
					m.cursor--
				} else {
					m.cursor = len(opts) - 1
				}
				return m, nil
			}
			if ch == "j" {
				if m.cursor < len(opts)-1 {
					m.cursor++
				} else {
					m.cursor = 0
				}
				return m, nil
			}
			if len(ch) == 1 && ch[0] >= '1' && ch[0] <= '9' {
				idx := int(ch[0] - '1')
				if idx < len(opts) {
					m.cursor = idx
					return m.selectOption(q, opts[idx].value)
				}
			}
		case tea.KeyEnter:
			if m.cursor < len(opts) {
				return m.selectOption(q, opts[m.cursor].value)
			}
		}

	case kindBool:
		switch msg.Type {
		case tea.KeyUp, tea.KeyDown, tea.KeyLeft, tea.KeyRight, tea.KeyTab:
			if m.cursor == 0 {
				m.cursor = 1
			} else {
				m.cursor = 0
			}
			return m, nil
		case tea.KeyRunes:
			ch := strings.ToLower(msg.String())
			if ch == "y" {
				m.cursor = 0
				return m.selectBool(q, true)
			}
			if ch == "n" {
				m.cursor = 1
				return m.selectBool(q, false)
			}
			if ch == "k" || ch == "j" {
				if m.cursor == 0 {
					m.cursor = 1
				} else {
					m.cursor = 0
				}
				return m, nil
			}
		case tea.KeyEnter:
			return m.selectBool(q, m.cursor == 0)
		}

	case kindText:
		switch msg.Type {
		case tea.KeyEnter:
			val := strings.TrimSpace(m.textInput.Value())
			if q.validate != nil {
				if err := q.validate(val); err != nil {
					m.errText = err.Error()
					return m, nil
				}
			}
			if q.str != nil {
				*q.str = val
			}
			m.errText = ""
			return m, m.commitCurrentAndAdvance()
		default:
			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m *model) selectOption(q question, val string) (tea.Model, tea.Cmd) {
	if conflictText, hasConflict := m.values.destructiveConflict(q.id, val); hasConflict {
		m.conflictMsg = conflictText
		m.conflictPendingApply = func() {
			if q.str != nil {
				*q.str = val
			}
			m.values.applyCascade(q.id, val)
			m.resetInvisible()
		}
		m.state = stateConflictConfirm
		return m, nil
	}

	if q.str != nil {
		*q.str = val
	}
	m.values.applyCascade(q.id, val)
	return m, m.commitCurrentAndAdvance()
}

func (m *model) selectBool(q question, val bool) (tea.Model, tea.Cmd) {
	valStr := "false"
	if val {
		valStr = "true"
	}
	if conflictText, hasConflict := m.values.destructiveConflict(q.id, valStr); hasConflict {
		m.conflictMsg = conflictText
		m.conflictPendingApply = func() {
			if q.boolean != nil {
				*q.boolean = val
			}
			m.values.applyCascade(q.id, val)
			m.resetInvisible()
		}
		m.state = stateConflictConfirm
		return m, nil
	}

	if q.boolean != nil {
		*q.boolean = val
	}
	m.values.applyCascade(q.id, val)
	return m, m.commitCurrentAndAdvance()
}

func (m *model) View() string {
	switch m.state {
	case stateDone:
		return m.viewDone()
	case stateFailed:
		return m.viewFailed()
	case stateGenerating:
		return m.viewGenerating()
	}

	if m.width >= 100 {
		leftPane := m.viewLeftPane(m.width - 44)
		rightPane := m.viewRightPane(40)
		return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)
	}

	return m.viewStacked()
}

func (m *model) viewLeftPane(width int) string {
	box := lipgloss.NewStyle().Width(width).Padding(1, 2)
	var content strings.Builder

	// Top Bar
	var stepText string
	if m.stepIdx == m.reviewIdx() {
		stepText = "Review & Launch"
	} else {
		stepText = fmt.Sprintf("Step %d of %d · %s", m.stepIdx+1, len(m.questions), m.current().category)
	}
	topBar := styleCategory.Render("GOCRAFT") + " " + styleDesc.Render("—") + " " + styleDesc.Render(stepText)
	if m.editingFromReview {
		topBar += " " + styleBadge.Render("Editing from Review")
	}
	content.WriteString(topBar + "\n\n")

	// Conflict confirmation view
	if m.state == stateConflictConfirm {
		content.WriteString(m.renderConflictBox(width - 6))
		return box.Render(content.String())
	}

	// Question Title & Desc
	q := m.current()
	content.WriteString(styleTitle.Render(q.title) + "\n")
	if q.desc != "" {
		content.WriteString(styleDesc.Render(q.desc) + "\n\n")
	} else {
		content.WriteString("\n")
	}

	// Body by kind
	switch q.kind {
	case kindReview:
		content.WriteString(m.renderReviewTable())
	case kindSelect:
		opts := q.options(m.values)
		for i, o := range opts {
			badge := m.values.detectCascadeBadges(q.id, o.value)
			var prefix, line string
			if i == m.cursor {
				prefix = styleCursor.Render("❯ ")
				line = styleOptionSelected.Render(fmt.Sprintf("%d. %s", i+1, o.label))
			} else {
				prefix = "  "
				line = styleOptionNormal.Render(fmt.Sprintf("%d. %s", i+1, o.label))
			}
			if badge != "" {
				if strings.HasPrefix(badge, "!") {
					line += "  " + styleConflictBadge.Render(badge)
				} else {
					line += "  " + styleBadge.Render(badge)
				}
			}
			content.WriteString(prefix + line + "\n")
		}
	case kindBool:
		opts := []struct {
			label string
			val   bool
		}{
			{"Yes", true},
			{"No", false},
		}
		for i, o := range opts {
			badge := m.values.detectCascadeBadges(q.id, fmt.Sprintf("%t", o.val))
			var prefix, line string
			if i == m.cursor {
				prefix = styleCursor.Render("❯ ")
				line = styleOptionSelected.Render(o.label)
			} else {
				prefix = "  "
				line = styleOptionNormal.Render(o.label)
			}
			if badge != "" {
				line += "  " + styleBadge.Render(badge)
			}
			content.WriteString(prefix + line + "\n")
		}
	case kindText:
		content.WriteString(styleInputPrompt.Render("❯ ") + m.textInput.View() + "\n")
	}

	// Error or notice line
	content.WriteString("\n")
	if m.errText != "" {
		content.WriteString(styleError.Render("✗ "+m.errText) + "\n")
	} else {
		content.WriteString("\n")
	}

	// Keybindings footer
	content.WriteString("\n")
	if q.kind == kindReview {
		content.WriteString(formatHelp([2]string{"1-9", "jump to step"}, [2]string{"enter", "generate"}, [2]string{"esc", "back"}, [2]string{"ctrl+c", "quit"}))
	} else if q.kind == kindText {
		content.WriteString(formatHelp([2]string{"enter", "confirm"}, [2]string{"esc", "back"}, [2]string{"ctrl+c", "quit"}))
	} else if q.kind == kindBool {
		content.WriteString(formatHelp([2]string{"y/n", "toggle"}, [2]string{"enter", "confirm"}, [2]string{"esc", "back"}, [2]string{"ctrl+c", "quit"}))
	} else {
		content.WriteString(formatHelp([2]string{"1-9", "quick select"}, [2]string{"j/k", "navigate"}, [2]string{"enter", "confirm"}, [2]string{"esc", "back"}, [2]string{"ctrl+c", "quit"}))
	}

	return box.Render(content.String())
}

func (m *model) renderConflictBox(width int) string {
	b := strings.Builder{}
	b.WriteString(styleConflictBadge.Render(" CAUTION: DESTRUCTIVE CHANGE ") + "\n\n")
	b.WriteString(styleOptionSelected.Render(m.conflictMsg) + "\n\n")
	b.WriteString(styleKeyHint.Render("Apply change and reset dependent options? ") + styleKey.Render("[y/N]") + "\n\n")
	b.WriteString(formatHelp([2]string{"y", "confirm"}, [2]string{"n/esc", "cancel"}))
	return b.String()
}

func (m *model) renderReviewTable() string {
	lines := []struct {
		num   string
		label string
		val   string
	}{
		{"1", "Application", string(m.values.appType)},
		{"2", "Module Path", m.values.modulePath},
		{"3", "Database", fmt.Sprintf("%s (migrations: %t)", m.values.database, m.values.migrations)},
		{"4", "Router", string(m.values.router)},
		{"5", "Config & Log", fmt.Sprintf("%s · %s", m.values.configSource, m.values.logFormat)},
		{"6", "Security", fmt.Sprintf("sessions: %t · basic_auth: %t · cookies: %t", m.values.sessions, m.values.basicAuth, m.values.secureCookies)},
		{"7", "Accounts", m.values.accounts},
		{"8", "Services", fmt.Sprintf("emails: %t · notifs: %t · https: %t · tests: %t", m.values.emails, m.values.errorNotifs, m.values.autoHTTPS, m.values.tests)},
		{"9", "Target Dir", m.values.targetDir},
	}

	var sb strings.Builder
	for _, l := range lines {
		keyBadge := styleKey.Render("[" + l.num + "]")
		field := styleDesc.Render(fmt.Sprintf("%-14s", l.label))
		val := styleOptionSelected.Render(l.val)
		sb.WriteString(fmt.Sprintf("%s %s %s\n", keyBadge, field, val))
	}
	return sb.String()
}

func (m *model) viewRightPane(width int) string {
	box := stylePaneBorder.Width(width).Padding(1, 1)
	var sb strings.Builder

	sb.WriteString(styleInspectorTitle.Render("ARCHITECTURE BLUEPRINT") + "\n")
	sb.WriteString(styleDesc.Render(strings.Repeat("─", width-6)) + "\n")

	type row struct {
		k, v string
	}
	rows := []row{
		{"Type", string(m.values.appType)},
		{"Module", m.values.modulePath},
		{"Database", string(m.values.database)},
		{"Router", string(m.values.router)},
		{"Config", string(m.values.configSource)},
		{"Logging", string(m.values.logFormat)},
		{"Accounts", m.values.accounts},
	}

	for _, r := range rows {
		sb.WriteString(fmt.Sprintf("%-10s %s\n", styleInspectorText.Render(r.k), styleInspectorActive.Render(r.v)))
	}

	sb.WriteString("\n" + styleInspectorTitle.Render("DEPENDENCY CONSTRAINTS") + "\n")
	sb.WriteString(styleDesc.Render(strings.Repeat("─", width-6)) + "\n")

	var deps []string
	if m.values.accounts == "full" {
		deps = append(deps, "• Full accounts ↳ SQLite + Sessions + Emails")
	} else if m.values.accounts == "jwt" {
		deps = append(deps, "• JWT accounts ↳ SQLite + Emails")
	}
	if m.values.migrations {
		deps = append(deps, "• SQL migrations ↳ Database required")
	}
	if m.values.sessions {
		deps = append(deps, "• Sessions ↳ Database required")
	}
	if m.values.errorNotifs {
		deps = append(deps, "• Error notifications ↳ Emails required")
	}

	if len(deps) == 0 {
		sb.WriteString(styleInspectorText.Render("No active constraints\n"))
	} else {
		for _, d := range deps {
			sb.WriteString(styleBadge.Render(d) + "\n")
		}
	}

	return box.Render(sb.String())
}

func (m *model) viewStacked() string {
	return m.viewLeftPane(m.width)
}

func (m *model) viewGenerating() string {
	box := lipgloss.NewStyle().Padding(2, 3)
	var sb strings.Builder
	sb.WriteString(styleCategory.Render("GOCRAFT") + "\n\n")
	sb.WriteString(fmt.Sprintf("%s Scaffolding codebase...\n", m.spinner.View()))
	sb.WriteString(styleDesc.Render("Writing template files, configuring dependencies, and resolving Go modules.\n"))
	return box.Render(sb.String())
}

func (m *model) viewDone() string {
	box := lipgloss.NewStyle().Padding(2, 3)
	var sb strings.Builder
	sb.WriteString(styleCategory.Render("GOCRAFT") + " " + styleDesc.Render("—") + " " + styleTitle.Render("Project Ready") + "\n\n")
	sb.WriteString(fmt.Sprintf("Successfully scaffolded %d files into:\n", m.genCount))
	sb.WriteString("  " + styleOptionSelected.Render(m.values.targetDir) + "\n\n")

	sb.WriteString(styleInspectorTitle.Render("Architecture:") + "\n")
	sb.WriteString(fmt.Sprintf("  • Go 1.27 %s application with %s router\n", m.values.appType, m.values.router))
	sb.WriteString(fmt.Sprintf("  • Database: %s (migrations: %t)\n", m.values.database, m.values.migrations))
	sb.WriteString(fmt.Sprintf("  • Configuration: %s · Logging: %s\n", m.values.configSource, m.values.logFormat))
	sb.WriteString(fmt.Sprintf("  • Accounts: %s\n\n", m.values.accounts))

	sb.WriteString(styleInspectorTitle.Render("Next steps:") + "\n")
	sb.WriteString(fmt.Sprintf("  cd %s\n", m.values.targetDir))
	sb.WriteString("  make help\n")
	sb.WriteString("  make run\n\n")

	sb.WriteString(formatHelp([2]string{"enter/q", "exit"}))
	return box.Render(sb.String())
}

func (m *model) viewFailed() string {
	box := lipgloss.NewStyle().Padding(2, 3)
	var sb strings.Builder
	sb.WriteString(styleError.Render("GOCRAFT — GENERATION FAILED") + "\n\n")
	if m.genErr != nil {
		sb.WriteString(styleOptionNormal.Render(m.genErr.Error()) + "\n\n")
	}
	sb.WriteString(formatHelp([2]string{"enter/q", "exit"}))
	return box.Render(sb.String())
}
