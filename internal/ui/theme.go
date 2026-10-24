package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// A quiet navy palette: the questions are the interface, everything
// else stays out of the way.
var (
	primaryColor = lipgloss.Color("32")  // navy blue
	mutedColor   = lipgloss.Color("245") // soft gray
	accentColor  = lipgloss.Color("39")  // bright blue
	errorColor   = lipgloss.Color("203") // warm red
	successColor = lipgloss.Color("35")  // green
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor)

	descStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	counterStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	barDoneStyle = lipgloss.NewStyle().
			Foreground(primaryColor)

	barTodoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("238"))

	cursorStyle = lipgloss.NewStyle().
			Foreground(accentColor)

	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor)

	commandStyle = lipgloss.NewStyle().
			Foreground(accentColor)

	successStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(successColor)

	warningStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(errorColor)

	hintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("238"))

	selectorMark = cursorStyle.Render("❯")

	selectedValueStyle = lipgloss.NewStyle().Bold(true)
)

// progress renders the progress bar and the n/total counter.
func progress(position, total int) string {
	segments := total
	if segments > 24 {
		segments = 24
	}
	done := 0
	if total > 0 {
		done = position * segments / total
	}
	bar := barDoneStyle.Render(strings.Repeat("▰", done)) +
		barTodoStyle.Render(strings.Repeat("▱", segments-done))

	return bar + " " + counterStyle.Render(fmt.Sprintf("%d/%d", position, total))
}

// questionLine renders the question title and its description on one
// compact line.
func questionLine(title, desc string) string {
	line := titleStyle.Render(title)
	if desc != "" {
		line += " " + descStyle.Render("— "+desc)
	}
	return line
}

// help renders the footer hint line for a question.
func help(hints ...string) string {
	return hintStyle.Render(strings.Join(hints, " · "))
}
