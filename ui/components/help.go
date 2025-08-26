package components

import (
	"github.com/NekoLambda/journal-tui/ui/styles"
)

func HelpView() string {
	return styles.BaseStyle.Render(
		styles.HeaderStyle.Render("Help\n\n") +
			"Navigation\n" +
			"  ↑/↓: Move selection\n" +
			"  Enter: View selected note\n" +
			"  Esc: Go back/close\n\n" +
			"Actions\n" +
			"  n: New note\n" +
			"  e: Edit note\n" +
			"  d: Delete note\n" +
			"  /: Search notes\n" +
			"  x: Export note\n\n" +
			"General\n" +
			"  h: Toggle help\n" +
			"  q: Quit\n",
	)
}

func AboutView() string {
	return styles.BaseStyle.Render(
		styles.HeaderStyle.Render("Journal TUI\n\n") +
			"A terminal-based journaling app\n" +
			"Built with Bubble Tea & ❤️\n\n" +
			styles.HelpStyle.Render("Press any key to close"),
	)
}
