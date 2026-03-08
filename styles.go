package main

import (
	"os"

	"charm.land/lipgloss/v2"
)

// noColor checks if color output should be disabled.
func noColor() bool {
	return os.Getenv("NO_COLOR") != ""
}

// Styles holds all lipgloss styles for the application.
type Styles struct {
	PathBar       lipgloss.Style
	Separator     lipgloss.Style
	DirName       lipgloss.Style
	FileName      lipgloss.Style
	ExecName      lipgloss.Style
	SymlinkName   lipgloss.Style
	HiddenEntry   lipgloss.Style
	SelectionBar  lipgloss.Style
	StatusBar     lipgloss.Style
	ErrorMsg      lipgloss.Style
	ItemCount     lipgloss.Style
	SortIndicator lipgloss.Style
	SearchPrompt  lipgloss.Style
	SearchInput   lipgloss.Style
	SearchCursor  lipgloss.Style
	GitBranch     lipgloss.Style
	GitModified   lipgloss.Style
	GitStaged     lipgloss.Style
	GitUntracked  lipgloss.Style
	GitConflict   lipgloss.Style
}

// NewStyles creates a new Styles with the color palette from the spec.
func NewStyles() Styles {
	if noColor() {
		return Styles{
			PathBar:       lipgloss.NewStyle(),
			Separator:     lipgloss.NewStyle(),
			DirName:       lipgloss.NewStyle(),
			FileName:      lipgloss.NewStyle(),
			ExecName:      lipgloss.NewStyle(),
			SymlinkName:   lipgloss.NewStyle(),
			HiddenEntry:   lipgloss.NewStyle(),
			SelectionBar:  lipgloss.NewStyle(),
			StatusBar:     lipgloss.NewStyle(),
			ErrorMsg:      lipgloss.NewStyle(),
			ItemCount:     lipgloss.NewStyle(),
			SortIndicator: lipgloss.NewStyle(),
			SearchPrompt:  lipgloss.NewStyle(),
			SearchInput:   lipgloss.NewStyle(),
			SearchCursor:  lipgloss.NewStyle(),
			GitBranch:     lipgloss.NewStyle(),
			GitModified:   lipgloss.NewStyle(),
			GitStaged:     lipgloss.NewStyle(),
			GitUntracked:  lipgloss.NewStyle(),
			GitConflict:   lipgloss.NewStyle(),
		}
	}

	return Styles{
		PathBar: lipgloss.NewStyle().
			Foreground(lipgloss.ANSIColor(7)).
			Bold(true).
			Faint(true),

		Separator: lipgloss.NewStyle().
			Foreground(lipgloss.ANSIColor(240)).
			Faint(true),

		DirName: lipgloss.NewStyle().
			Foreground(lipgloss.ANSIColor(4)).
			Bold(true),

		FileName: lipgloss.NewStyle().
			Foreground(lipgloss.ANSIColor(7)),

		ExecName: lipgloss.NewStyle().
			Foreground(lipgloss.ANSIColor(2)).
			Bold(true),

		SymlinkName: lipgloss.NewStyle().
			Foreground(lipgloss.ANSIColor(6)).
			Italic(true),

		HiddenEntry: lipgloss.NewStyle().
			Foreground(lipgloss.ANSIColor(242)).
			Faint(true),

		SelectionBar: lipgloss.NewStyle().
			Foreground(lipgloss.ANSIColor(15)).
			Background(lipgloss.Color("#3b3b3b")).
			Bold(true),

		StatusBar: lipgloss.NewStyle().
			Foreground(lipgloss.ANSIColor(242)).
			Faint(true),

		ErrorMsg: lipgloss.NewStyle().
			Foreground(lipgloss.ANSIColor(1)).
			Bold(true),

		ItemCount: lipgloss.NewStyle().
			Foreground(lipgloss.ANSIColor(242)).
			Faint(true),

		SortIndicator: lipgloss.NewStyle().
			Foreground(lipgloss.ANSIColor(242)).
			Faint(true),

		SearchPrompt: lipgloss.NewStyle().
			Foreground(lipgloss.ANSIColor(3)).
			Bold(true),

		SearchInput: lipgloss.NewStyle().
			Foreground(lipgloss.ANSIColor(15)),

		SearchCursor: lipgloss.NewStyle().
			Foreground(lipgloss.ANSIColor(15)).
			Bold(true),

		GitBranch: lipgloss.NewStyle().
			Foreground(lipgloss.ANSIColor(5)).
			Bold(true),

		GitModified: lipgloss.NewStyle().
			Foreground(lipgloss.ANSIColor(3)),

		GitStaged: lipgloss.NewStyle().
			Foreground(lipgloss.ANSIColor(2)),

		GitUntracked: lipgloss.NewStyle().
			Foreground(lipgloss.ANSIColor(1)),

		GitConflict: lipgloss.NewStyle().
			Foreground(lipgloss.ANSIColor(1)).
			Bold(true),
	}
}
