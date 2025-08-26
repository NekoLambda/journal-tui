package storage

import (
	"os"
	"os/exec"
)

// EditEntry opens the entry file in $EDITOR
func EditEntry(path string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nano" // Default to nano
	}

	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
