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
