package ui

import (
	"context"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/vj/nomoreboiler/internal/generator"
)

// generateTask is a long-running generation job executed while a spinner
// is shown.
type generateTask func() (generator.Result, error)

// taskDoneMsg carries the outcome of the generation task back to the
// spinner model.
type taskDoneMsg struct {
	result generator.Result
	err    error
}

// generateModel shows a spinner while the generation task runs.
type generateModel struct {
	spinner spinner.Model
	task    generateTask
	result  generator.Result
	err     error
	done    bool
}

func newGenerateModel(task generateTask) generateModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return generateModel{spinner: s, task: task}
}

func (m generateModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.runTask)
}

func (m generateModel) runTask() tea.Msg {
	result, err := m.task()
	return taskDoneMsg{result: result, err: err}
}

func (m generateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case spinner.TickMsg:
		if m.done {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case taskDoneMsg:
		m.result = msg.result
		m.err = msg.err
		m.done = true
		return m, tea.Quit
	}
	return m, nil
}

func (m generateModel) View() string {
	if m.done {
		return ""
	}
	return m.spinner.View() + " Generating codebase..."
}

// runWithSpinner executes task, showing a spinner until it completes.
// Bubble Tea handles Ctrl+C itself; the context is accepted so callers can
// propagate cancellation to the generation task in future.
func runWithSpinner(ctx context.Context, task generateTask) (generator.Result, error) {
	_ = ctx

	final, err := tea.NewProgram(newGenerateModel(task)).Run()
	if err != nil {
		return generator.Result{}, err
	}

	m, ok := final.(generateModel)
	if !ok {
		return generator.Result{}, err
	}
	return m.result, m.err
}
