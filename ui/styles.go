package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Base styles
	BaseStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F8F8F2")).
		Background(lipgloss.Color("#282A36"))

	// Text styles
	HeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF79C6"))

	SelectedStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#50FA7B"))

	NormalStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F8F8F2"))

	HelpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6272A4"))

	// Input/Form styles
	InputStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Padding(0, 1)

	// Modal styles
	ModalStyle = lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		Padding(1, 2).
		BorderForeground(lipgloss.Color("#BD93F9"))

	// Preview styles
	PreviewStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1).
		BorderForeground(lipgloss.Color("#6272A4"))
)
