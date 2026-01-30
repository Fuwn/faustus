package app

import (
	"fmt"
	"github.com/Fuwn/faustus/internal/claude"
	"github.com/Fuwn/faustus/internal/ui"
	"github.com/charmbracelet/lipgloss"
	"strings"
	"time"
)

func (m Model) View() string {
	if m.width == 0 {
		return "Loading …"
	}

	var builder strings.Builder

	builder.WriteString(m.renderHeader())
	builder.WriteString("\n")
	builder.WriteString(m.renderTabs())
	builder.WriteString("\n\n")

	if m.mode == ModeSearch {
		builder.WriteString(m.renderSearch())
		builder.WriteString("\n")
	} else if m.mode == ModeDeepSearch {
		builder.WriteString(m.renderDeepSearch())
		builder.WriteString("\n")
	} else if m.searchInput.Value() != "" {
		builder.WriteString(ui.SearchStyle.Render("/ " + m.searchInput.Value()))
		builder.WriteString("\n")
	}

	if m.deepSearchQuery != "" && len(m.deepSearchResults) > 0 {
		status := fmt.Sprintf("Search: \"%s\" • %d of %d • n / N to navigate",
			m.deepSearchQuery, m.deepSearchIndex+1, len(m.deepSearchResults))

		builder.WriteString(ui.SearchMatchStyle.Render(status))
		builder.WriteString("\n")
	}

	if m.mode == ModeRename {
		builder.WriteString(m.renderRename())
		builder.WriteString("\n")
	}

	if m.mode == ModeReassign {
		builder.WriteString(m.renderReassign())
		builder.WriteString("\n")
	}

	if m.mode == ModeConfirm {
		builder.WriteString(m.renderConfirm())
		builder.WriteString("\n\n")
	}

	if m.showPreview {
		builder.WriteString(m.renderSplitView())
	} else {
		builder.WriteString(m.renderList())
	}

	if m.message != "" && time.Since(m.messageTime) < 3*time.Second {
		builder.WriteString("\n")
		builder.WriteString(ui.StatusBarStyle.Render(m.message))
	}

	if m.showHelp {
		builder.WriteString("\n")
		builder.WriteString(m.renderHelp())
	} else {
		builder.WriteString("\n")

		previewHint := ""

		if m.showPreview {
			if m.previewFocus {
				previewHint = " • Preview focused"
			} else {
				previewHint = " • Tab to focus preview"
			}
		}

		builder.WriteString(ui.HelpStyle.Render("? Help • j/k Navigate • h/l Tabs • / Filter • p Preview" + previewHint))
	}

	return builder.String()
}

func (m Model) renderSplitView() string {
	listWidth := m.listWidth()
	previewWidth := m.previewWidth()
	listHeight := m.listHeight()
	listContent := m.renderListCompact(listWidth-2, listHeight-2)
	previewContent := m.renderPreview(previewWidth-2, listHeight-2)

	var listStyle, previewStyle lipgloss.Style

	if m.previewFocus {
		listStyle = ui.ListBoxStyle
		previewStyle = ui.PreviewFocusedStyle
	} else {
		listStyle = ui.ListBoxFocusedStyle
		previewStyle = ui.PreviewStyle
	}

	listBox := listStyle.
		Width(listWidth).
		Height(listHeight).
		Render(listContent)
	previewBox := previewStyle.
		Width(previewWidth).
		Height(listHeight).
		Render(previewContent)

	return lipgloss.JoinHorizontal(lipgloss.Top, listBox, " ", previewBox)
}

func (m Model) renderListCompact(width, height int) string {
	if len(m.filtered) == 0 {
		if m.tab == TabTrash {
			return ui.MetaStyle.Render("  Bin is empty")
		}

		if m.searchInput.Value() != "" {
			return ui.MetaStyle.Render("  No matching sessions")
		}

		return ui.MetaStyle.Render("  No sessions")
	}

	var builder strings.Builder

	for index := m.offset; index < min(m.offset+height, len(m.filtered)); index++ {
		session := m.filtered[index]
		isSelected := index == m.cursor

		builder.WriteString(m.renderSessionCompact(&session, isSelected, width))
		builder.WriteString("\n")
	}

	if len(m.filtered) > height {
		indicator := fmt.Sprintf("[%d/%d]", m.cursor+1, len(m.filtered))

		builder.WriteString(ui.MetaStyle.Render(indicator))
	}

	return builder.String()
}

func (m Model) renderSessionCompact(session *claude.Session, isSelected bool, maxWidth int) string {
	cursor := "  "

	if isSelected {
		cursor = ui.CursorStyle.Render("▸ ")
	}

	summary := session.Summary

	if summary == "" {
		summary = truncate(session.FirstPrompt, 40)
	}

	if summary == "" {
		summary = "(No summary)"
	}

	maxSummary := maxWidth - 4
	summary = truncate(summary, maxSummary)

	if isSelected {
		return cursor + ui.SelectedItemStyle.Render(summary)
	}

	return cursor + ui.TitleStyle.Render(summary)
}

func (m Model) renderPreview(width, height int) string {
	preview := m.preview()

	if preview == nil {
		return ui.MetaStyle.Render("No session selected")
	}

	if preview.Error != "" {
		return ui.MetaStyle.Render(preview.Error)
	}

	var lines []string

	if m.cursor < len(m.filtered) {
		session := &m.filtered[m.cursor]
		header := ui.PreviewHeaderStyle.Render(truncate(session.Summary, width-4))
		lines = append(lines, header)
		meta := ui.MetaStyle.Render(fmt.Sprintf("%s • %s • %d messages",
			session.ProjectName, formatTime(session.Modified), len(preview.Messages)))
		lines = append(lines, meta)
		lines = append(lines, ui.PreviewDividerStyle.Render(strings.Repeat("─", width-4)))
		lines = append(lines, "")
	}

	for messageIndex, previewMessage := range preview.Messages {
		var roleStyle, contentStyle lipgloss.Style
		var prefix string

		isMatch := false
		isCurrentMatch := false

		for matchNumber, matchMessageIndex := range m.previewSearchMatches {
			if matchMessageIndex == messageIndex {
				isMatch = true

				if matchNumber == m.previewSearchIndex {
					isCurrentMatch = true
				}

				break
			}
		}

		switch previewMessage.Role {
		case "user":
			roleStyle = ui.UserRoleStyle
			contentStyle = ui.UserContentStyle
			prefix = "You"
		case "assistant":
			roleStyle = ui.AssistantRoleStyle
			contentStyle = ui.AssistantContentStyle
			prefix = "Claude"
		case "tool":
			roleStyle = ui.ToolRoleStyle
			contentStyle = ui.ToolContentStyle
			prefix = "Tool"
		case "thinking":
			roleStyle = ui.ThinkingRoleStyle
			contentStyle = ui.ThinkingContentStyle
			prefix = "Thinking"
		}

		matchIndicator := ""

		if isCurrentMatch {
			matchIndicator = ui.HighlightStyle.Render(" ◀ ")
		} else if isMatch {
			matchIndicator = ui.SearchMatchStyle.Render(" ● ")
		}

		lines = append(lines, roleStyle.Render(prefix+":")+matchIndicator)
		content := previewMessage.Content

		if m.previewSearchQuery != "" && isMatch {
			content = highlightMatches(content, m.previewSearchQuery)
		}

		wrapped := wrapText(content, width-6)

		for _, line := range strings.Split(wrapped, "\n") {
			lines = append(lines, "  "+contentStyle.Render(line))
		}

		lines = append(lines, "")
	}

	maxScroll := max(0, len(lines)-height+1)
	scroll := m.previewScroll

	if scroll < 0 {
		scroll = 0
	}

	if scroll > maxScroll {
		scroll = maxScroll
	}

	if scroll > 0 && scroll < len(lines) {
		lines = lines[scroll:]
	}

	if len(lines) > height {
		lines = lines[:height]
	}

	if maxScroll > 0 {
		indicator := fmt.Sprintf("─── %d/%d ───", scroll+1, maxScroll+1)

		if len(lines) > 0 {
			lines[len(lines)-1] = ui.MetaStyle.Render(indicator)
		}
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderHeader() string {
	logo := ui.LogoStyle.Render("🛎️ Faustus")
	subtitle := ui.MetaStyle.Render(" • Session Manager for Claude Code")
	count := ui.CountStyle.Render(fmt.Sprintf("%d sessions", len(m.filtered)))
	gap := m.width - lipgloss.Width(logo) - lipgloss.Width(subtitle) - lipgloss.Width(count) - 4

	if gap < 0 {
		gap = 0
	}

	return logo + subtitle + strings.Repeat(" ", gap) + count
}

func (m Model) renderTabs() string {
	var sessionsCount, trashCount int

	for _, session := range m.sessions {
		if session.InTrash {
			trashCount += 1
		} else {
			sessionsCount += 1
		}
	}

	sessionsTab := fmt.Sprintf("Sessions (%d)", sessionsCount)
	binTab := fmt.Sprintf("Bin (%d)", trashCount)

	if m.tab == TabSessions {
		return ui.ActiveTabStyle.Render("● "+sessionsTab) + "    " + ui.TabStyle.Render(binTab)
	}

	return ui.TabStyle.Render(sessionsTab) + "    " + ui.ActiveTabStyle.Render("● "+binTab)
}

func (m Model) renderSearch() string {
	label := "/"

	if m.showPreview && m.previewFocus {
		label = "/ (preview)"
	}

	return ui.SearchInputStyle.Render(label + " " + m.searchInput.View())
}

func (m Model) renderDeepSearch() string {
	return ui.SearchInputStyle.Render("s: " + m.deepSearchInput.View())
}

func (m Model) renderRename() string {
	return ui.SearchInputStyle.Render("✏️  " + m.renameInput.View())
}

func (m Model) renderReassign() string {
	label := "📁 Reassign folder"

	if m.reassignAll {
		label = "📁 Reassign ALL sessions with this folder"
	}

	return ui.SearchInputStyle.Render(label + ": " + m.reassignInput.View())
}

func (m Model) renderConfirm() string {
	var confirmMessage string

	switch m.confirmAction {
	case ConfirmDelete:
		confirmMessage = "Move this session to the Bin?"
	case ConfirmRestore:
		confirmMessage = "Restore this session from the Bin?"
	case ConfirmPermanentDelete:
		confirmMessage = "Delete this session permanently? This cannot be undone."
	case ConfirmEmptyTrash:
		confirmMessage = "Empty the Bin? All sessions will be permanently deleted."
	}

	return ui.ModalStyle.Render(
		ui.ConfirmStyle.Render(confirmMessage) + "\n\n" +
			ui.HelpKeyStyle.Render("y") + ui.HelpStyle.Render(" confirm  ") +
			ui.HelpKeyStyle.Render("n/esc") + ui.HelpStyle.Render(" cancel"),
	)
}

func (m Model) renderList() string {
	if len(m.filtered) == 0 {
		if m.tab == TabTrash {
			return ui.MetaStyle.Render("  Bin is empty")
		}

		if m.searchInput.Value() != "" {
			return ui.MetaStyle.Render("  No matching sessions")
		}

		return ui.MetaStyle.Render("  No sessions")
	}

	var builder strings.Builder

	listHeight := m.listHeight()

	for index := m.offset; index < min(m.offset+listHeight, len(m.filtered)); index++ {
		session := m.filtered[index]
		isSelected := index == m.cursor

		builder.WriteString(m.renderSession(&session, isSelected))
		builder.WriteString("\n")
	}

	if len(m.filtered) > listHeight {
		position := float64(m.offset) / float64(len(m.filtered)-listHeight)
		indicator := fmt.Sprintf(" [%d-%d of %d]", m.offset+1, min(m.offset+listHeight, len(m.filtered)), len(m.filtered))

		builder.WriteString(ui.MetaStyle.Render(indicator))

		scrollPosition := int(position * 10)
		scrollBar := strings.Repeat("─", scrollPosition) + "●" + strings.Repeat("─", 10-scrollPosition)

		builder.WriteString(" " + ui.MetaStyle.Render(scrollBar))
	}

	return builder.String()
}

func (m Model) renderSession(session *claude.Session, isSelected bool) string {
	var builder strings.Builder

	cursor := "  "

	if isSelected {
		cursor = ui.CursorStyle.Render("▸ ")
	}

	builder.WriteString(cursor)

	summary := session.Summary

	if summary == "" {
		summary = truncate(session.FirstPrompt, 60)
	}

	if summary == "" {
		summary = "(No summary)"
	}

	if isSelected {
		builder.WriteString(ui.SelectedItemStyle.Render(truncate(summary, m.width-20)))
	} else {
		builder.WriteString(ui.TitleStyle.Render(truncate(summary, m.width-20)))
	}

	builder.WriteString("\n")

	meta := fmt.Sprintf("    %s", ui.ProjectStyle.Render(session.ProjectName))

	if session.GitBranch != "" {
		meta += ui.MetaStyle.Render(" @ " + session.GitBranch)
	}

	meta += ui.MetaStyle.Render(fmt.Sprintf(" • %d messages • %s", session.MessageCount, formatTime(session.Modified)))

	if session.InTrash {
		meta += " " + ui.TrashStyle.Render("In Bin")
	}

	builder.WriteString(meta)

	return builder.String()
}

func (m Model) renderHelp() string {
	var builder strings.Builder

	builder.WriteString(ui.HeaderStyle.Render("Keyboard Shortcuts"))
	builder.WriteString("\n\n")

	helpItems := []struct{ key, description string }{
		{"j / k", "Navigate up and down"},
		{"h / l", "Switch between tabs"},
		{"g g / G", "Jump to top or bottom"},
		{"ctrl+u / ctrl+d", "Page up or down"},
		{"/", "Filter sessions"},
		{"s", "Search all sessions"},
		{"n / N", "Next or previous match"},
		{"p", "Toggle preview pane"},
		{"tab", "Switch focus"},
		{"d", "Move to Bin"},
		{"u", "Restore from Bin"},
		{"c", "Rename session"},
		{"r", "Reassign folder"},
		{"R", "Reassign all with folder"},
		{"D", "Empty Bin"},
		{"?", "Show help"},
		{"q", "Quit"},
	}

	for _, item := range helpItems {
		builder.WriteString("  ")
		builder.WriteString(ui.HelpKeyStyle.Render(fmt.Sprintf("%-16s", item.key)))
		builder.WriteString(ui.HelpStyle.Render(item.description))
		builder.WriteString("\n")
	}

	return builder.String()
}
