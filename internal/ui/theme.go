// Package ui renders the interactive terminal interface for gocraft.
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Warm Editorial Palette:
// Grounded charcoal, warm muted stone, and rich copper/amber accent.
// Adheres strictly to the anti-AI-slop frontend-design principles:
// restrained, typography-first, no purple gradients, no glowing boxes.
var (
	colorFg      = lipgloss.Color("255") // #f5f5f4 chalk white
	colorMuted   = lipgloss.Color("244") // #78716c warm stone gray
	colorBorder  = lipgloss.Color("238") // #27272a dark slate border
	colorAccent  = lipgloss.Color("178") // #d97706 warm amber / copper
	colorSuccess = lipgloss.Color("71")  // #16a34a forest green / sage
	colorWarning = lipgloss.Color("167") // #dc2626 warm terracotta
	colorBadgeBg = lipgloss.Color("236") // subtle badge background
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
			Foreground(colorAccent)

	styleBadge = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Background(colorBadgeBg).
			Padding(0, 1)

	styleConflictBadge = lipgloss.NewStyle().
				Foreground(colorWarning).
				Background(colorBadgeBg).
				Padding(0, 1)

	styleInspectorTitle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorAccent)

	styleInspectorText = lipgloss.NewStyle().
				Foreground(colorMuted)

	styleInspectorActive = lipgloss.NewStyle().
				Foreground(colorFg)

	styleKeyHint = lipgloss.NewStyle().
			Foreground(colorMuted)

	styleKey = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

	stylePaneBorder = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(colorBorder).
			PaddingLeft(2)

	styleInput = lipgloss.NewStyle().
			Foreground(colorFg).
			Bold(true)

	styleInputPrompt = lipgloss.NewStyle().
				Foreground(colorAccent).
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
