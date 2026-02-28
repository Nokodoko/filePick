package main

import (
	"os"
	"path/filepath"
	"time"
)

// DirEntry represents a single filesystem entry for display.
type DirEntry struct {
	Name         string
	Path         string // Full absolute path
	IsDir        bool
	IsSymlink    bool
	IsExecutable bool
	IsHidden     bool
	Size         int64
	ModTime      time.Time
}

// readDir reads the given directory and returns a slice of DirEntry.
// It handles stat failures gracefully by skipping entries that disappear
// between readdir and stat (race condition with filesystem events).
func readDir(dir string) ([]DirEntry, error) {
	osEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	entries := make([]DirEntry, 0, len(osEntries))
	for _, ose := range osEntries {
		fullPath := filepath.Join(dir, ose.Name())

		entry := DirEntry{
			Name:     ose.Name(),
			Path:     fullPath,
			IsHidden: len(ose.Name()) > 0 && ose.Name()[0] == '.',
		}

		// Use Lstat to detect symlinks without following them
		linfo, err := os.Lstat(fullPath)
		if err != nil {
			// Entry disappeared between ReadDir and Lstat — skip
			continue
		}

		entry.IsSymlink = linfo.Mode()&os.ModeSymlink != 0

		// For display purposes, resolve symlink to get the actual type
		var info os.FileInfo
		if entry.IsSymlink {
			info, err = os.Stat(fullPath)
			if err != nil {
				// Broken symlink — use lstat info
				info = linfo
			}
		} else {
			info = linfo
		}

		entry.IsDir = info.IsDir()
		entry.IsExecutable = !entry.IsDir && info.Mode()&0o111 != 0
		entry.Size = info.Size()
		entry.ModTime = info.ModTime()

		entries = append(entries, entry)
	}

	return entries, nil
}

// filterHidden returns only non-hidden entries if showHidden is false,
// or all entries if showHidden is true.
func filterHidden(entries []DirEntry, showHidden bool) []DirEntry {
	if showHidden {
		return entries
	}
	filtered := make([]DirEntry, 0, len(entries))
	for _, e := range entries {
		if !e.IsHidden {
			filtered = append(filtered, e)
		}
	}
	return filtered
}
