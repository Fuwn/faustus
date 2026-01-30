package app

import (
	"github.com/Fuwn/faustus/internal/claude"
	"strings"
)

func (m *Model) invalidatePreviewCache() {
	m.previewCache = nil
	m.previewFor = ""
	m.previewScroll = 0
}

func (m *Model) preview() *claude.PreviewContent {
	if len(m.filtered) == 0 {
		return nil
	}

	session := &m.filtered[m.cursor]

	if m.previewCache != nil && m.previewFor == session.SessionID {
		return m.previewCache
	}

	previewContent := claude.LoadSessionPreview(session, 50)
	m.previewCache = &previewContent
	m.previewFor = session.SessionID

	return m.previewCache
}

type previewMetrics struct {
	totalLines   int
	messageLines []int
}

func (m *Model) calculatePreviewMetrics() previewMetrics {
	preview := m.preview()

	if preview == nil || preview.Error != "" {
		return previewMetrics{}
	}

	width := m.previewWidth() - 8

	if width <= 0 {
		width = 40
	}

	var metrics previewMetrics

	lineCount := 0

	if m.cursor < len(m.filtered) {
		lineCount += 1
		lineCount += 1
		lineCount += 1
		lineCount += 1
	}

	for _, previewMessage := range preview.Messages {
		metrics.messageLines = append(metrics.messageLines, lineCount)
		lineCount += 1
		wrapped := wrapText(previewMessage.Content, width)
		contentLines := strings.Count(wrapped, "\n") + 1
		lineCount += contentLines
		lineCount += 1
	}

	metrics.totalLines = lineCount

	return metrics
}

func (m *Model) clampPreviewScroll() {
	height := m.listHeight() - 2
	metrics := m.calculatePreviewMetrics()
	maxScroll := max(0, metrics.totalLines-height+1)

	if m.previewScroll < 0 {
		m.previewScroll = 0
	}

	if m.previewScroll > maxScroll {
		m.previewScroll = maxScroll
	}
}

func (m *Model) scrollToPreviewMatch() {
	if len(m.previewSearchMatches) == 0 {
		return
	}

	matchMessageIndex := m.previewSearchMatches[m.previewSearchIndex]
	metrics := m.calculatePreviewMetrics()

	if matchMessageIndex >= len(metrics.messageLines) {
		return
	}

	lineNumber := metrics.messageLines[matchMessageIndex]
	height := m.listHeight() - 2
	m.previewScroll = max(0, lineNumber-height/3)

	m.clampPreviewScroll()
}
