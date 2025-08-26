package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSafeFilename(t *testing.T) {
	got := safeFilename("Hello World!!")
	if got != "hello_world" {
		t.Errorf("expected hello_world, got %s", got)
	}
}

func TestNewAndLoadEntry(t *testing.T) {
	os.RemoveAll("data")
	entry, err := NewEntry("TestNote", "This is a test")
	if err != nil {
		t.Fatalf("NewEntry failed: %v", err)
	}
	if !strings.HasPrefix(entry.Filename, "testnote") {
		t.Errorf("unexpected filename: %s", entry.Filename)
	}

	entries, err := LoadEntries()
	if err != nil {
		t.Fatalf("LoadEntries failed: %v", err)
	}
	if len(entries) == 0 {
		t.Errorf("expected at least one entry, got %d", len(entries))
	}
}

func TestExportEntry(t *testing.T) {
	os.RemoveAll("exports")
	entry, err := NewEntry("ExportMe", "Some content")
	if err != nil {
		t.Fatalf("NewEntry failed: %v", err)
	}
	path := filepath.Join("data", entry.Filename)
	if err := ExportEntry(path); err != nil {
		t.Fatalf("ExportEntry failed: %v", err)
	}
	files, err := os.ReadDir("exports")
	if err != nil || len(files) == 0 {
		t.Fatalf("expected export file, got none")
	}
}

// We can’t fully test EditEntry (it opens nano/editor),
// but we can at least check it doesn't error with invalid $EDITOR.
func TestEditEntry_NoEditor(t *testing.T) {
	entry, _ := NewEntry("Dummy", "content")
	path := filepath.Join("data", entry.Filename)
	os.Setenv("EDITOR", "/bin/true") // no-op editor
	if err := EditEntry(path); err != nil {
		t.Errorf("EditEntry failed: %v", err)
	}
}
