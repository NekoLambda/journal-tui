package main

import (
	"log"

	"github.com/NekoLambda/journal-tui/internal/model"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(model.New())
	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}

// Model represents the state of your TUI
type Model struct{}

// New returns a new instance of Model
func New() Model {
	return Model{}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

// View implements tea.Model
func (m Model) View() string {
	return "Hello, Journal-TUI!"
}
