package model

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	choices []string // later: entries from storage
	cursor  int
}

func New() Model {
	return Model{
		choices: []string{"Write new journal", "View past entries", "Quit"},
	}
}

// Init runs on start
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages (keyboard input, timers, etc.)
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		}
	}
	return m, nil
}

// View is the UI
func (m Model) View() string {
	s := "What do you want to do?\n\n"
	for i, choice := range m.choices {
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // current cursor
		}
		s += cursor + " " + choice + "\n"
	}
	s += "\nPress q to quit.\n"
	return s
}
