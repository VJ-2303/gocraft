package ui

import (
	"context"
	"fmt"
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
)

// Run runs the interactive TUI wizard.
func Run(ctx context.Context, logger *slog.Logger) error {
	m := newModel(nil, logger)
	p := tea.NewProgram(m, tea.WithContext(ctx), tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("running wizard: %w", err)
	}

	fm, ok := finalModel.(*model)
	if ok && fm.state == stateFailed && fm.genErr != nil {
		return fm.genErr
	}
	return nil
}
