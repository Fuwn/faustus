package app

import (
	"fmt"
	"time"

	"github.com/Fuwn/faustus/internal/claude"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch typedMessage := message.(type) {
	case tea.WindowSizeMsg:
		m.width = typedMessage.Width
		m.height = typedMessage.Height

		return m, nil

	case tea.KeyMsg:
		if time.Since(m.messageTime) > 3*time.Second {
			m.message = ""
		}

		switch m.mode {
		case ModeSearch:
			return m.handleSearchMode(typedMessage)
		case ModeDeepSearch:
			return m.handleDeepSearchMode(typedMessage)
		case ModeRename:
			return m.handleRenameMode(typedMessage)
		case ModeConfirm:
			return m.handleConfirmMode(typedMessage)
		default:
			return m.handleNormalMode(typedMessage)
		}
	}

	return m, nil
}

func (m Model) handleNormalMode(keyMessage tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(keyMessage, m.keys.Quit):
		return m, tea.Quit

	case key.Matches(keyMessage, m.keys.Help):
		m.showHelp = !m.showHelp

	case key.Matches(keyMessage, m.keys.Preview):
		m.showPreview = !m.showPreview
		m.previewFocus = false
		m.previewScroll = 0

		m.invalidatePreviewCache()

	case key.Matches(keyMessage, m.keys.Up):
		if m.showPreview && m.previewFocus {
			m.previewScroll -= 1

			m.clampPreviewScroll()
		} else if m.cursor > 0 {
			m.cursor -= 1

			m.ensureVisible()
			m.invalidatePreviewCache()
		}

	case key.Matches(keyMessage, m.keys.Down):
		if m.showPreview && m.previewFocus {
			m.previewScroll += 1

			m.clampPreviewScroll()
		} else if m.cursor < len(m.filtered)-1 {
			m.cursor += 1

			m.ensureVisible()
			m.invalidatePreviewCache()
		}

	case key.Matches(keyMessage, m.keys.HalfUp):
		if m.showPreview && m.previewFocus {
			m.previewScroll -= 10

			m.clampPreviewScroll()
		} else {
			m.cursor = max(0, m.cursor-10)

			m.ensureVisible()
			m.invalidatePreviewCache()
		}

	case key.Matches(keyMessage, m.keys.HalfDown):
		if m.showPreview && m.previewFocus {
			m.previewScroll += 10

			m.clampPreviewScroll()
		} else {
			m.cursor = min(len(m.filtered)-1, m.cursor+10)

			m.ensureVisible()
			m.invalidatePreviewCache()
		}

	case key.Matches(keyMessage, m.keys.Top):
		if m.showPreview && m.previewFocus {
			m.previewScroll = 0
		} else {
			m.cursor = 0

			m.ensureVisible()
			m.invalidatePreviewCache()
		}

	case key.Matches(keyMessage, m.keys.Bottom):
		if m.showPreview && m.previewFocus {
			m.previewScroll = 99999

			m.clampPreviewScroll()
		} else {
			m.cursor = max(0, len(m.filtered)-1)

			m.ensureVisible()
			m.invalidatePreviewCache()
		}

	case key.Matches(keyMessage, m.keys.Tab):
		if m.showPreview {
			m.previewFocus = !m.previewFocus
		} else {
			if m.tab == TabSessions {
				m.tab = TabTrash
			} else {
				m.tab = TabSessions
			}

			m.cursor = 0
			m.offset = 0

			m.updateFiltered()
			m.invalidatePreviewCache()
		}

	case key.Matches(keyMessage, m.keys.Left):
		if m.tab != TabSessions {
			m.tab = TabSessions
			m.cursor = 0
			m.offset = 0

			m.updateFiltered()
			m.invalidatePreviewCache()
		}

	case key.Matches(keyMessage, m.keys.Right):
		if m.tab != TabTrash {
			m.tab = TabTrash
			m.cursor = 0
			m.offset = 0

			m.updateFiltered()
			m.invalidatePreviewCache()
		}

	case key.Matches(keyMessage, m.keys.Search):
		m.mode = ModeSearch

		m.searchInput.Focus()

		return m, textinput.Blink

	case key.Matches(keyMessage, m.keys.Delete):
		if len(m.filtered) > 0 {
			if m.tab == TabTrash {
				m.confirmAction = ConfirmPermanentDelete
			} else {
				m.confirmAction = ConfirmDelete
			}

			m.mode = ModeConfirm
		}

	case key.Matches(keyMessage, m.keys.Restore):
		if len(m.filtered) > 0 && m.tab == TabTrash {
			m.confirmAction = ConfirmRestore
			m.mode = ModeConfirm
		}

	case key.Matches(keyMessage, m.keys.Rename):
		if len(m.filtered) > 0 {
			session := &m.filtered[m.cursor]

			m.renameInput.SetValue(session.Summary)
			m.renameInput.Focus()

			m.mode = ModeRename

			return m, textinput.Blink
		}

	case key.Matches(keyMessage, m.keys.Clear):
		if m.tab == TabTrash {
			m.confirmAction = ConfirmEmptyTrash
			m.mode = ModeConfirm
		}

	case key.Matches(keyMessage, m.keys.DeepSearch):
		m.mode = ModeDeepSearch

		m.deepSearchInput.SetValue(m.deepSearchQuery)
		m.deepSearchInput.Focus()

		return m, textinput.Blink

	case key.Matches(keyMessage, m.keys.NextMatch):
		if m.showPreview && m.previewFocus && len(m.previewSearchMatches) > 0 {
			m.previewSearchIndex = (m.previewSearchIndex + 1) % len(m.previewSearchMatches)

			m.scrollToPreviewMatch()
		} else if len(m.deepSearchResults) > 0 {
			m.deepSearchIndex = (m.deepSearchIndex + 1) % len(m.deepSearchResults)

			m.jumpToSearchResult()
		}

	case key.Matches(keyMessage, m.keys.PrevMatch):
		if m.showPreview && m.previewFocus && len(m.previewSearchMatches) > 0 {
			m.previewSearchIndex -= 1

			if m.previewSearchIndex < 0 {
				m.previewSearchIndex = len(m.previewSearchMatches) - 1
			}

			m.scrollToPreviewMatch()
		} else if len(m.deepSearchResults) > 0 {
			m.deepSearchIndex -= 1

			if m.deepSearchIndex < 0 {
				m.deepSearchIndex = len(m.deepSearchResults) - 1
			}

			m.jumpToSearchResult()
		}
	}

	return m, nil
}

func (m Model) handleSearchMode(keyMessage tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(keyMessage, m.keys.Escape):
		m.mode = ModeNormal

		m.searchInput.Blur()
		m.searchInput.SetValue("")
		m.updateFiltered()

		m.previewSearchQuery = ""
		m.previewSearchMatches = nil

		return m, nil

	case key.Matches(keyMessage, m.keys.Enter):
		m.mode = ModeNormal

		m.searchInput.Blur()

		if m.showPreview && m.previewFocus {
			query := m.searchInput.Value()
			m.previewSearchQuery = query

			if preview := m.preview(); preview != nil {
				m.previewSearchMatches = claude.SearchPreview(preview, query)
				m.previewSearchIndex = 0

				if len(m.previewSearchMatches) > 0 {
					m.scrollToPreviewMatch()
					m.setMessage(fmt.Sprintf("%d matches", len(m.previewSearchMatches)))
				} else if query != "" {
					m.setMessage("No matches")
				}
			}

			m.searchInput.SetValue("")
		} else {
			m.updateFiltered()
		}

		return m, nil
	}

	var command tea.Cmd

	m.searchInput, command = m.searchInput.Update(keyMessage)

	if !m.showPreview || !m.previewFocus {
		m.updateFiltered()
	}

	return m, command
}

func (m Model) handleDeepSearchMode(keyMessage tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(keyMessage, m.keys.Escape):
		m.mode = ModeNormal

		m.deepSearchInput.Blur()

		return m, nil

	case key.Matches(keyMessage, m.keys.Enter):
		query := m.deepSearchInput.Value()

		if query != "" {
			m.deepSearchQuery = query
			m.deepSearchResults = claude.SearchAllSessions(m.sessions, query)
			m.deepSearchIndex = 0

			if len(m.deepSearchResults) > 0 {
				m.jumpToSearchResult()
				m.setMessage(fmt.Sprintf("%d matches across all sessions", len(m.deepSearchResults)))
			} else {
				m.setMessage("No matches")
			}
		}

		m.mode = ModeNormal

		m.deepSearchInput.Blur()

		return m, nil
	}

	var command tea.Cmd

	m.deepSearchInput, command = m.deepSearchInput.Update(keyMessage)

	return m, command
}

func (m Model) handleRenameMode(keyMessage tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(keyMessage, m.keys.Escape):
		m.mode = ModeNormal

		m.renameInput.Blur()

		return m, nil

	case key.Matches(keyMessage, m.keys.Enter):
		if len(m.filtered) > 0 {
			newName := m.renameInput.Value()

			if newName != "" {
				session := m.selectedSession()

				if session != nil {
					if renameError := claude.RenameSession(session, newName); renameError != nil {
						m.setMessage(fmt.Sprintf("Error: %v", renameError))
					} else {
						session.Summary = newName

						m.updateFilteredFromOriginal()
						m.setMessage("Renamed")
					}
				}
			}
		}

		m.mode = ModeNormal

		m.renameInput.Blur()

		return m, nil
	}

	var command tea.Cmd

	m.renameInput, command = m.renameInput.Update(keyMessage)

	return m, command
}

func (m Model) handleConfirmMode(keyMessage tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(keyMessage, m.keys.Escape), keyMessage.String() == "n", keyMessage.String() == "N":
		m.mode = ModeNormal
		m.confirmAction = ConfirmNone

		return m, nil

	case key.Matches(keyMessage, m.keys.Confirm):
		return m.executeConfirmedAction()
	}

	return m, nil
}

func (m Model) executeConfirmedAction() (tea.Model, tea.Cmd) {
	switch m.confirmAction {
	case ConfirmDelete:
		if session := m.selectedSession(); session != nil {
			if deleteError := claude.MoveToTrash(session); deleteError != nil {
				m.setMessage(fmt.Sprintf("Error: %v", deleteError))
			} else {
				m.setMessage("Moved to Bin")
				m.reloadSessions()
			}
		}

	case ConfirmRestore:
		if session := m.selectedSession(); session != nil {
			if restoreError := claude.RestoreFromTrash(session); restoreError != nil {
				m.setMessage(fmt.Sprintf("Error: %v", restoreError))
			} else {
				m.setMessage("Restored")
				m.reloadSessions()
			}
		}

	case ConfirmPermanentDelete:
		if session := m.selectedSession(); session != nil {
			if deleteError := claude.PermanentlyDelete(session); deleteError != nil {
				m.setMessage(fmt.Sprintf("Error: %v", deleteError))
			} else {
				m.setMessage("Permanently deleted")
				m.reloadSessions()
			}
		}

	case ConfirmEmptyTrash:
		if emptyError := claude.EmptyTrash(); emptyError != nil {
			m.setMessage(fmt.Sprintf("Error: %v", emptyError))
		} else {
			m.setMessage("Bin emptied")
			m.reloadSessions()
		}
	}

	m.mode = ModeNormal
	m.confirmAction = ConfirmNone

	return m, nil
}
