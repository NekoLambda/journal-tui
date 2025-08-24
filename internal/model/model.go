package model

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/NekoLambda/journal-tui/internal/storage"
	textinput "github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

type Mode int

const (
	ModeList Mode = iota
	ModeInputTitle
	ModeInputTags
	ModeSearch
	ModeView
	ModeConfirmDelete
	ModeShowMessage
	ModeHelp
	ModeAbout
	ModeQuitting
)

type Model struct {
	entries    []storage.Entry
	filtered   []storage.Entry // after search/filter
	cursor     int
	mode       Mode
	ti         textinput.Model // text input for title
	searchTI   textinput.Model // used for search and tags input
	viewText   string
	confirmMsg string
	err        error
	msg        string
	vp         viewport.Model

	// styles
	headerStyle   lipgloss.Style
	selectedStyle lipgloss.Style
	normalStyle   lipgloss.Style
	inputStyle    lipgloss.Style
	helpStyle     lipgloss.Style
	borderStyle   lipgloss.Style
}

func New() Model {
	_ = storage.EnsureDataDir()

	// styles
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF79C6"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#50FA7B"))
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F8F8F2"))
	inputStyle := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1).MarginTop(1)
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4"))
	borderStyle := lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).BorderForeground(lipgloss.Color("#6272A4")).Padding(1)

	// create textinput instances
	tiInst := textinput.New()
	tiInst.Placeholder = "Entry title..."
	tiInst.CharLimit = 200
	tiInst.Focus()

	searchInst := textinput.New()
	searchInst.Placeholder = "Search..."
	searchInst.CharLimit = 200

	// viewport: initial size — updated on WindowSizeMsg
	vp := viewport.New(80, 12)
	vp.SetContent("")

	m := Model{
		mode:          ModeList,
		ti:            tiInst,
		searchTI:      searchInst,
		vp:            vp,
		headerStyle:   headerStyle,
		selectedStyle: selectedStyle,
		normalStyle:   normalStyle,
		inputStyle:    inputStyle,
		helpStyle:     helpStyle,
		borderStyle:   borderStyle,
	}
	m.reloadEntries()
	return m
}

func (m *Model) reloadEntries() {
	ents, err := storage.LoadEntries()
	if err != nil {
		m.err = err
		m.entries = []storage.Entry{}
		m.filtered = []storage.Entry{}
		return
	}
	m.entries = ents
	// default filtered = all
func (m *Model) applyFilter(query string) {
    if strings.TrimSpace(query) == "" {
        m.filtered = m.entries
        return
    }

    ranked := fuzzy.RankFindFold(query, titlesFromEntries(m.entries))
    sort.Sort(ranked) // lowest distance first

    var res []Entry
    for _, r := range ranked {
        res = append(res, m.entries[r.OriginalIndex])
    }
    m.filtered = res
}

func titlesFromEntries(entries []Entry) []string {
    out := make([]string, len(entries))
    for i, e := range entries {
        out[i] = e.Title
    }
    return out
}

	if len(m.filtered) == 0 {
		m.cursor = 0
	}
}

func (m Model) Init() tea.Cmd { return nil }

func containsIgnoreCase(s, q string) bool {
	s = strings.ToLower(s)
	q = strings.ToLower(q)
	return strings.Contains(s, q)
}

func (m *Model) applyFilter(query string) {
	if query == "" {
		m.filtered = make([]storage.Entry, len(m.entries))
		copy(m.filtered, m.entries)
		return
	}
	q := strings.TrimSpace(query)
	out := []storage.Entry{}
	for _, e := range m.entries {
		if containsIgnoreCase(e.Title, q) || containsIgnoreCase(e.Content, q) {
			out = append(out, e)
		} else {
			// check tags
			for _, t := range e.Tags {
				if containsIgnoreCase(t, q) {
					out = append(out, e)
					break
				}
			}
		}
	}
	m.filtered = out
	if m.cursor >= len(m.filtered) && len(m.filtered) > 0 {
		m.cursor = len(m.filtered) - 1
	} else if len(m.filtered) == 0 {
		m.cursor = 0
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// window resizing: update viewport dims
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// header + help area ~6 lines
		height := msg.Height - 6
		if height < 6 {
			height = 6
		}
		m.vp.Width = msg.Width
		m.vp.Height = height
		m.vp.SetContent(m.viewText)
		// continue to handle other messages
	}

	switch m.mode {
	case ModeList:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "j", "down":
				if m.cursor < len(m.filtered)-1 {
					m.cursor++
				}
			case "k", "up":
				if m.cursor > 0 {
					m.cursor--
				}
			case "n":
				m.mode = ModeInputTitle
				m.ti.SetValue("")
				m.ti.Focus()
			case "/":
				m.mode = ModeSearch
				m.searchTI.SetValue("")
				m.searchTI.Focus()
			case "enter":
				if len(m.filtered) > 0 {
					ent := m.filtered[m.cursor]
					fullPath := filepath.Join("data", ent.Filename)
					rendered, err := renderWithGlow(fullPath)
					if err == nil && rendered != "" {
						m.viewText = rendered
					} else {
						m.viewText = renderSimpleMarkdown(ent.Content, m.headerStyle, m.normalStyle)
					}
					m.vp.SetContent(m.viewText)
					m.mode = ModeView
				}
			case "d":
				if len(m.filtered) > 0 {
					ent := m.filtered[m.cursor]
					m.confirmMsg = fmt.Sprintf("Delete '%s'?", ent.Title)
					m.mode = ModeConfirmDelete
				}
			case "h":
				m.mode = ModeHelp
			case "a":
				m.mode = ModeAbout
			case "e":
				// Edit selected entry in-place using the system editor (Notepad on Windows)
				if len(m.filtered) == 0 {
					break
				}
				ent := m.filtered[m.cursor]

				fullPath := filepath.Join("data", ent.Filename)
				var openCmd *exec.Cmd
				if runtime.GOOS == "windows" {
					openCmd = exec.Command("notepad", fullPath)
				} else {
					editor := os.Getenv("EDITOR")
					if editor == "" {
						editor = "vi"
					}
					openCmd = exec.Command(editor, fullPath)
					openCmd.Stdin = os.Stdin
					openCmd.Stdout = os.Stdout
					openCmd.Stderr = os.Stderr
				}
				openCmd.Stdin = os.Stdin
				openCmd.Stdout = os.Stdout
				openCmd.Stderr = os.Stderr

				if err := openCmd.Run(); err != nil {
					m.err = err
				} else {
					// reload entries and update viewport content for this file
					m.reloadEntries()
					// find updated entry and update viewText + viewport
					for _, e := range m.entries {
						if e.Filename == ent.Filename {
							rendered, err := renderWithGlow(filepath.Join("data", e.Filename))
							if err == nil && rendered != "" {
								m.viewText = rendered
							} else {
								m.viewText = renderSimpleMarkdown(e.Content, m.headerStyle, m.normalStyle)
							}
							m.vp.SetContent(m.viewText)
							m.vp.GotoTop()
							break
						}
					}

				}

			case "E":
				zipPath, err := storage.ExportAll()
				if err != nil {
					m.err = err
				} else {
					m.msg = "Exported to " + zipPath
					m.mode = ModeShowMessage
				}
			}
		}
	case ModeHelp:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.mode = ModeList
			}
		}
	case ModeAbout:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.mode = ModeList
			}
		}
	case ModeSearch:
		var cmd tea.Cmd
		m.searchTI, cmd = m.searchTI.Update(msg)
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				m.applyFilter(m.searchTI.Value())
				m.mode = ModeList
			case "esc":
				m.mode = ModeList
			default:
				// live filter
				m.applyFilter(m.searchTI.Value())
			}
		}
		return m, cmd
	case ModeInputTitle:
		var cmd tea.Cmd
		m.ti, cmd = m.ti.Update(msg)
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "enter" {
				title := strings.TrimSpace(m.ti.Value())
				if title == "" {
					m.err = fmt.Errorf("title cannot be empty")
				} else {
					// go to tags input before launching editor
					m.mode = ModeInputTags
					m.searchTI.SetValue("") // reuse searchTI as tags input
					m.searchTI.Placeholder = "comma-separated tags (optional)"
					m.searchTI.Focus()
				}
			} else if msg.String() == "esc" {
				m.mode = ModeList
			}
		}
		return m, cmd
	case ModeInputTags:
		var cmd tea.Cmd
		m.searchTI, cmd = m.searchTI.Update(msg)
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "enter" {
				tagsRaw := strings.TrimSpace(m.searchTI.Value())
				tags := []string{}
				if tagsRaw != "" {
					for _, t := range strings.Split(tagsRaw, ",") {
						if s := strings.TrimSpace(t); s != "" {
							tags = append(tags, s)
						}
					}
				}
				title := strings.TrimSpace(m.ti.Value())
				// create tmp file for body
				tmpdir := filepath.Join(".", "data", "tmp")
				_ = os.MkdirAll(tmpdir, 0o755)
				tmpfile := filepath.Join(tmpdir, fmt.Sprintf("entry-%d.md", time.Now().UnixNano()))
				f, err := os.Create(tmpfile)
				if err != nil {
					m.err = err
					m.mode = ModeList
					return m, cmd
				}
				_ = f.Close()
				// open editor
				var openCmd *exec.Cmd
				if runtime.GOOS == "windows" {
					openCmd = exec.Command("notepad", tmpfile)
				} else {
					editor := os.Getenv("EDITOR")
					if editor == "" {
						editor = "vi"
					}
					openCmd = exec.Command(editor, tmpfile)
					openCmd.Stdin = os.Stdin
					openCmd.Stdout = os.Stdout
					openCmd.Stderr = os.Stderr
				}
				openCmd.Stdin = os.Stdin
				openCmd.Stdout = os.Stdout
				openCmd.Stderr = os.Stderr
				if err := openCmd.Run(); err != nil {
					m.err = err
					m.mode = ModeList
					return m, cmd
				}
				body, err := os.ReadFile(tmpfile)
				if err != nil {
					m.err = err
					m.mode = ModeList
					return m, cmd
				}
				_, err = storage.SaveEntry(title, string(body), tags)
				if err != nil {
					m.err = err
					m.mode = ModeList
					return m, cmd
				}
				_ = os.Remove(tmpfile)
				_ = os.RemoveAll(tmpdir)
				m.reloadEntries()
				m.mode = ModeList
			} else if msg.String() == "esc" {
				m.mode = ModeList
			}
		}
		return m, cmd
	case ModeView:
		// viewport handles scroll commands
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "q", "esc":
				// If a search is active, clear it so we return to the full list
				// (searchTI exists in this model; applyFilter resets filtered list)
				if strings.TrimSpace(m.searchTI.Value()) != "" {
					m.searchTI.SetValue("")
					m.applyFilter("") // reset filtered -> all entries
				}
				// ensure viewport is reset (optional)
				m.vp.GotoTop()
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
			}
		case tea.WindowSizeMsg:
			// viewport handled above; nothing more
		}
	case ModeConfirmDelete:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "y", "Y", "enter":
				if len(m.filtered) > 0 {
					ent := m.filtered[m.cursor]
					_ = storage.DeleteEntry(ent)
					m.reloadEntries()
				}
				m.mode = ModeList
			case "n", "N", "esc":
				m.mode = ModeList
			}
		}
	case ModeShowMessage:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			// any key returns to list
			_ = msg
			m.mode = ModeList
		}
	}

	return m, nil
}

func (m Model) View() string {
	// clear error after rendering once
	if m.err != nil {
		defer func() { m.err = nil }()
	}

	var b strings.Builder
	b.WriteString(m.headerStyle.Render("📓 Journal — simple TUI (n: new, /: search, E: export)") + "\n\n")

	switch m.mode {
	case ModeList:
		if len(m.filtered) == 0 {
			b.WriteString(m.normalStyle.Render("(no entries yet)") + "\n")
		}
		for i, e := range m.filtered {
			if i == m.cursor {
				line := m.selectedStyle.Render("> " + e.Title)
				// show tags lightly
				if len(e.Tags) > 0 {
					line += " " + lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")).Render(" ["+strings.Join(e.Tags, ", ")+"]")
				}
				b.WriteString(line + "\n")
			} else {
				line := m.normalStyle.Render("  " + e.Title)
				if len(e.Tags) > 0 {
					line += " " + lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")).Render(" ["+strings.Join(e.Tags, ", ")+"]")
				}
				b.WriteString(line + "\n")
			}
		}
		b.WriteString("\n")
		b.WriteString(m.helpStyle.Render("n = new   / = search   Enter = view   d = delete   E = export   q = quit") + "\n")
		if m.err != nil {
			b.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Render("Error: "+m.err.Error()) + "\n")
		}
	case ModeSearch:
		b.WriteString(m.normalStyle.Render("Search entries (live):\n\n"))
		b.WriteString(m.inputStyle.Render(m.searchTI.View()) + "\n")
	case ModeInputTitle:
		b.WriteString(m.normalStyle.Render("New entry — enter a title and press Enter to continue\n\n"))
		b.WriteString(m.inputStyle.Render(m.ti.View()) + "\n")
	case ModeInputTags:
		b.WriteString(m.normalStyle.Render("Add tags (comma-separated, optional), press Enter to open editor\n\n"))
		b.WriteString(m.inputStyle.Render(m.searchTI.View()) + "\n")
	case ModeView:
		b.WriteString(m.normalStyle.Render("[Viewing entry — press q or Esc to go back]\n\n"))
		// use viewport view so content scrolls
		b.WriteString(m.vp.View() + "\n")
	case ModeConfirmDelete:
		b.WriteString(m.normalStyle.Render("CONFIRM: " + m.confirmMsg + " (y/n)\n"))
	case ModeShowMessage:
		b.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")).Render(m.msg) + "\n")
	case ModeHelp:
		return lipgloss.NewStyle().Padding(1, 2).Render(
			"Help Menu\n\n" +
				"n: new note\n" +
				"e: edit note\n" +
				"d: delete note\n" +
				"Enter: view note\n" +
				"/: search notes\n" +
				"E: export note\n" +
				"h: help\n" +
				"a: about\n" +
				"q: quit\n\n" +
				"(press q or esc to return)",
		)
	case ModeAbout:
		return lipgloss.NewStyle().Padding(1, 2).Render(
			"Journal-TUI v0.1\n" +
				"Built with Charmbracelet Bubble Tea + Lipgloss + Glow\n" +
				"https://github.com/NekoLambda/journal-tui\n\n" +
				"(press q or esc to return)",
		)
	}
	return b.String()
}

// renderWithGlow attempts to run 'glow <path>' and capture its ANSI output.
func renderWithGlow(path string) (string, error) {
	_, err := exec.LookPath("glow")
	if err != nil {
		return "", err
	}
	cmd := exec.Command("glow", path)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	// try to avoid pager
	cmd.Env = append(os.Environ(), "PAGER=cat")
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return out.String(), nil
}

// renderSimpleMarkdown produces a basic styled representation of markdown content.
func renderSimpleMarkdown(raw string, headerStyle, normalStyle lipgloss.Style) string {
	var b strings.Builder
	lines := strings.Split(raw, "\n")
	inCode := false
	for _, L := range lines {
		line := L
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "```") {
			if !inCode {
				inCode = true
				b.WriteString(lipgloss.NewStyle().Background(lipgloss.Color("#282A36")).Foreground(lipgloss.Color("#8BE9FD")).Render("```") + "\n")
			} else {
				inCode = false
				b.WriteString(lipgloss.NewStyle().Background(lipgloss.Color("#282A36")).Foreground(lipgloss.Color("#8BE9FD")).Render("```") + "\n")
			}
			continue
		}
		if inCode {
			b.WriteString(lipgloss.NewStyle().Background(lipgloss.Color("#1E1F29")).Foreground(lipgloss.Color("#C792EA")).Render(line) + "\n")
			continue
		}
		if strings.HasPrefix(trim, "# ") {
			content := strings.TrimSpace(strings.TrimPrefix(trim, "# "))
			b.WriteString(headerStyle.Render(content) + "\n\n")
			continue
		}
		if strings.HasPrefix(trim, "## ") {
			content := strings.TrimSpace(strings.TrimPrefix(trim, "## "))
			b.WriteString(lipgloss.NewStyle().Bold(true).Render(content) + "\n")
			continue
		}
		if trim == "" {
			b.WriteString("\n")
		} else {
			b.WriteString(normalStyle.Render(line) + "\n")
		}
	}
	return b.String()
}
