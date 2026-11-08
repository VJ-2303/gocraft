package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// wrapToWidth breaks s into lines of at most width runes, preferring to
// break at spaces; tokens longer than width are hard-broken.
func wrapToWidth(s string, width int) []string {
	if width <= 0 || s == "" {
		return []string{s}
	}
	var lines []string
	for len([]rune(s)) > width {
		r := []rune(s)
		cut := -1
		for i := width; i > 0; i-- {
			if r[i] == ' ' {
				cut = i
				break
			}
		}
		if cut <= 0 {
			cut = width
			lines = append(lines, string(r[:cut]))
			s = string(r[cut:])
			continue
		}
		lines = append(lines, string(r[:cut]))
		s = strings.TrimLeft(string(r[cut:]), " ")
	}
	return append(lines, s)
}

// renderRow renders "prefix value" with the value wrapped to the given
// width. Continuation lines align with the value column.
func renderRow(prefix string, prefixWidth int, value string, width int, valueStyle lipgloss.Style) string {
	body := wrapToWidth(value, width-prefixWidth)
	indent := strings.Repeat(" ", prefixWidth)
	var sb strings.Builder
	sb.WriteString(prefix)
	sb.WriteString(valueStyle.Render(body[0]))
	for _, l := range body[1:] {
		sb.WriteString("\n")
		sb.WriteString(indent)
		sb.WriteString(valueStyle.Render(l))
	}
	return sb.String()
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

	// The review step duplicates the blueprint; give it the full width.
	if m.current().kind == kindReview {
		return m.viewLeftPane(m.width)
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

	// Top bar
	var stepText string
	if m.stepIdx == m.reviewIdx() {
		stepText = "review"
	} else {
		stepText = fmt.Sprintf("step %d/%d · %s", m.visiblePos(m.stepIdx), m.visibleTotal(), strings.ToLower(m.current().category))
	}
	topBar := styleCategory.Render("gocraft") + " " + styleDesc.Render("· "+stepText)
	if m.editingFromReview {
		topBar += " " + styleDesc.Render("· editing from review")
	}
	content.WriteString(topBar)
	content.WriteString("\n\n")

	// Conflict confirmation view
	if m.state == stateConflictConfirm {
		content.WriteString(m.renderConflictBox(width - 6))
		return box.Render(content.String())
	}

	// Question
	q := m.current()
	content.WriteString(styleTitle.Render(q.title))
	content.WriteString("\n")
	if q.desc != "" {
		content.WriteString(styleDesc.Render(q.desc))
		content.WriteString("\n\n")
	} else {
		content.WriteString("\n")
	}

	// Body by kind
	switch q.kind {
	case kindReview:
		content.WriteString(m.renderReviewTable(width - 6))
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
			content.WriteString(prefix)
			content.WriteString(line)
			content.WriteString("\n")
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
			content.WriteString(prefix)
			content.WriteString(line)
			content.WriteString("\n")
		}
	case kindText:
		content.WriteString(styleInputPrompt.Render("❯ "))
		content.WriteString(m.textInput.View())
		content.WriteString("\n")
	}

	// Hint line: error, pending-quit notice, or blank
	content.WriteString("\n")
	switch {
	case m.errText != "":
		content.WriteString(styleError.Render("✗ " + m.errText))
		content.WriteString("\n")
	case m.pendingCtrlC:
		content.WriteString(styleDesc.Render("press ctrl+c again to quit"))
		content.WriteString("\n")
	default:
		content.WriteString("\n")
	}

	// Keybindings footer
	content.WriteString("\n")
	if q.kind == kindReview {
		content.WriteString(formatHelp([2]string{"1-9", "edit"}, [2]string{"enter", "generate"}, [2]string{"esc", "back"}, [2]string{"ctrl+c", "quit"}))
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
	var b strings.Builder
	for _, l := range wrapToWidth(m.conflictMsg, width) {
		b.WriteString(styleConflictBadge.Render(l))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(styleKeyHint.Render("apply and continue? "))
	b.WriteString(styleKey.Render("[y/N]"))
	b.WriteString("\n\n")
	b.WriteString(formatHelp([2]string{"y", "apply"}, [2]string{"n/esc", "keep current answers"}))
	return b.String()
}

func (m *model) renderReviewTable(width int) string {
	rows := []struct {
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
	for _, r := range rows {
		prefix := styleKey.Render("["+r.num+"]") + " " + styleDesc.Render(fmt.Sprintf("%-14s", r.label)) + " "
		sb.WriteString(renderRow(prefix, 19, r.val, width, styleOptionSelected))
		sb.WriteString("\n")
	}
	return sb.String()
}

func (m *model) viewRightPane(width int) string {
	var sb strings.Builder

	sb.WriteString(styleInspectorTitle.Render("summary"))
	sb.WriteString("\n")

	row := func(k, v string) {
		prefix := fmt.Sprintf("%-10s", k) + " "
		sb.WriteString(renderRow(styleInspectorText.Render(prefix), 11, v, width-2, styleInspectorActive))
		sb.WriteString("\n")
	}
	row("type", string(m.values.appType))
	row("module", m.values.modulePath)
	row("database", string(m.values.database))
	row("router", string(m.values.router))
	row("config", string(m.values.configSource))
	row("logging", string(m.values.logFormat))
	row("accounts", m.values.accounts)

	var deps []string
	if m.values.accounts == "full" {
		deps = append(deps, "full accounts → sqlite + sessions + emails + migrations")
	} else if m.values.accounts == "jwt" {
		deps = append(deps, "jwt accounts → sqlite + emails + migrations")
	}
	if m.values.migrations {
		deps = append(deps, "migrations → database")
	}
	if m.values.sessions {
		deps = append(deps, "sessions → database")
	}
	if m.values.errorNotifs {
		deps = append(deps, "error notifications → emails")
	}

	if len(deps) > 0 {
		sb.WriteString("\n")
		sb.WriteString(styleInspectorTitle.Render("also included"))
		sb.WriteString("\n")
		for _, d := range deps {
			sb.WriteString(renderRow(styleBadge.Render("· "), 2, d, width-2, styleBadge))
			sb.WriteString("\n")
		}
	}

	pane := lipgloss.NewStyle().Width(width).PaddingLeft(2)
	return pane.Render(sb.String())
}

func (m *model) viewStacked() string {
	return m.viewLeftPane(m.width)
}

func (m *model) viewGenerating() string {
	box := lipgloss.NewStyle().Padding(2, 3)
	var sb strings.Builder
	sb.WriteString(styleCategory.Render("gocraft"))
	sb.WriteString("\n\n")
	sb.WriteString(fmt.Sprintf("%s Scaffolding codebase...\n", m.spinner.View()))
	sb.WriteString(styleDesc.Render("Writing template files, configuring dependencies, and resolving Go modules.\n"))
	return box.Render(sb.String())
}

func (m *model) viewDone() string {
	box := lipgloss.NewStyle().Padding(2, 3)
	var sb strings.Builder
	sb.WriteString(styleCategory.Render("gocraft"))
	sb.WriteString(" ")
	sb.WriteString(styleDesc.Render("· project ready"))
	sb.WriteString("\n\n")
	sb.WriteString(fmt.Sprintf("Successfully scaffolded %d files into\n", m.genCount))
	sb.WriteString(styleOptionSelected.Render("  " + m.values.targetDir))
	sb.WriteString("\n\n")

	sb.WriteString(styleInspectorTitle.Render("summary"))
	sb.WriteString("\n")
	row := func(k, v string) {
		sb.WriteString(renderRow(styleInspectorText.Render(fmt.Sprintf("%-10s", k)+" "), 11, v, 60, styleInspectorActive))
		sb.WriteString("\n")
	}
	row("type", fmt.Sprintf("go 1.27 %s · %s router", m.values.appType, m.values.router))
	row("database", fmt.Sprintf("%s (migrations: %t)", m.values.database, m.values.migrations))
	row("config", fmt.Sprintf("%s · %s", m.values.configSource, m.values.logFormat))
	row("accounts", m.values.accounts)

	sb.WriteString("\n")
	sb.WriteString(styleInspectorTitle.Render("next steps"))
	sb.WriteString("\n")
	sb.WriteString(styleKey.Render("  cd " + m.values.targetDir))
	sb.WriteString("\n")
	sb.WriteString(styleKey.Render("  make help"))
	sb.WriteString("\n")
	sb.WriteString(styleKey.Render("  make run"))
	sb.WriteString("\n\n")

	sb.WriteString(formatHelp([2]string{"enter/q", "exit"}))
	return box.Render(sb.String())
}

func (m *model) viewFailed() string {
	box := lipgloss.NewStyle().Padding(2, 3)
	var sb strings.Builder
	sb.WriteString(styleError.Render("gocraft · generation failed"))
	sb.WriteString("\n\n")
	if m.genErr != nil {
		sb.WriteString(styleOptionNormal.Render(m.genErr.Error()))
		sb.WriteString("\n\n")
	}
	sb.WriteString(formatHelp([2]string{"enter/q", "exit"}))
	return box.Render(sb.String())
}
