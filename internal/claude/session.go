package claude

import (
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
			continue
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
					continue
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
