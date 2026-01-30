package app

import (
	"github.com/Fuwn/faustus/internal/claude"
	"strings"
)

func (m *Model) jumpToSearchResult() {
	if len(m.deepSearchResults) == 0 {
		return
	}

	result := m.deepSearchResults[m.deepSearchIndex]
	setupPreview := func() {
		m.invalidatePreviewCache()

		m.showPreview = true
		m.previewFocus = true
		m.previewSearchQuery = m.deepSearchQuery

		if preview := m.preview(); preview != nil {
			m.previewSearchMatches = claude.SearchPreview(preview, m.deepSearchQuery)
			m.previewSearchIndex = 0

			if len(m.previewSearchMatches) > 0 && result.Content != "" {
				bestMatch := 0

				for matchIndex, messageIndex := range m.previewSearchMatches {
					if messageIndex < len(preview.Messages) {
						if strings.Contains(strings.ToLower(preview.Messages[messageIndex].Content),
							strings.ToLower(extractSearchSnippet(result.Content))) {
							bestMatch = matchIndex

							break
						}
					}
				}

				m.previewSearchIndex = bestMatch
			}

			m.scrollToPreviewMatch()
		}
	}

	for index, session := range m.filtered {
		if session.SessionID == result.Session.SessionID {
			m.cursor = index

			m.ensureVisible()
			setupPreview()

			return
		}
	}

	for _, session := range m.sessions {
		if session.SessionID == result.Session.SessionID {
			if session.InTrash && m.tab != TabTrash {
				m.tab = TabTrash

				m.updateFiltered()
			} else if !session.InTrash && m.tab != TabSessions {
				m.tab = TabSessions

				m.updateFiltered()
			}

			for filteredIndex, filteredSession := range m.filtered {
				if filteredSession.SessionID == session.SessionID {
					m.cursor = filteredIndex

					m.ensureVisible()
					setupPreview()

					break
				}
			}

			break
		}
	}
}

func extractSearchSnippet(content string) string {
	content = strings.TrimPrefix(content, "… ")
	content = strings.TrimSuffix(content, " …")
	content = strings.TrimSpace(content)

	if len(content) > 30 {
		content = content[:30]
	}

	return content
}
