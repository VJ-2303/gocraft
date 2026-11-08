// Package ui renders the interactive terminal interface for gocraft.
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Palette: neutral base with a single muted steel-blue accent. Restrained,
// typography-first. No shouting caps, no background chips behind text.
var (
	colorFg      = lipgloss.Color("254") // near-white, body text
	colorMuted   = lipgloss.Color("245") // warm neutral gray, secondary text
	colorAccent  = lipgloss.Color("67")  // muted steel blue, interactive accent
	colorSuccess = lipgloss.Color("108") // subdued sage green
	colorWarning = lipgloss.Color("131") // muted brick red
	colorSteel   = lipgloss.Color("110") // brighter steel, keys and markers
)

var (
	styleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorFg)

	styleCategory = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorAccent)

	styleDesc = lipgloss.NewStyle().
			Foreground(colorMuted)

	styleOptionNormal = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	styleOptionSelected = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorFg)

	styleCursor = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorSteel)

	styleBadge = lipgloss.NewStyle().
			Foreground(colorSuccess)

	styleConflictBadge = lipgloss.NewStyle().
				Foreground(colorWarning)

	styleInspectorTitle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorMuted)

	styleInspectorText = lipgloss.NewStyle().
				Foreground(colorMuted)

	styleInspectorActive = lipgloss.NewStyle().
				Foreground(colorFg)

	styleKeyHint = lipgloss.NewStyle().
			Foreground(colorMuted)

	styleKey = lipgloss.NewStyle().
			Foreground(colorSteel).
			Bold(true)

	styleInput = lipgloss.NewStyle().
			Foreground(colorFg).
			Bold(true)

	styleInputPrompt = lipgloss.NewStyle().
				Foreground(colorSteel).
				Bold(true)

	styleError = lipgloss.NewStyle().
			Foreground(colorWarning).
			Bold(true)
)

// formatHelp returns a formatted keybinding hint string.
func formatHelp(hints ...[2]string) string {
	var parts []string
	for _, h := range hints {
		parts = append(parts, fmt.Sprintf("%s %s", styleKey.Render(h[0]), styleKeyHint.Render(h[1])))
	}
	return strings.Join(parts, styleKeyHint.Render(" · "))
}
