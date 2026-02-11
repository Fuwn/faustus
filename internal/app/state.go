package app

import (
	"github.com/Fuwn/faustus/internal/claude"
	"strings"
	"time"
)

func (m *Model) updateFiltered() {
	query := strings.ToLower(m.searchInput.Value())
	m.filtered = nil

	for _, session := range m.sessions {
		if m.tab == TabTrash && !session.InTrash {
			continue
		}

		if m.tab == TabSessions && session.InTrash {
			continue
		}

		if query != "" {
			searchable := strings.ToLower(session.Summary + " " + session.FirstPrompt + " " + session.ProjectName + " " + session.GitBranch)

			if !strings.Contains(searchable, query) {
				continue
			}
		}

		m.filtered = append(m.filtered, session)
	}

	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
}

func (m *Model) setMessage(statusMessage string) {
	m.message = statusMessage
	m.messageTime = time.Now()
}

func (m *Model) selectedSession() *claude.Session {
	if m.cursor >= 0 && m.cursor < len(m.filtered) {
		sessionID := m.filtered[m.cursor].SessionID

		for index := range m.sessions {
			if m.sessions[index].SessionID == sessionID {
				return &m.sessions[index]
			}
		}
	}

	return nil
}

func (m *Model) updateFilteredFromOriginal() {
	if m.cursor >= 0 && m.cursor < len(m.filtered) {
		sessionID := m.filtered[m.cursor].SessionID

		for _, session := range m.sessions {
			if session.SessionID == sessionID {
				m.filtered[m.cursor] = session

				break
			}
		}
	}
}

func (m *Model) reloadSessions() {
	sessions, loadError := claude.LoadAllSessions()

	if loadError != nil {
		m.setMessage("Reload error: " + loadError.Error())

		return
	}

	m.sessions = sessions

	m.updateFiltered()
	m.invalidatePreviewCache()

	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
}

func (m *Model) ensureVisible() {
	visible := m.visibleItemCount()

	if m.cursor < m.offset {
		m.offset = m.cursor
	}

	if m.cursor >= m.offset+visible {
		m.offset = m.cursor - visible + 1
	}
}

func (m Model) listWidth() int {
	if m.showPreview {
		return m.width / 2
	}

	return m.width
}

func (m Model) previewWidth() int {
	return m.width - m.listWidth() - 3
}

func (m Model) listHeight() int {
	reserved := 8

	if m.showHelp {
		reserved += 12
	}

	return max(1, m.height-reserved)
}

func (m Model) visibleItemCount() int {
	if m.showPreview {
		return m.listHeight() - 2
	}

	return max(1, m.listHeight()/2)
}
