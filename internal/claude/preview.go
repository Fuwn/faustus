package claude

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
)

type RawMessage struct {
	Type    string          `json:"type"`
	Message json.RawMessage `json:"message"`
}

type UserMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AssistantMessage struct {
	Role    string         `json:"role"`
	Content []ContentBlock `json:"content"`
}

type ContentBlock struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Thinking string `json:"thinking,omitempty"`
	Name     string `json:"name,omitempty"`
	Input    any    `json:"input,omitempty"`
}

type PreviewContent struct {
	Messages []PreviewMessage
	Error    string
}

type PreviewMessage struct {
	Role    string
	Content string
}

func LoadSessionPreview(session *Session, maxMessages int) PreviewContent {
	if session == nil || session.FullPath == "" {
		return PreviewContent{Error: "No session selected"}
	}

	file, openError := os.Open(session.FullPath)

	if openError != nil {
		return PreviewContent{Error: "Could not open session file"}
	}

	defer func() { _ = file.Close() }()

	var messages []PreviewMessage

	scanner := bufio.NewScanner(file)
	scanBuffer := make([]byte, 0, 64*1024)

	scanner.Buffer(scanBuffer, 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			continue
		}

		var rawMessage RawMessage

		if unmarshalError := json.Unmarshal([]byte(line), &rawMessage); unmarshalError != nil {
			continue
		}

		parsedMessages := parseRawMessage(rawMessage)

		messages = append(messages, parsedMessages...)
	}

	if len(messages) > maxMessages {
		messages = messages[len(messages)-maxMessages:]
	}

	if len(messages) == 0 {
		return PreviewContent{Error: "No messages in session"}
	}

	return PreviewContent{Messages: messages}
}

func parseRawMessage(rawMessage RawMessage) []PreviewMessage {
	var result []PreviewMessage

	switch rawMessage.Type {
	case "user":
		var userMessage UserMessage

		if unmarshalError := json.Unmarshal(rawMessage.Message, &userMessage); unmarshalError != nil {
			return nil
		}

		content := userMessage.Content

		if len(content) > 500 {
			content = content[:500] + " …"
		}

		if content != "" {
			result = append(result, PreviewMessage{Role: "user", Content: content})
		}

	case "assistant":
		var assistantMessage AssistantMessage

		if unmarshalError := json.Unmarshal(rawMessage.Message, &assistantMessage); unmarshalError != nil {
			return nil
		}

		for _, contentBlock := range assistantMessage.Content {
			switch contentBlock.Type {
			case "text":
				text := contentBlock.Text

				if len(text) > 500 {
					text = text[:500] + " …"
				}

				if text != "" {
					result = append(result, PreviewMessage{Role: "assistant", Content: text})
				}

			case "tool_use":
				toolInfo := contentBlock.Name

				if toolInfo != "" {
					if inputMap, isMap := contentBlock.Input.(map[string]interface{}); isMap {
						if command, exists := inputMap["command"]; exists {
							if commandString, isString := command.(string); isString {
								if len(commandString) > 60 {
									commandString = commandString[:60] + " …"
								}

								toolInfo += ": " + commandString
							}
						} else if pattern, exists := inputMap["pattern"]; exists {
							if patternString, isString := pattern.(string); isString {
								toolInfo += ": " + patternString
							}
						} else if filePath, exists := inputMap["file_path"]; exists {
							if filePathString, isString := filePath.(string); isString {
								pathParts := strings.Split(filePathString, "/")

								toolInfo += ": " + pathParts[len(pathParts)-1]
							}
						}
					}

					result = append(result, PreviewMessage{Role: "tool", Content: toolInfo})
				}

			case "thinking":
				thinking := contentBlock.Thinking

				if len(thinking) > 200 {
					thinking = thinking[:200] + " …"
				}

				if thinking != "" {
					result = append(result, PreviewMessage{Role: "thinking", Content: thinking})
				}
			}
		}
	}

	return result
}
