package ui

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

// spinnerModel is the spinner shown while the codebase generates.
type spinnerModel = spinner.Model

// newSpinnerModel returns the wizard's spinner.
func newSpinnerModel() spinnerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(accentColor)
	return s
}
