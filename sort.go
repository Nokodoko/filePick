package main

import (
	"sort"
	"strings"
)

// SortMode represents the available sort modes.
type SortMode int

const (
	SortByName SortMode = iota
	SortByTime
	SortBySize
)

// String returns the display name for the sort mode.
func (s SortMode) String() string {
	switch s {
	case SortByName:
		return "name"
	case SortByTime:
		return "time"
	case SortBySize:
		return "size"
	default:
		return "name"
	}
}

// NextSortMode cycles through sort modes: name -> time -> size -> name.
func NextSortMode(current SortMode) SortMode {
	switch current {
	case SortByName:
		return SortByTime
	case SortByTime:
		return SortBySize
	case SortBySize:
		return SortByName
	default:
		return SortByName
	}
}

// sortEntries sorts entries in place according to the given sort mode.
// If dirsFirst is true, directories are always sorted before files.
func sortEntries(entries []DirEntry, mode SortMode, dirsFirst bool) {
	sort.SliceStable(entries, func(i, j int) bool {
		a, b := entries[i], entries[j]

		// Directories first
		if dirsFirst {
			if a.IsDir && !b.IsDir {
				return true
			}
			if !a.IsDir && b.IsDir {
				return false
			}
		}

		// Within same type (or if dirsFirst is off), sort by mode
		switch mode {
		case SortByTime:
			// Newest first
			if !a.ModTime.Equal(b.ModTime) {
				return a.ModTime.After(b.ModTime)
			}
			return strings.ToLower(a.Name) < strings.ToLower(b.Name)
		case SortBySize:
			// Largest first
			if a.Size != b.Size {
				return a.Size > b.Size
			}
			return strings.ToLower(a.Name) < strings.ToLower(b.Name)
		default: // SortByName
			return strings.ToLower(a.Name) < strings.ToLower(b.Name)
		}
	})
}
