package app

import (
	"fmt"
	"strings"
	"time"
)

func truncate(text string, maxLength int) string {
	if maxLength <= 0 {
		return ""
	}

	text = strings.ReplaceAll(text, "\n", " ")

	if len(text) <= maxLength {
		return text
	}

	if maxLength <= 2 {
		return text[:maxLength]
	}

	return text[:maxLength-2] + " …"
}

func formatTime(timestamp time.Time) string {
	now := time.Now()
	difference := now.Sub(timestamp)

	switch {
	case difference < time.Minute:
		return "just now"
	case difference < time.Hour:
		return fmt.Sprintf("%dm ago", int(difference.Minutes()))
	case difference < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(difference.Hours()))
	case difference < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(difference.Hours()/24))
	default:
		return timestamp.Format("Jan 2")
	}
}

func max(first, second int) int {
	if first > second {
		return first
	}

	return second
}

func min(first, second int) int {
	if first < second {
		return first
	}

	return second
}

func highlightMatches(text, query string) string {
	if query == "" {
		return text
	}

	queryLower := strings.ToLower(query)
	textLower := strings.ToLower(text)

	var result strings.Builder

	lastEnd := 0

	for {
		index := strings.Index(textLower[lastEnd:], queryLower)

		if index == -1 {
			result.WriteString(text[lastEnd:])

			break
		}

		matchStart := lastEnd + index

		result.WriteString(text[lastEnd:matchStart])

		matchEnd := matchStart + len(query)

		result.WriteString("\033[43;30m")
		result.WriteString(text[matchStart:matchEnd])
		result.WriteString("\033[0m")

		lastEnd = matchEnd
	}

	return result.String()
}

func wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}

	var result strings.Builder

	words := strings.Fields(text)
	lineLength := 0

	for wordIndex, word := range words {
		wordLength := len(word)

		if lineLength+wordLength+1 > width && lineLength > 0 {
			result.WriteString("\n")

			lineLength = 0
		}

		if lineLength > 0 {
			result.WriteString(" ")

			lineLength += 1
		}

		result.WriteString(word)

		lineLength += wordLength

		if lineLength > width && wordIndex < len(words)-1 {
			result.WriteString("\n")

			lineLength = 0
		}
	}

	return result.String()
}
