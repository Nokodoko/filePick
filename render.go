package main

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

// renderPathBar renders the top path bar with truncation and optional git branch.
func renderPathBar(cwd string, width int, styles Styles, gitInfo *GitInfo) string {
	var branchStr string
	if gitInfo != nil && gitInfo.IsRepo {
		branchStr = " " + styles.GitBranch.Render(gitBranchIcon+" "+gitInfo.Branch)
	}

	// Account for branch indicator width when truncating path
	branchWidth := lipgloss.Width(branchStr)
	path := truncatePath(cwd, width-4-branchWidth)
	return styles.PathBar.Render("  "+path) + branchStr
}

// renderSeparator renders a horizontal rule separator line.
func renderSeparator(width int, styles Styles) string {
	return styles.Separator.Render(strings.Repeat("─", width))
}

// renderEntry renders a single directory entry line with icon, name, and git status.
func renderEntry(entry DirEntry, selected bool, width int, noIcons bool, styles Styles, gitStatus GitStatus) string {
	icon, iconColor := resolveIcon(&entry, noIcons)

	// Build the icon part with color
	var iconStr string
	if noColor() {
		iconStr = icon
	} else {
		iconStyle := lipgloss.NewStyle().Foreground(iconColor)
		iconStr = iconStyle.Render(icon)
	}

	// Build the name part with appropriate style
	var nameStr string
	name := entry.Name

	// Add symlink indicator
	if entry.IsSymlink {
		name += " ->"
	}

	// Add directory trailing slash for clarity
	if entry.IsDir {
		name += "/"
	}

	switch {
	case entry.IsHidden:
		nameStr = styles.HiddenEntry.Render(name)
	case entry.IsDir:
		nameStr = styles.DirName.Render(name)
	case entry.IsExecutable:
		nameStr = styles.ExecName.Render(name)
	case entry.IsSymlink:
		nameStr = styles.SymlinkName.Render(name)
	default:
		nameStr = styles.FileName.Render(name)
	}

	// Build git status indicator
	var gitIndicator string
	if gitStatus != GitNone {
		if sym, ok := gitStatusIcons[gitStatus]; ok {
			var gitStyle lipgloss.Style
			switch gitStatus {
			case GitModified:
				gitStyle = styles.GitModified
			case GitStaged, GitAdded:
				gitStyle = styles.GitStaged
			case GitUntracked:
				gitStyle = styles.GitUntracked
			case GitConflict:
				gitStyle = styles.GitConflict
			default:
				gitStyle = styles.GitModified
			}
			if noColor() {
				gitIndicator = " " + sym
			} else {
				gitIndicator = " " + gitStyle.Render(sym)
			}
		}
	}

	line := fmt.Sprintf("  %s  %s%s", iconStr, nameStr, gitIndicator)

	if selected {
		// Apply selection bar styling across full width
		var selGit string
		if gitStatus != GitNone {
			if sym, ok := gitStatusIcons[gitStatus]; ok {
				selGit = " " + sym
			}
		}
		line = styles.SelectionBar.Width(width).Render(
			fmt.Sprintf("  %s  %s%s", icon, name, selGit),
		)
	}

	return line
}

// renderFileList renders the scrollable list of entries.
func renderFileList(entries []DirEntry, cursor, viewportOffset, height, width int, noIcons bool, styles Styles, gitInfo *GitInfo) string {
	if len(entries) == 0 {
		empty := styles.HiddenEntry.Render("  (empty)")
		return empty
	}

	var lines []string
	end := viewportOffset + height
	if end > len(entries) {
		end = len(entries)
	}

	for i := viewportOffset; i < end; i++ {
		selected := i == cursor
		gs := lookupGitStatus(gitInfo, entries[i].Name)
		lines = append(lines, renderEntry(entries[i], selected, width, noIcons, styles, gs))
	}

	return strings.Join(lines, "\n")
}

// renderStatusBar renders the bottom status bar with item count and sort mode.
func renderStatusBar(count int, sortMode SortMode, width int, styles Styles) string {
	left := styles.ItemCount.Render(fmt.Sprintf("  %d items", count))
	right := styles.SortIndicator.Render(fmt.Sprintf("  %s  ", sortMode.String()))

	// Calculate padding to right-align the sort indicator
	padding := width - lipgloss.Width(left) - lipgloss.Width(right)
	if padding < 0 {
		padding = 0
	}

	return left + strings.Repeat(" ", padding) + right
}

// renderSearchBar renders the search input bar.
func renderSearchBar(query string, width int, styles Styles) string {
	prompt := styles.SearchPrompt.Render("  / ")
	input := styles.SearchInput.Render(query)
	cursor := styles.SearchCursor.Render("_")
	return prompt + input + cursor
}

// renderView composes the full TUI view.
func renderView(cwd string, entries []DirEntry, cursor, viewportOffset, width, height int,
	sortMode SortMode, noIcons bool, errMsg string, styles Styles,
	searchMode bool, searchQuery string, gitInfo *GitInfo) string {

	var sections []string

	// Path bar (1 line)
	sections = append(sections, renderPathBar(cwd, width, styles, gitInfo))

	// Separator (1 line)
	sections = append(sections, renderSeparator(width, styles))

	// File list (height - 4 lines: 1 path + 1 sep + 1 sep + 1 status)
	listHeight := height - 4
	if searchMode {
		listHeight-- // search bar takes 1 line
	}
	if listHeight < 1 {
		listHeight = 1
	}
	sections = append(sections, renderFileList(entries, cursor, viewportOffset, listHeight, width, noIcons, styles, gitInfo))

	// Separator (1 line)
	sections = append(sections, renderSeparator(width, styles))

	// Search bar or status bar (1 line)
	if searchMode {
		sections = append(sections, renderSearchBar(searchQuery, width, styles))
	} else if errMsg != "" {
		sections = append(sections, styles.ErrorMsg.Render("  "+errMsg))
	} else {
		sections = append(sections, renderStatusBar(len(entries), sortMode, width, styles))
	}

	return strings.Join(sections, "\n")
}

// colorToLipgloss converts a color.Color to a lipgloss-compatible color.
// This is a helper for icon coloring.
func colorToLipgloss(c color.Color) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(c)
}
