---

# Journal-TUI

A heavily work-in-progress terminal-based journaling and note-taking app built with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## ✨ Features

- 📂 Organize notes into folders
- 📝 Create, view, edit, and delete notes
- 🔍 Fuzzy search (title + content)
- 📤 Export notes to plain text
- 🖥️ Minimal TUI interface with [Charm](https://charm.sh) ecosystem
- ❓ Help and About screens for quick reference

## 📦 Project Structure

```markdown
```
```
journal-tui/
├── cmd/
│   └── journal-tui/
│       └── main.go          # entrypoint
├── internal/
│   ├── model/               
│   │   └── model.go         # state machine, modes, key handling
│   └── storage/
│       ├── storage.go       # File ops (save, edit, delete, etc.)
│       └── storage_test.go  # Unit tests
├── ui/                      # All Terminal UI related code
│   ├── components/          # Reusable widgets (note list, dialogs, help view)
│   │   ├── list.go          # Entry list (using Bubbles list)
│   │   ├── modal.go         # Reusable modal dialogs
│   │   ├── input.go         # Text input form
│   │   ├── preview.go       # Markdown preview (Glow)
│   │   └── help.go          # Help and About view
│   └── styles.go            # Lipgloss themes, colors, spacing
├── scripts/                 # Optional shell helpers (using Gum)
│   └── quicknote.sh         # Example quick journaling script
│
├── go.mod
└── go.sum
```

## 🚀 Getting Started

### Prerequisites
- Go 1.22+
- Git

### Install & Run

```bash
git clone https://github.com/<your-username>/journal-tui.git
cd journal-tui/cmd/journal
go run .
````

Or build:

```bash
go build -o journal .
./journal
```

## 🛠 Development

Run tests:

```bash
go test ./internal/storage/...
```

## 🔮 Roadmap

* [ ] Nested folders
* [ ] Better export formats (Markdown, PDF)
* [ ] Configurable keybindings
* [ ] Cloud sync

## 🤝 Contributing

Contributions are welcome! Feel free to fork and submit a PR.

## 📜 License

MIT

```