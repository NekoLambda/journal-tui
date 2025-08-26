package components

import (
	"github.com/NekoLambda/journal-tui/ui/styles"
	"github.com/charmbracelet/lipgloss"
)

type ModalType int

const (
	ModalConfirm ModalType = iota
	ModalPrompt
	ModalInfo
)

type Modal struct {
	Type    ModalType
	Title   string
	Content string
	Visible bool
	Result  chan bool // For async response handling
}

func NewModal(modalType ModalType, title, content string) *Modal {
	return &Modal{
		Type:    modalType,
		Title:   title,
		Content: content,
		Result:  make(chan bool, 1),
	}
}

func (m *Modal) Show() {
	m.Visible = true
}

func (m *Modal) Hide() {
	m.Visible = false
}

func (m Modal) View() string {
	if !m.Visible {
		return ""
	}

	var hint string
	switch m.Type {
	case ModalConfirm:
		hint = "(y/n)"
	case ModalPrompt:
		hint = "(enter to confirm, esc to cancel)"
	case ModalInfo:
		hint = "(press any key)"
	}

	content := m.Content
	if hint != "" {
		content += "\n\n" + styles.HelpStyle.Render(hint)
	}

	box := styles.ModalStyle.Render(m.Title + "\n\n" + content)
	return lipgloss.Place(60, 20, lipgloss.Center, lipgloss.Center, box)
}
