package ui

import (
	"log/slog"
	"strings"

	"github.com/VJ-2303/gocraft/internal/generator"
	"github.com/VJ-2303/gocraft/internal/spec"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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
	errText string

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
