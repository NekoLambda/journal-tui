package components

import (
	"github.com/NekoLambda/journal-tui/ui/styles"
	"github.com/charmbracelet/glamour"
)

type Preview struct {
	content  string
	renderer *glamour.TermRenderer
}

func NewPreview() (*Preview, error) {
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		return nil, err
	}

	return &Preview{
		renderer: r,
	}, nil
}

func (p *Preview) SetContent(content string) error {
	p.content = content
	return nil
}

func (p Preview) View() string {
	if p.content == "" {
		return styles.PreviewStyle.Render("No content to display")
	}

	out, err := p.renderer.Render(p.content)
	if err != nil {
		return styles.PreviewStyle.Render("Error rendering markdown: " + err.Error())
	}

	return styles.PreviewStyle.Render(out)
}
