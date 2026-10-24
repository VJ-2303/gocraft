package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Styles for the wizard. A quiet navy palette, generous whitespace, no
// boxes or borders — the questions are the interface.
var (
	primaryColor = lipgloss.Color("32")  // navy blue
	mutedColor   = lipgloss.Color("245") // soft gray
	accentColor  = lipgloss.Color("39")  // bright blue
	errorColor   = lipgloss.Color("203") // warm red
	successColor = lipgloss.Color("35")  // green
)

var (
	brandStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Background(primaryColor).
			Padding(0, 1)

	versionStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	counterStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	barDoneStyle = lipgloss.NewStyle().
			Foreground(primaryColor)

	barTodoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("238"))

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor)

	descStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

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
)

// header renders the brand line, the progress bar and the position
// counter for the current question.
func header(version string, position, total int) string {
	brand := brandStyle.Render(" nomoreboiler ") + versionStyle.Render(" "+version)

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

	counter := counterStyle.Render(fmt.Sprintf("%d/%d", position, total))

	return brand + "\n" + bar + " " + counter
}

// help renders the footer hint line for a question.
func help(hints ...string) string {
	return hintStyle.Render(strings.Join(hints, " · "))
}
