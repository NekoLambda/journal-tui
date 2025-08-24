package main

import (
	"fmt"

	"github.com/NekoLambda/journal-tui/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(ui.NewModel())
	if err := p.Start(); err != nil {
		fmt.Println("Error starting app:", err)
		return
	}
}
