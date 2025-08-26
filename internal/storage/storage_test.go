package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSlugify(t *testing.T) {
	got := slugify("Hello World!!")
	if got != "hello-world" {
		t.Errorf("expected hello-world, got %s", got)
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
	exportPath, err := ExportEntry(path)
	if err != nil {
		t.Fatalf("ExportEntry failed: %v", err)
	}
	if _, err := os.Stat(exportPath); err != nil {
		t.Fatalf("expected export file at %s, got error: %v", exportPath, err)
	}
}

func TestEditEntry(t *testing.T) {
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "test-*.md")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Set test editor environment
	oldEditor := os.Getenv("EDITOR")
	os.Setenv("EDITOR", "echo") // Use echo as a mock editor that does nothing
	defer os.Setenv("EDITOR", oldEditor)

	// Test editing
	err = EditEntry(tmpFile.Name())
	if err != nil {
		t.Errorf("EditEntry failed: %v", err)
	}

	// Test with invalid path
	err = EditEntry("/path/that/does/not/exist")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
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
