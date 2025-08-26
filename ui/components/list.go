package components

import (
	"github.com/NekoLambda/journal-tui/ui/styles"
	"github.com/charmbracelet/bubbles/list"
)

type EntryItem struct {
	title    string
	filename string
}

func (i EntryItem) Title() string       { return i.title }
func (i EntryItem) Description() string { return "" }
func (i EntryItem) FilterValue() string { return i.title }

type EntryList struct {
	List     list.Model
	Selected EntryItem
}

func NewEntryList(height int) EntryList {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, height)
	l.SetShowHelp(false)
	l.Title = "Journal Entries"
	l.Styles.Title = styles.HeaderStyle

	return EntryList{
		List: l,
	}
}

func (el *EntryList) SetItems(items []EntryItem) {
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = item
	}
	el.List.SetItems(listItems)
}

func (el EntryList) View() string {
	return el.List.View()
}
