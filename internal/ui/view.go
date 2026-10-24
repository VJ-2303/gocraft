package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

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
