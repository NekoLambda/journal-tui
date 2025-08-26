package components

import (
	"github.com/NekoLambda/journal-tui/ui/styles"
	"github.com/charmbracelet/bubbles/textinput"
)

type Input struct {
	textinput.Model
	Label string
}

func NewInput(label string) Input {
	ti := textinput.New()
	ti.Placeholder = label
	ti.Width = 40
	ti.Focus()

	return Input{
		Model: ti,
		Label: label,
	}
}

func (i Input) View() string {
	return styles.InputStyle.Render(i.Label + ": " + i.Model.View())
}
