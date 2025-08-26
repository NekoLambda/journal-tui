---

# Journal-TUI

A heavily work-in-progress terminal-based journaling and note-taking app built with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## ✨ Features

- 📂 Organize notes into folders
- 📝 Create, view, edit, and delete notes
- 🔍 Fuzzy search (title + content)
- 📤 Export notes to plain text
 🖥️ Minimal TUI interface with [Charm](https://charm.sh) ecosystem
- ❓ Help and About screens for quick reference

## 📦 Project Structure

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
│       ├── storage.go       # file ops (save, edit, delete, etc.)
│       └── storage_test.go  # unit tests
├── ui/
│   ├── components/          # reusable widgets (note list, dialogs, help view)
│   │   ├── list.go
│   │   ├── modal.go
│   │   └── help.go
│   └── styles.go            # Lipgloss themes, colors, spacing
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