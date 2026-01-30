package main

import (
	"fmt"
	"github.com/Fuwn/faustus/internal/app"
	"github.com/Fuwn/faustus/internal/claude"
	tea "github.com/charmbracelet/bubbletea"
	"os"
)

func main() {
	sessions, err := claude.LoadAllSessions()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading sessions: %v\n", err)
		os.Exit(1)
	}

	m := app.NewModel(sessions)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
