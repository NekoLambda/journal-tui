package model

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"

	textinput "github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lithammer/fuzzysearch/fuzzy"

	"github.com/NekoLambda/journal-tui/internal/storage"
	"github.com/NekoLambda/journal-tui/ui"
)

type Mode int

const (
	ModeList Mode = iota
	ModeView
	ModeSearch
	ModeHelp
	ModeAbout
)

type Model struct {
	mode     Mode
	entries  []storage.Entry
	filtered []storage.Entry
	cursor   int
	ti       textinput.Model // title / small inputs
	searchTI textinput.Model // used for search / tag input reuse
	vp       viewport.Model
	viewText string
	err      error
	msg      string

	// styles
	headerStyle   lipgloss.Style
	selectedStyle lipgloss.Style
	normalStyle   lipgloss.Style
	helpStyle     lipgloss.Style
	inputStyle    lipgloss.Style
}

func New() Model {
	entries, _ := storage.LoadEntries()

	// textinputs
	ti := textinput.New()
	ti.Placeholder = "Title..."
	ti.CharLimit = 200
	ti.Width = 40

	search := textinput.New()
	search.Placeholder = "Search..."
	search.CharLimit = 200
	search.Width = 40

	// viewport (height tuned later on WindowSizeMsg)
	vp := viewport.New(80, 12)
	vp.SetContent("")

	// styles are assigned directly in the Model struct initialization

	m := Model{
		mode:          ModeList,
		entries:       entries,
		filtered:      entries,
		ti:            ti,
		searchTI:      search,
		vp:            vp,
		headerStyle:   ui.HeaderStyle,
		selectedStyle: ui.SelectedStyle,
		normalStyle:   ui.NormalStyle,
		helpStyle:     ui.HelpStyle,
		inputStyle:    ui.InputStyle,
	}
	return m
}

func (m Model) Init() tea.Cmd { return nil }

// -------------------- Update --------------------
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// handle resizing for viewport
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// reserve ~6 lines for header/help/etc
		h := msg.Height - 6
		if h < 6 {
			h = 6
		}
		m.vp.Width = msg.Width
		m.vp.Height = h
		if m.mode == ModeView {
			m.vp.SetContent(m.viewText)
		}
		// continue
	}

	switch m.mode {
	case ModeList:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "j", "down":
				if m.cursor < len(m.filtered)-1 {
					m.cursor++
				}
			case "k", "up":
				if m.cursor > 0 {
					m.cursor--
				}
			case "enter":
				if len(m.filtered) > 0 {
					ent := m.filtered[m.cursor]
					content, err := storage.LoadEntryContent(ent)
					if err != nil {
						m.err = err
					} else {
						m.viewText = renderSimpleMarkdown(content, m.headerStyle, m.normalStyle)
						m.vp.SetContent(m.viewText)
						m.vp.GotoTop()
						m.mode = ModeView
					}
				}
			case "n":
				// create new entry workflow: ask title, open editor, save
				m.mode = ModeSearch // reuse searchTI as title input step (short)
				m.searchTI.SetValue("")
				m.searchTI.Placeholder = "New title (type and Enter)..."
				m.searchTI.Focus()
			case "d":
				if len(m.filtered) > 0 {
					ent := m.filtered[m.cursor]
					_ = storage.DeleteEntry(ent)
					m.reloadEntries()
				}
			case "e":
				// edit in-place (opens editor on the file)
				if len(m.filtered) > 0 {
					ent := m.filtered[m.cursor]
					path := filepath.Join("data", ent.Filename)
					if err := storage.EditEntry(path); err != nil {
						m.err = err
					} else {
						// after editing, reload and maybe rename
						old := ent // keep original
						m.reloadEntries()
						m.renameIfTitleChanged(old)
					}
				}
			case "x":
				// export selected entry (single)
				if len(m.filtered) > 0 {
					ent := m.filtered[m.cursor]
					path := filepath.Join("data", ent.Filename)
					if _, err := storage.ExportEntry(path); err != nil {
						m.err = err
					} else {
						m.msg = "Exported."
					}
				}
			case "/":
				m.mode = ModeSearch
				m.searchTI.SetValue("")
				m.searchTI.Placeholder = "Search..."
				m.searchTI.Focus()
			case "h":
				m.mode = ModeHelp
			case "a":
				m.mode = ModeAbout
			}
		}
	case ModeSearch:
		// text input driven (live filter)
		var cmd tea.Cmd
		m.searchTI, cmd = m.searchTI.Update(msg)
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				// finalise search (already applied)
				m.mode = ModeList
			case "esc":
				// cancel search -> clear
				m.searchTI.SetValue("")
				m.applyFilter("")
				m.mode = ModeList
			default:
				// live filtering
				m.applyFilter(m.searchTI.Value())
			}
		}
		return m, cmd
	case ModeView:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "q", "esc":
				// when leaving view, reset any active search so user returns to full list
				if strings.TrimSpace(m.searchTI.Value()) != "" {
					m.searchTI.SetValue("")
					m.applyFilter("")
				}
				m.mode = ModeList
			case "j", "down":
				m.vp.LineDown(1)
			case "k", "up":
				m.vp.LineUp(1)
			case "pgdown":
				m.vp.SetYOffset(m.vp.YOffset + int(m.vp.Height))
			case "pgup":
				m.vp.SetYOffset(m.vp.YOffset - int(m.vp.Height))
			case "g":
				m.vp.GotoTop()
			case "G":
				m.vp.GotoBottom()
			case "e":
				// edit current entry
				if m.cursor < len(m.filtered) {
					ent := m.filtered[m.cursor]
					path := filepath.Join("data", ent.Filename)
					if err := storage.EditEntry(path); err != nil {
						m.err = err
					} else {
						old := ent
						m.reloadEntries()
						m.renameIfTitleChanged(old)
						// refresh view content for this item
						for _, e := range m.entries {
							if e.Filename == ent.Filename {
								content, err := storage.LoadEntryContent(e)
								if err == nil {
									m.viewText = renderSimpleMarkdown(content, m.headerStyle, m.normalStyle)
									m.vp.SetContent(m.viewText)
									m.vp.GotoTop()
								}
								break
							}
						}
					}
				}
			}
		}
	case ModeHelp:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "q", "esc":
				m.mode = ModeList
			}
		}
	case ModeAbout:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "q", "esc":
				m.mode = ModeList
			}
		}
	}

	return m, nil
}

// -------------------- View --------------------
func (m Model) View() string {
	var b strings.Builder
	b.WriteString(m.headerStyle.Render("📓 Journal-TUI") + "\n\n")

	switch m.mode {
	case ModeList:
		if len(m.filtered) == 0 {
			b.WriteString(m.normalStyle.Render("(no entries)") + "\n")
		}
		for i, e := range m.filtered {
			if i == m.cursor {
				line := m.selectedStyle.Render("> " + e.Title)
				b.WriteString(line + "\n")
			} else {
				line := m.normalStyle.Render("  " + e.Title)
				b.WriteString(line + "\n")
			}
		}
		b.WriteString("\n")
		b.WriteString(m.helpStyle.Render("n: new  e: edit  d: delete  enter: view  /: search  x: export  h: help  a: about  q: quit"))
		if m.err != nil {
			b.WriteString("\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Render("Error: "+m.err.Error()))
		}
		if m.msg != "" {
			b.WriteString("\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")).Render(m.msg))
		}
	case ModeSearch:
		b.WriteString(m.normalStyle.Render("Search (live):\n\n"))
		b.WriteString(m.inputStyle.Render(m.searchTI.View()) + "\n\n")
		b.WriteString(m.renderListSnippet())
	case ModeView:
		b.WriteString(m.normalStyle.Render("[Viewing — press q to go back]\n\n"))
		b.WriteString(m.vp.View())
		b.WriteString("\n")
	case ModeHelp:
		b.WriteString(lipgloss.NewStyle().Padding(1, 2).Render(
			"Help\n\n" +
				"n : new note (asks for title, then opens editor)\n" +
				"e : edit selected note\n" +
				"d : delete selected note\n" +
				"Enter : view selected note\n" +
				"/ : search notes (live)\n" +
				"x : export selected note\n" +
				"h : help\n" +
				"a : about\n" +
				"q : quit\n\n" +
				"(press q or Esc to return)",
		))
	case ModeAbout:
		b.WriteString(lipgloss.NewStyle().Padding(1, 2).Render(
			"Journal-TUI\n\n" +
				"Minimal TUI journal built with Bubble Tea and friends.\n" +
				"GitHub: github.com/NekoLambda/journal-tui\n\n" +
				"(press q or Esc to return)",
		))
	}

	return b.String()
}

// -------------------- Helpers --------------------
func (m *Model) reloadEntries() {
	ents, _ := storage.LoadEntries()
	m.entries = ents
	// default filtered set
	m.filtered = make([]storage.Entry, len(ents))
	copy(m.filtered, ents)
	// clamp cursor
	if m.cursor >= len(m.filtered) && len(m.filtered) > 0 {
		m.cursor = len(m.filtered) - 1
	}
	if len(m.filtered) == 0 {
		m.cursor = 0
	}
}

func (m *Model) applyFilter(query string) {
	q := strings.TrimSpace(query)
	if q == "" {
		m.filtered = make([]storage.Entry, len(m.entries))
		copy(m.filtered, m.entries)
		return
	}
	// fuzzy-rank on titles (best matches first). We'll include content matches by fallback.
	titles := make([]string, len(m.entries))
	for i := range m.entries {
		titles[i] = m.entries[i].Title
	}
	ranked := fuzzy.RankFindFold(q, titles)
	sort.Sort(ranked)
	out := []storage.Entry{}
	added := map[int]bool{}
	for _, r := range ranked {
		out = append(out, m.entries[r.OriginalIndex])
		added[r.OriginalIndex] = true
	}
	// also include any entries where content contains q (case-insensitive) but not already included
	lq := strings.ToLower(q)
	for i, e := range m.entries {
		if added[i] {
			continue
		}
		if strings.Contains(strings.ToLower(e.Content), lq) {
			out = append(out, e)
		}
	}
	m.filtered = out
	if m.cursor >= len(m.filtered) && len(m.filtered) > 0 {
		m.cursor = len(m.filtered) - 1
	} else if len(m.filtered) == 0 {
		m.cursor = 0
	}
}

func (m *Model) renderListSnippet() string {
	var b strings.Builder
	for i, e := range m.filtered {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}
		b.WriteString(fmt.Sprintf("%s %s\n", cursor, e.Title))
	}
	return b.String()
}

// renameIfTitleChanged: if title in file header changed to a different value, rename file accordingly
func (m *Model) renameIfTitleChanged(old storage.Entry) {
	// reload entries and find matching filename
	for _, e := range m.entries {
		if e.Filename == old.Filename {
			// if title differs, rename file
			if e.Title != old.Title {
				oldPath := filepath.Join("data", old.Filename)
				newBase := slugify(e.Title)
				newFilename := newBase + ".md"
				newPath := filepath.Join("data", newFilename)

				// avoid clobbering existing file
				if _, err := os.Stat(newPath); err == nil {
					// file exists — append timestamp
					newFilename = fmt.Sprintf("%s-%d.md", newBase, time.Now().Unix())
					newPath = filepath.Join("data", newFilename)
				}
				if err := os.Rename(oldPath, newPath); err != nil {
					m.err = err
				} else {
					// reload to pick up new filename
					m.reloadEntries()
				}
			}
			break
		}
	}
}

func slugify(title string) string {
	t := strings.ToLower(title)
	t = strings.ReplaceAll(t, " ", "_")
	t = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' {
			return r
		}
		return -1
	}, t)
	if t == "" {
		t = "untitled"
	}
	return t
}

// very small markdown -> styled plaintext renderer (headings, code fences, paragraphs)
func renderSimpleMarkdown(raw string, headerStyle, normalStyle lipgloss.Style) string {
	var b strings.Builder
	lines := strings.Split(raw, "\n")
	inCode := false
	for _, L := range lines {
		trim := strings.TrimSpace(L)
		if strings.HasPrefix(trim, "```") {
			inCode = !inCode
			if inCode {
				b.WriteString(lipgloss.NewStyle().Background(lipgloss.Color("#282A36")).Render("```") + "\n")
			} else {
				b.WriteString(lipgloss.NewStyle().Background(lipgloss.Color("#282A36")).Render("```") + "\n")
			}
			continue
		}
		if inCode {
			b.WriteString(lipgloss.NewStyle().Background(lipgloss.Color("#1E1F29")).Render(L) + "\n")
			continue
		}
		if strings.HasPrefix(trim, "# ") {
			h := strings.TrimSpace(strings.TrimPrefix(trim, "# "))
			b.WriteString(headerStyle.Render(h) + "\n\n")
			continue
		}
		if strings.HasPrefix(trim, "## ") {
			h := strings.TrimSpace(strings.TrimPrefix(trim, "## "))
			b.WriteString(lipgloss.NewStyle().Bold(true).Render(h) + "\n")
			continue
		}
		if trim == "" {
			b.WriteString("\n")
		} else {
			b.WriteString(normalStyle.Render(L) + "\n")
		}
	}
	return b.String()
}
