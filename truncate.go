package main

import (
	"os"
	"path/filepath"
	"strings"
)

// truncatePath truncates a path to fit within maxWidth characters.
// It replaces $HOME with ~ and truncates from the left at path component
// boundaries, prefixing with "..." when truncated.
func truncatePath(path string, maxWidth int) string {
	// Replace home directory with ~
	home, err := os.UserHomeDir()
	if err == nil && strings.HasPrefix(path, home) {
		path = "~" + path[len(home):]
	}

	if len(path) <= maxWidth {
		return path
	}

	// Split into components
	components := strings.Split(filepath.ToSlash(path), "/")

	// Try removing components from the left until it fits
	prefix := "..."
	for i := 1; i < len(components); i++ {
		candidate := prefix + "/" + strings.Join(components[i:], "/")
		if len(candidate) <= maxWidth {
			return candidate
		}
	}

	// If even the last component is too long, truncate it
	last := components[len(components)-1]
	if len(prefix)+1+len(last) > maxWidth {
		available := maxWidth - len(prefix)
		if available > 0 {
			return prefix + last[:available]
		}
		return prefix
	}

	return prefix + "/" + last
}
