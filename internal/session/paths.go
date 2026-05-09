package session

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/mensfeld/code-on-incus/internal/tool"
)

// GetSessionsDir returns the sessions directory path for a given tool.
// For example: ~/.coi/sessions-claude, ~/.coi/sessions-aider, etc.
func GetSessionsDir(baseDir string, t tool.Tool) string {
	return filepath.Join(baseDir, t.SessionsDirName())
}

// GetAllSessionsDirs returns all sessions-* directories under baseDir.
// This is used by commands like `coi list` that need to find metadata
// across all tools, not just the currently configured one.
func GetAllSessionsDirs(baseDir string) []string {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil
	}

	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "sessions-") {
			dirs = append(dirs, filepath.Join(baseDir, entry.Name()))
		}
	}
	return dirs
}
