package storage

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"
)

type Entry struct {
	Title    string
	Filename string
	Content  string // Populate when loading
	ModTime  time.Time
	Tags     []string
}

const dataDir = "data"
const metaFile = "metadata.json"

// EnsureDataDir makes sure data dir exists
func EnsureDataDir() error {
	return os.MkdirAll(dataDir, 0o755)
}

// metadata is a simple map: filename -> tags
func loadMetadata() (map[string][]string, error) {
	mp := map[string][]string{}
	path := filepath.Join(dataDir, metaFile)
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return mp, nil
		}
		return nil, err
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&mp); err != nil {
		return nil, err
	}
	return mp, nil
}

func saveMetadata(mp map[string][]string) error {
	path := filepath.Join(dataDir, metaFile)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(mp)
}

// sanitize a string to slug
func slugify(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			b.WriteRune(r)
		} else if unicode.IsSpace(r) || r == '-' || r == '_' {
			b.WriteRune('-')
		}
	}
	out := b.String()
	out = strings.Trim(out, "-")
	if out == "" {
		out = "entry"
	}
	return out
}

// SaveEntry writes a markdown file named with timestamp + slug and returns Entry
// now accepts tags
func SaveEntry(title string, content string, tags []string) (Entry, error) {
	if err := EnsureDataDir(); err != nil {
		return Entry{}, err
	}
	ts := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("%s-%s.md", ts, slugify(title))
	path := filepath.Join(dataDir, filename)

	f, err := os.Create(path)
	if err != nil {
		return Entry{}, err
	}
	defer f.Close()

	// write a simple markdown: title + body
	_, err = f.WriteString("# " + title + "\n\n" + content)
	if err != nil {
		return Entry{}, err
	}
	fi, _ := f.Stat()

	// update metadata
	mp, err := loadMetadata()
	if err != nil {
		return Entry{}, err
	}
	if tags == nil {
		tags = []string{}
	}
	mp[filename] = tags
	if err := saveMetadata(mp); err != nil {
		// metadata is important
		return Entry{}, err
	}

	return Entry{Title: title, Filename: filename, Content: content, ModTime: fi.ModTime(), Tags: tags}, nil
}

// LoadEntries lists markdown files and returns entries with title (from file first line if present)
func LoadEntries() ([]Entry, error) {
	if err := EnsureDataDir(); err != nil {
		return nil, err
	}
	entries := []Entry{}

	mp, err := loadMetadata()
	if err != nil {
		return nil, err
	}

	err = filepath.WalkDir(dataDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		// skip metadata
		if filepath.Base(path) == metaFile {
			return nil
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}
		// skip tmp folder files
		if strings.Contains(path, string(filepath.Separator)+"tmp"+string(filepath.Separator)) {
			return nil
		}
		bytes, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		content := string(bytes)

		// determine title from first heading or filename
		var titleStr string
		lines := strings.SplitN(content, "\n", 2)
		if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[0]), "# ") {
			titleStr = strings.TrimSpace(strings.TrimPrefix(lines[0], "# "))
		} else {
			titleStr = filepath.Base(path)
		}

		fi, _ := os.Stat(path)
		filename := filepath.Base(path)
		tags := mp[filename]
		entries = append(entries, Entry{
			Title:    titleStr,
			Filename: filename,
			Content:  content,
			ModTime:  fi.ModTime(),
			Tags:     tags,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	// sort by ModTime descending
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].ModTime.After(entries[i].ModTime) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}
	return entries, nil
}

// LoadEntryContent reads full markdown content (returns raw string)
func LoadEntryContent(e Entry) (string, error) {
	path := filepath.Join(dataDir, e.Filename)
	bytes, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// DeleteEntry removes the entry file and updates metadata
func DeleteEntry(e Entry) error {
	path := filepath.Join(dataDir, e.Filename)
	if err := os.Remove(path); err != nil {
		return err
	}
	mp, _ := loadMetadata()
	delete(mp, e.Filename)
	_ = saveMetadata(mp)
	return nil
}

// ExportAll zips every .md file in data/ into exports/<timestamp>.zip
func ExportAll() (string, error) {
	if err := EnsureDataDir(); err != nil {
		return "", err
	}
	exportDir := "exports"
	if err := os.MkdirAll(exportDir, 0o755); err != nil {
		return "", err
	}
	ts := time.Now().Format("20060102-150405")
	zipPath := filepath.Join(exportDir, fmt.Sprintf("journal-export-%s.zip", ts))

	zf, err := os.Create(zipPath)
	if err != nil {
		return "", err
	}
	defer zf.Close()

	zw := zip.NewWriter(zf)
	defer zw.Close()

	err = filepath.WalkDir(dataDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}
		if filepath.Base(path) == metaFile {
			return nil
		}
		rel := filepath.Base(path)
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		w, err := zw.Create(rel)
		if err != nil {
			return err
		}
		_, err = io.Copy(w, f)
		return err
	})
	if err != nil {
		return "", err
	}
	return zipPath, nil
}
