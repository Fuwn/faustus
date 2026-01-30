package claude

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
)

type SearchResult struct {
	Session       *Session
	MessageIndex  int
	Role          string
	Content       string
	MatchPosition int
}

func SearchAllSessions(sessions []Session, query string) []SearchResult {
	if query == "" {
		return nil
	}

	query = strings.ToLower(query)

	var results []SearchResult

	for sessionIndex := range sessions {
		session := &sessions[sessionIndex]
		matches := searchSession(session, query)

		results = append(results, matches...)
	}

	return results
}

func searchSession(session *Session, query string) []SearchResult {
	file, openError := os.Open(session.FullPath)

	if openError != nil {
		return nil
	}

	defer func() { _ = file.Close() }()

	var results []SearchResult

	scanner := bufio.NewScanner(file)
	scanBuffer := make([]byte, 0, 64*1024)

	scanner.Buffer(scanBuffer, 10*1024*1024)

	messageIndex := 0

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			continue
		}

		var rawMessage RawMessage

		if unmarshalError := json.Unmarshal([]byte(line), &rawMessage); unmarshalError != nil {
			continue
		}

		matches := searchRawMessage(session, &rawMessage, query, messageIndex)

		results = append(results, matches...)

		if rawMessage.Type == "user" || rawMessage.Type == "assistant" {
			messageIndex += 1
		}
	}

	return results
}

func searchRawMessage(session *Session, rawMessage *RawMessage, query string, messageIndex int) []SearchResult {
	var results []SearchResult

	switch rawMessage.Type {
	case "user":
		var userMessage UserMessage

		if unmarshalError := json.Unmarshal(rawMessage.Message, &userMessage); unmarshalError != nil {
			return nil
		}

		contentLowercase := strings.ToLower(userMessage.Content)

		if matchPosition := strings.Index(contentLowercase, query); matchPosition != -1 {
			content := matchContext(userMessage.Content, matchPosition, len(query))

			results = append(results, SearchResult{
				Session:       session,
				MessageIndex:  messageIndex,
				Role:          "user",
				Content:       content,
				MatchPosition: matchPosition,
			})
		}

	case "assistant":
		var assistantMessage AssistantMessage

		if unmarshalError := json.Unmarshal(rawMessage.Message, &assistantMessage); unmarshalError != nil {
			return nil
		}

		for _, contentBlock := range assistantMessage.Content {
			if contentBlock.Type == "text" {
				textLowercase := strings.ToLower(contentBlock.Text)

				if matchPosition := strings.Index(textLowercase, query); matchPosition != -1 {
					content := matchContext(contentBlock.Text, matchPosition, len(query))

					results = append(results, SearchResult{
						Session:       session,
						MessageIndex:  messageIndex,
						Role:          "assistant",
						Content:       content,
						MatchPosition: matchPosition,
					})
				}
			}
		}
	}

	return results
}

func matchContext(text string, matchPosition, matchLength int) string {
	const contextLength = 100

	start := matchPosition - contextLength/2

	if start < 0 {
		start = 0
	}

	end := matchPosition + matchLength + contextLength/2

	if end > len(text) {
		end = len(text)
	}

	result := text[start:end]
	result = strings.ReplaceAll(result, "\n", " ")
	result = strings.Join(strings.Fields(result), " ")

	if start > 0 {
		result = "… " + result
	}

	if end < len(text) {
		result = result + " …"
	}

	return result
}

func SearchPreview(preview *PreviewContent, query string) []int {
	if preview == nil || query == "" {
		return nil
	}

	query = strings.ToLower(query)

	var matches []int

	for messageIndex, previewMessage := range preview.Messages {
		if strings.Contains(strings.ToLower(previewMessage.Content), query) {
			matches = append(matches, messageIndex)
		}
	}

	return matches
}
