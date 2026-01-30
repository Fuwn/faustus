package app

import (
	"time"

	"github.com/Fuwn/faustus/internal/claude"
	"github.com/Fuwn/faustus/internal/ui"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Tab int

const (
	TabSessions Tab = iota
	TabTrash
)

type Mode int

const (
	ModeNormal Mode = iota
	ModeSearch
	ModeDeepSearch
	ModeRename
	ModeConfirm
	ModeReassign
)

type ConfirmAction int

const (
	ConfirmNone ConfirmAction = iota
	ConfirmDelete
	ConfirmRestore
	ConfirmEmptyTrash
	ConfirmPermanentDelete
)

type Model struct {
	sessions             []claude.Session
	filtered             []claude.Session
	cursor               int
	offset               int
	width                int
	height               int
	tab                  Tab
	mode                 Mode
	confirmAction        ConfirmAction
	searchInput          textinput.Model
	renameInput          textinput.Model
	keys                 ui.KeyMap
	showHelp             bool
	message              string
	messageTime          time.Time
	showPreview          bool
	previewFocus         bool
	previewScroll        int
	previewCache         *claude.PreviewContent
	previewFor           string
	deepSearchInput      textinput.Model
	deepSearchResults    []claude.SearchResult
	deepSearchIndex      int
	deepSearchQuery      string
	previewSearchQuery   string
	previewSearchMatches []int
	previewSearchIndex   int
	reassignInput        textinput.Model
	reassignAll          bool
}

func NewModel(sessions []claude.Session) Model {
	searchInput := textinput.New()

	searchInput.Placeholder = "Filter sessions"
	searchInput.CharLimit = 100
	searchInput.Width = 40

	renameInput := textinput.New()

	renameInput.Placeholder = "Enter new name"
	renameInput.CharLimit = 200
	renameInput.Width = 60

	deepSearchInput := textinput.New()

	deepSearchInput.Placeholder = "Search all sessions"
	deepSearchInput.CharLimit = 100
	deepSearchInput.Width = 50

	reassignInput := textinput.New()

	reassignInput.Placeholder = "Enter new project path"
	reassignInput.CharLimit = 500
	reassignInput.Width = 80

	model := Model{
		sessions:        sessions,
		keys:            ui.DefaultKeyMap(),
		searchInput:     searchInput,
		renameInput:     renameInput,
		deepSearchInput: deepSearchInput,
		reassignInput:   reassignInput,
		showPreview:     false,
	}

	model.updateFiltered()

	return model
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}
