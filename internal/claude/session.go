package claude

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Session struct {
	SessionID    string    `json:"sessionId"`
	FullPath     string    `json:"fullPath"`
	FirstPrompt  string    `json:"firstPrompt"`
	Summary      string    `json:"summary"`
	MessageCount int       `json:"messageCount"`
	Created      time.Time `json:"created"`
	Modified     time.Time `json:"modified"`
	GitBranch    string    `json:"gitBranch"`
	ProjectPath  string    `json:"projectPath"`
	IsSidechain  bool      `json:"isSidechain"`

	ProjectName string `json:"-"`
	InTrash     bool   `json:"-"`
}

type SessionIndex struct {
	Version      int       `json:"version"`
	Entries      []Session `json:"entries"`
	OriginalPath string    `json:"originalPath"`
}

func ClaudeDir() string {
	homeDirectory, _ := os.UserHomeDir()

	return filepath.Join(homeDirectory, ".claude")
}

func ProjectsDir() string {
	return filepath.Join(ClaudeDir(), "projects")
}

func TrashDir() string {
	return filepath.Join(ClaudeDir(), "faustus-trash")
}

func EnsureTrashDir() error {
	return os.MkdirAll(TrashDir(), 0o755)
}

func LoadAllSessions() ([]Session, error) {
	var allSessions []Session

	projectsDirectory := ProjectsDir()
	directoryEntries, readError := os.ReadDir(projectsDirectory)

	if readError != nil {
		return nil, readError
	}

	for _, directoryEntry := range directoryEntries {
		if !directoryEntry.IsDir() {
			continue
		}

		projectDirectory := filepath.Join(projectsDirectory, directoryEntry.Name())
		indexPath := filepath.Join(projectDirectory, "sessions-index.json")
		sessions, loadError := loadSessionsFromIndex(indexPath, directoryEntry.Name(), false)

		if loadError != nil {
			sessions = loadSessionsFromJsonlFiles(projectDirectory, directoryEntry.Name(), false)
		}

		allSessions = append(allSessions, sessions...)
	}

	trashDirectory := TrashDir()

	if _, statError := os.Stat(trashDirectory); statError == nil {
		trashEntries, readError := os.ReadDir(trashDirectory)

		if readError == nil {
			for _, directoryEntry := range trashEntries {
				if !directoryEntry.IsDir() {
					continue
				}

				projectDirectory := filepath.Join(trashDirectory, directoryEntry.Name())
				indexPath := filepath.Join(projectDirectory, "sessions-index.json")
				sessions, loadError := loadSessionsFromIndex(indexPath, directoryEntry.Name(), true)

				if loadError != nil {
					sessions = loadSessionsFromJsonlFiles(projectDirectory, directoryEntry.Name(), true)
				}

				allSessions = append(allSessions, sessions...)
			}
		}
	}

	sort.Slice(allSessions, func(first, second int) bool {
		return allSessions[first].Modified.After(allSessions[second].Modified)
	})

	return allSessions, nil
}

func loadSessionsFromIndex(indexPath, projectDirectoryName string, inTrash bool) ([]Session, error) {
	fileData, readError := os.ReadFile(indexPath)

	if readError != nil {
		return nil, readError
	}

	var sessionIndex SessionIndex

	if unmarshalError := json.Unmarshal(fileData, &sessionIndex); unmarshalError != nil {
		return nil, unmarshalError
	}

	projectName := deriveProjectName(projectDirectoryName)

	for entryIndex := range sessionIndex.Entries {
		sessionIndex.Entries[entryIndex].ProjectName = projectName
		sessionIndex.Entries[entryIndex].InTrash = inTrash

		if inTrash {
			sessionIndex.Entries[entryIndex].FullPath = filepath.Join(TrashDir(), projectDirectoryName,
				sessionIndex.Entries[entryIndex].SessionID+".jsonl")
		}
	}

	return sessionIndex.Entries, nil
}

type jsonlFirstLine struct {
	SessionID   string    `json:"sessionId"`
	Cwd         string    `json:"cwd"`
	GitBranch   string    `json:"gitBranch"`
	Timestamp   time.Time `json:"timestamp"`
	IsSidechain bool      `json:"isSidechain"`
	Message     struct {
		Role    string `json:"role"`
		Content any    `json:"content"`
	} `json:"message"`
}

func loadSessionsFromJsonlFiles(projectDirectory, projectDirectoryName string, inTrash bool) []Session {
	var sessions []Session

	entries, readError := os.ReadDir(projectDirectory)

	if readError != nil {
		return sessions
	}

	projectName := deriveProjectName(projectDirectoryName)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}

		fullPath := filepath.Join(projectDirectory, entry.Name())
		session := parseSessionFromJsonl(fullPath, projectName, inTrash)

		if session != nil {
			sessions = append(sessions, *session)
		}
	}

	return sessions
}

func parseSessionFromJsonl(filePath, projectName string, inTrash bool) *Session {
	file, openError := os.Open(filePath)

	if openError != nil {
		return nil
	}

	defer func() { _ = file.Close() }()

	fileInfo, statError := file.Stat()

	if statError != nil {
		return nil
	}

	scanner := bufio.NewScanner(file)
	scanBuffer := make([]byte, 0, 64*1024)

	scanner.Buffer(scanBuffer, 10*1024*1024)

	var firstLine jsonlFirstLine
	var firstUserContent string
	var messageCount int
	var created time.Time

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			continue
		}

		var raw struct {
			Type      string    `json:"type"`
			Timestamp time.Time `json:"timestamp"`
		}

		if unmarshalError := json.Unmarshal([]byte(line), &raw); unmarshalError != nil {
			continue
		}

		if raw.Type == "user" || raw.Type == "assistant" {
			messageCount += 1
		}

		if raw.Type == "user" && firstUserContent == "" {
			var userLine jsonlFirstLine

			if unmarshalError := json.Unmarshal([]byte(line), &userLine); unmarshalError == nil {
				if created.IsZero() {
					created = userLine.Timestamp
				}

				if firstLine.SessionID == "" {
					firstLine = userLine
				}

				switch content := userLine.Message.Content.(type) {
				case string:
					if !userLine.IsSidechain {
						firstUserContent = content
					}
				case []any:
					for _, block := range content {
						if blockMap, ok := block.(map[string]any); ok {
							if text, ok := blockMap["text"].(string); ok {
								if !userLine.IsSidechain {
									firstUserContent = text

									break
								}
							}
						}
					}
				}
			}
		}
	}

	if firstLine.SessionID == "" {
		return nil
	}

	sessionID := strings.TrimSuffix(filepath.Base(filePath), ".jsonl")

	return &Session{
		SessionID:    sessionID,
		FullPath:     filePath,
		FirstPrompt:  truncateString(firstUserContent, 200),
		Summary:      "",
		MessageCount: messageCount,
		Created:      created,
		Modified:     fileInfo.ModTime(),
		GitBranch:    firstLine.GitBranch,
		ProjectPath:  firstLine.Cwd,
		IsSidechain:  firstLine.IsSidechain,
		ProjectName:  projectName,
		InTrash:      inTrash,
	}
}

func truncateString(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}

	return text[:maxLength-1] + "…"
}

func deriveProjectName(directoryName string) string {
	parts := strings.Split(directoryName, "-")

	if len(parts) > 0 {
		for partIndex := len(parts) - 1; partIndex >= 0; partIndex-- {
			if parts[partIndex] != "" {
				if partIndex > 0 && len(parts[partIndex-1]) > 0 {
					return parts[partIndex-1] + "/" + parts[partIndex]
				}

				return parts[partIndex]
			}
		}
	}

	return directoryName
}

func ProjectDir(session *Session) string {
	return filepath.Dir(session.FullPath)
}

func MoveToTrash(session *Session) error {
	if session.InTrash {
		return nil
	}

	if ensureError := EnsureTrashDir(); ensureError != nil {
		return ensureError
	}

	sourceProjectDirectory := ProjectDir(session)
	projectDirectoryName := filepath.Base(sourceProjectDirectory)
	destinationProjectDirectory := filepath.Join(TrashDir(), projectDirectoryName)

	if mkdirError := os.MkdirAll(destinationProjectDirectory, 0o755); mkdirError != nil {
		return mkdirError
	}

	sourceFile := session.FullPath
	destinationFile := filepath.Join(destinationProjectDirectory, session.SessionID+".jsonl")

	if renameError := os.Rename(sourceFile, destinationFile); renameError != nil {
		return renameError
	}

	sourceAssociatedDirectory := filepath.Join(sourceProjectDirectory, session.SessionID)

	if _, statError := os.Stat(sourceAssociatedDirectory); statError == nil {
		destinationAssociatedDirectory := filepath.Join(destinationProjectDirectory, session.SessionID)

		_ = os.Rename(sourceAssociatedDirectory, destinationAssociatedDirectory)
	}

	if removeError := removeFromIndex(sourceProjectDirectory, session.SessionID); removeError != nil {
		return removeError
	}

	session.InTrash = true
	session.FullPath = destinationFile

	return addToIndex(destinationProjectDirectory, session)
}

func RestoreFromTrash(session *Session) error {
	if !session.InTrash {
		return nil
	}

	sourceProjectDirectory := ProjectDir(session)
	projectDirectoryName := filepath.Base(sourceProjectDirectory)
	destinationProjectDirectory := filepath.Join(ProjectsDir(), projectDirectoryName)

	if mkdirError := os.MkdirAll(destinationProjectDirectory, 0o755); mkdirError != nil {
		return mkdirError
	}

	sourceFile := session.FullPath
	destinationFile := filepath.Join(destinationProjectDirectory, session.SessionID+".jsonl")

	if renameError := os.Rename(sourceFile, destinationFile); renameError != nil {
		return renameError
	}

	sourceAssociatedDirectory := filepath.Join(sourceProjectDirectory, session.SessionID)

	if _, statError := os.Stat(sourceAssociatedDirectory); statError == nil {
		destinationAssociatedDirectory := filepath.Join(destinationProjectDirectory, session.SessionID)

		_ = os.Rename(sourceAssociatedDirectory, destinationAssociatedDirectory)
	}

	if removeError := removeFromIndex(sourceProjectDirectory, session.SessionID); removeError != nil {
		return removeError
	}

	session.InTrash = false
	session.FullPath = destinationFile

	return addToIndex(destinationProjectDirectory, session)
}

func PermanentlyDelete(session *Session) error {
	projectDirectory := ProjectDir(session)

	if removeError := os.Remove(session.FullPath); removeError != nil && !os.IsNotExist(removeError) {
		return removeError
	}

	associatedDirectory := filepath.Join(projectDirectory, session.SessionID)

	_ = os.RemoveAll(associatedDirectory)

	return removeFromIndex(projectDirectory, session.SessionID)
}

func EmptyTrash() error {
	trashDirectory := TrashDir()

	if _, statError := os.Stat(trashDirectory); os.IsNotExist(statError) {
		return nil
	}

	return os.RemoveAll(trashDirectory)
}

func RenameSession(session *Session, newSummary string) error {
	projectDirectory := ProjectDir(session)
	indexPath := filepath.Join(projectDirectory, "sessions-index.json")
	fileData, readError := os.ReadFile(indexPath)

	if readError != nil {
		return readError
	}

	var sessionIndex SessionIndex

	if unmarshalError := json.Unmarshal(fileData, &sessionIndex); unmarshalError != nil {
		return unmarshalError
	}

	for entryIndex := range sessionIndex.Entries {
		if sessionIndex.Entries[entryIndex].SessionID == session.SessionID {
			sessionIndex.Entries[entryIndex].Summary = newSummary

			break
		}
	}

	return writeIndex(indexPath, &sessionIndex)
}

func removeFromIndex(projectDirectory, sessionID string) error {
	indexPath := filepath.Join(projectDirectory, "sessions-index.json")
	fileData, readError := os.ReadFile(indexPath)

	if readError != nil {
		if os.IsNotExist(readError) {
			return nil
		}

		return readError
	}

	var sessionIndex SessionIndex

	if unmarshalError := json.Unmarshal(fileData, &sessionIndex); unmarshalError != nil {
		return unmarshalError
	}

	filteredEntries := make([]Session, 0, len(sessionIndex.Entries)-1)

	for _, entry := range sessionIndex.Entries {
		if entry.SessionID != sessionID {
			filteredEntries = append(filteredEntries, entry)
		}
	}

	sessionIndex.Entries = filteredEntries

	return writeIndex(indexPath, &sessionIndex)
}

func addToIndex(projectDirectory string, session *Session) error {
	indexPath := filepath.Join(projectDirectory, "sessions-index.json")

	var sessionIndex SessionIndex

	fileData, readError := os.ReadFile(indexPath)

	if readError != nil {
		if os.IsNotExist(readError) {
			sessionIndex = SessionIndex{
				Version:      1,
				Entries:      []Session{},
				OriginalPath: session.ProjectPath,
			}
		} else {
			return readError
		}
	} else {
		if unmarshalError := json.Unmarshal(fileData, &sessionIndex); unmarshalError != nil {
			return unmarshalError
		}
	}

	found := false

	for entryIndex := range sessionIndex.Entries {
		if sessionIndex.Entries[entryIndex].SessionID == session.SessionID {
			sessionIndex.Entries[entryIndex] = *session
			found = true

			break
		}
	}

	if !found {
		sessionIndex.Entries = append(sessionIndex.Entries, *session)
	}

	return writeIndex(indexPath, &sessionIndex)
}

func writeIndex(indexPath string, sessionIndex *SessionIndex) error {
	jsonData, marshalError := json.MarshalIndent(sessionIndex, "", "  ")

	if marshalError != nil {
		return marshalError
	}

	return os.WriteFile(indexPath, jsonData, 0o644)
}

func ReassignSessionPath(session *Session, newPath string) error {
	oldProjectDirectory := ProjectDir(session)
	newDirectoryName := pathToDirectoryName(newPath)

	var newProjectDirectory string

	if session.InTrash {
		newProjectDirectory = filepath.Join(TrashDir(), newDirectoryName)
	} else {
		newProjectDirectory = filepath.Join(ProjectsDir(), newDirectoryName)
	}

	if oldProjectDirectory == newProjectDirectory {
		return nil
	}

	if mkdirError := os.MkdirAll(newProjectDirectory, 0o755); mkdirError != nil {
		return mkdirError
	}

	oldJsonlPath := session.FullPath
	newJsonlPath := filepath.Join(newProjectDirectory, filepath.Base(oldJsonlPath))

	if moveError := os.Rename(oldJsonlPath, newJsonlPath); moveError != nil {
		return moveError
	}

	if updateError := updateJsonlCwd(newJsonlPath, newPath); updateError != nil {
		_ = os.Rename(newJsonlPath, oldJsonlPath)

		return updateError
	}

	oldIndexPath := filepath.Join(oldProjectDirectory, "sessions-index.json")

	_ = removeFromIndex(oldProjectDirectory, session.SessionID)

	newIndexPath := filepath.Join(newProjectDirectory, "sessions-index.json")

	session.FullPath = newJsonlPath
	session.ProjectPath = newPath

	_ = addToIndexWithPath(newIndexPath, session, newPath)

	if isEmpty, _ := isDirectoryEmpty(oldProjectDirectory); isEmpty {
		_ = os.Remove(oldIndexPath)
		_ = os.Remove(oldProjectDirectory)
	}

	return nil
}

func ReassignProjectPath(oldPath, newPath string) (int, error) {
	var updatedCount int

	projectsDirectory := ProjectsDir()
	directoryEntries, readError := os.ReadDir(projectsDirectory)

	if readError != nil {
		return 0, readError
	}

	for _, directoryEntry := range directoryEntries {
		if !directoryEntry.IsDir() {
			continue
		}

		projectDirectory := filepath.Join(projectsDirectory, directoryEntry.Name())
		count, updateError := reassignInProject(projectDirectory, oldPath, newPath, false)

		if updateError != nil {
			return updatedCount, updateError
		}

		updatedCount += count
	}

	trashDirectory := TrashDir()

	if _, statError := os.Stat(trashDirectory); statError == nil {
		trashEntries, readError := os.ReadDir(trashDirectory)

		if readError == nil {
			for _, directoryEntry := range trashEntries {
				if !directoryEntry.IsDir() {
					continue
				}

				projectDirectory := filepath.Join(trashDirectory, directoryEntry.Name())
				count, updateError := reassignInProject(projectDirectory, oldPath, newPath, true)

				if updateError != nil {
					return updatedCount, updateError
				}

				updatedCount += count
			}
		}
	}

	return updatedCount, nil
}

func reassignInProject(projectDirectory, oldPath, newPath string, inTrash bool) (int, error) {
	indexPath := filepath.Join(projectDirectory, "sessions-index.json")
	fileData, readError := os.ReadFile(indexPath)

	if readError != nil {
		if os.IsNotExist(readError) {
			return reassignOrphanedSessions(projectDirectory, oldPath, newPath, inTrash)
		}

		return 0, nil
	}

	var sessionIndex SessionIndex

	if unmarshalError := json.Unmarshal(fileData, &sessionIndex); unmarshalError != nil {
		return 0, nil
	}

	if sessionIndex.OriginalPath != oldPath {
		hasMatchingSessions := false

		for _, entry := range sessionIndex.Entries {
			if entry.ProjectPath == oldPath {
				hasMatchingSessions = true

				break
			}
		}

		if !hasMatchingSessions {
			return 0, nil
		}
	}

	var updatedCount int

	for _, entry := range sessionIndex.Entries {
		if entry.ProjectPath == oldPath {
			entry.InTrash = inTrash

			if reassignError := ReassignSessionPath(&entry, newPath); reassignError != nil {
				continue
			}

			updatedCount += 1
		}
	}

	if isEmpty, _ := isDirectoryEmpty(projectDirectory); isEmpty {
		_ = os.Remove(indexPath)
		_ = os.Remove(projectDirectory)
	}

	return updatedCount, nil
}

func reassignOrphanedSessions(projectDirectory, oldPath, newPath string, inTrash bool) (int, error) {
	entries, readError := os.ReadDir(projectDirectory)

	if readError != nil {
		return 0, nil
	}

	var updatedCount int

	projectName := deriveProjectName(filepath.Base(projectDirectory))

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}

		fullPath := filepath.Join(projectDirectory, entry.Name())
		currentPath := getJsonlProjectPath(fullPath)

		if currentPath != oldPath {
			continue
		}

		session := parseSessionFromJsonl(fullPath, projectName, inTrash)

		if session == nil {
			continue
		}

		if reassignError := ReassignSessionPath(session, newPath); reassignError != nil {
			continue
		}

		updatedCount += 1
	}

	if isEmpty, _ := isDirectoryEmpty(projectDirectory); isEmpty {
		_ = os.Remove(projectDirectory)
	}

	return updatedCount, nil
}

func getJsonlProjectPath(filePath string) string {
	file, openError := os.Open(filePath)

	if openError != nil {
		return ""
	}

	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	scanBuffer := make([]byte, 0, 64*1024)

	scanner.Buffer(scanBuffer, 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			continue
		}

		var lineData struct {
			Type string `json:"type"`
			Cwd  string `json:"cwd"`
		}

		if unmarshalError := json.Unmarshal([]byte(line), &lineData); unmarshalError != nil {
			continue
		}

		if lineData.Type == "user" && lineData.Cwd != "" {
			return lineData.Cwd
		}
	}

	return ""
}

func updateJsonlProjectPath(filePath, newPath string) error {
	fileData, readError := os.ReadFile(filePath)

	if readError != nil {
		return readError
	}

	lines := strings.Split(string(fileData), "\n")
	var updatedLines []string

	for _, line := range lines {
		if line == "" {
			updatedLines = append(updatedLines, line)

			continue
		}

		var lineData map[string]any

		if unmarshalError := json.Unmarshal([]byte(line), &lineData); unmarshalError != nil {
			updatedLines = append(updatedLines, line)

			continue
		}

		if _, hasCwd := lineData["cwd"]; hasCwd {
			lineData["cwd"] = newPath
		}

		updatedLine, marshalError := json.Marshal(lineData)

		if marshalError != nil {
			updatedLines = append(updatedLines, line)

			continue
		}

		updatedLines = append(updatedLines, string(updatedLine))
	}

	tempPath := filePath + ".tmp"

	if writeError := os.WriteFile(tempPath, []byte(strings.Join(updatedLines, "\n")), 0o644); writeError != nil {
		return writeError
	}

	return os.Rename(tempPath, filePath)
}

func pathToDirectoryName(projectPath string) string {
	return strings.ReplaceAll(projectPath, "/", "-")
}

func updateJsonlCwd(filePath, newPath string) error {
	return updateJsonlProjectPath(filePath, newPath)
}

func addToIndexWithPath(indexPath string, session *Session, originalPath string) error {
	var sessionIndex SessionIndex

	fileData, readError := os.ReadFile(indexPath)

	if readError != nil {
		if os.IsNotExist(readError) {
			sessionIndex = SessionIndex{
				Version:      1,
				Entries:      []Session{},
				OriginalPath: originalPath,
			}
		} else {
			return readError
		}
	} else {
		if unmarshalError := json.Unmarshal(fileData, &sessionIndex); unmarshalError != nil {
			return unmarshalError
		}
	}

	for _, entry := range sessionIndex.Entries {
		if entry.SessionID == session.SessionID {
			return nil
		}
	}

	sessionIndex.Entries = append(sessionIndex.Entries, *session)

	return writeIndex(indexPath, &sessionIndex)
}

func isDirectoryEmpty(path string) (bool, error) {
	entries, readError := os.ReadDir(path)

	if readError != nil {
		return false, readError
	}

	return len(entries) == 0, nil
}
