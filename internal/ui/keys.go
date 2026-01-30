package ui

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Up          key.Binding
	Down        key.Binding
	Left        key.Binding
	Right       key.Binding
	Enter       key.Binding
	Delete      key.Binding
	Restore     key.Binding
	Rename      key.Binding
	Reassign    key.Binding
	ReassignAll key.Binding
	Search      key.Binding
	DeepSearch  key.Binding
	NextMatch   key.Binding
	PrevMatch   key.Binding
	Tab         key.Binding
	Clear       key.Binding
	Quit        key.Binding
	Help        key.Binding
	Escape      key.Binding
	Confirm     key.Binding
	HalfUp      key.Binding
	HalfDown    key.Binding
	Top         key.Binding
	Bottom      key.Binding
	Preview     key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j", "move down"),
		),
		Left: key.NewBinding(
			key.WithKeys("h", "left"),
			key.WithHelp("h", "previous tab"),
		),
		Right: key.NewBinding(
			key.WithKeys("l", "right"),
			key.WithHelp("l", "next tab"),
		),
		HalfUp: key.NewBinding(
			key.WithKeys("ctrl+u"),
			key.WithHelp("ctrl+u", "page up"),
		),
		HalfDown: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("ctrl+d", "page down"),
		),
		Top: key.NewBinding(
			key.WithKeys("g", "home"),
			key.WithHelp("gg", "jump to top"),
		),
		Bottom: key.NewBinding(
			key.WithKeys("G", "end"),
			key.WithHelp("G", "jump to bottom"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("return", "select"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d", "x"),
			key.WithHelp("d", "move to bin"),
		),
		Restore: key.NewBinding(
			key.WithKeys("u"),
			key.WithHelp("u", "restore"),
		),
		Rename: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "rename"),
		),
		Reassign: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "reassign folder"),
		),
		ReassignAll: key.NewBinding(
			key.WithKeys("R"),
			key.WithHelp("R", "reassign all"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		DeepSearch: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "search"),
		),
		NextMatch: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "next match"),
		),
		PrevMatch: key.NewBinding(
			key.WithKeys("N"),
			key.WithHelp("N", "previous match"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch focus"),
		),
		Clear: key.NewBinding(
			key.WithKeys("D"),
			key.WithHelp("D", "empty bin"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("y", "Y"),
			key.WithHelp("y", "confirm"),
		),
		Preview: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "toggle preview"),
		),
	}
}
