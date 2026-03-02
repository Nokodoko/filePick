package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/fsnotify/fsnotify"
)

// AppModel is the bubbletea model for filePick.
type AppModel struct {
	// Directory state
	cwd             string
	entries         []DirEntry
	filteredEntries []DirEntry

	// UI state
	cursor         int
	viewportOffset int
	termWidth      int
	termHeight     int

	// Settings
	showHidden bool
	sortMode   SortMode
	dirsFirst  bool
	noIcons    bool

	// Filesystem watcher
	watcher *fsnotify.Watcher
	program *tea.Program

	// Display
	styles Styles

	// Error state
	err string

	// Selected file path (printed to stdout on quit)
	selected string

	// Whether we quit (vs selected a file)
	quitting bool

	// Search mode state
	searchMode    bool
	searchQuery   string
	searchResults []DirEntry
}

// NewAppModel creates a new AppModel with the given configuration.
func NewAppModel(cwd string, showHidden bool, sortMode SortMode, noIcons bool) AppModel {
	return AppModel{
		cwd:        cwd,
		showHidden: showHidden,
		sortMode:   sortMode,
		dirsFirst:  true,
		noIcons:    noIcons,
		styles:     NewStyles(),
		termWidth:  80,
		termHeight: 24,
	}
}

// Init implements tea.Model.
func (m *AppModel) Init() tea.Cmd {
	m.loadDir()
	return nil
}

// loadDir reads the current directory and updates entries.
func (m *AppModel) loadDir() {
	entries, err := readDir(m.cwd)
	if err != nil {
		m.err = fmt.Sprintf("cannot read directory '%s': %s", m.cwd, err)
		m.entries = nil
		m.filteredEntries = nil
		return
	}

	m.err = ""
	m.entries = entries
	sortEntries(m.entries, m.sortMode, m.dirsFirst)
	m.filteredEntries = filterHidden(m.entries, m.showHidden)
	m.clampCursor()
}

// clampCursor ensures the cursor is within valid bounds.
func (m *AppModel) clampCursor() {
	entries := m.activeEntries()
	if len(entries) == 0 {
		m.cursor = 0
		m.viewportOffset = 0
		return
	}
	if m.cursor >= len(entries) {
		m.cursor = len(entries) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}

	// Adjust viewport — account for search bar taking 1 extra line
	listHeight := m.termHeight - 4
	if m.searchMode {
		listHeight-- // search input bar takes 1 line
	}
	if listHeight < 1 {
		listHeight = 1
	}
	if m.cursor < m.viewportOffset {
		m.viewportOffset = m.cursor
	}
	if m.cursor >= m.viewportOffset+listHeight {
		m.viewportOffset = m.cursor - listHeight + 1
	}
}

// runSearch executes rg --files in the current directory and filters results
// by the search query. Returns matching DirEntry items.
func (m *AppModel) runSearch() {
	if m.searchQuery == "" {
		m.searchResults = nil
		return
	}

	// Run rg --files to get all files, then filter by query
	cmd := exec.Command("rg", "--files", "--hidden", "--glob", "!.git", m.cwd)
	out, err := cmd.Output()
	if err != nil {
		// If rg is not found or errors, fall back to empty results
		m.searchResults = nil
		return
	}

	query := strings.ToLower(m.searchQuery)
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")

	var results []DirEntry
	for _, line := range lines {
		if line == "" {
			continue
		}

		// Get the relative path for matching
		relPath := line
		if strings.HasPrefix(line, m.cwd) {
			relPath = strings.TrimPrefix(line, m.cwd+"/")
		}

		// Fuzzy match: check if the query characters appear in order
		if !fuzzyMatch(strings.ToLower(relPath), query) {
			continue
		}

		// Build a DirEntry for the result
		info, statErr := os.Lstat(line)
		if statErr != nil {
			continue
		}

		entry := DirEntry{
			Name:     relPath,
			Path:     line,
			IsDir:    info.IsDir(),
			IsHidden: len(filepath.Base(line)) > 0 && filepath.Base(line)[0] == '.',
			Size:     info.Size(),
			ModTime:  info.ModTime(),
		}

		entry.IsSymlink = info.Mode()&os.ModeSymlink != 0
		entry.IsExecutable = !entry.IsDir && info.Mode()&0o111 != 0

		results = append(results, entry)

		// Cap results to avoid overwhelming the UI
		if len(results) >= 100 {
			break
		}
	}

	m.searchResults = results
}

// fuzzyMatch checks if all characters in pattern appear in str in order.
func fuzzyMatch(str, pattern string) bool {
	pi := 0
	for si := 0; si < len(str) && pi < len(pattern); si++ {
		if str[si] == pattern[pi] {
			pi++
		}
	}
	return pi == len(pattern)
}

// activeEntries returns the entries currently being displayed,
// depending on whether search mode is active.
func (m *AppModel) activeEntries() []DirEntry {
	if m.searchMode && m.searchQuery != "" {
		return m.searchResults
	}
	return m.filteredEntries
}

// Update implements tea.Model.
func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height
		m.clampCursor()
		return m, nil

	case fsEventMsg:
		// Filesystem change detected — re-read directory
		if !m.searchMode {
			m.loadDir()
		}
		return m, nil

	case tea.KeyPressMsg:
		// Handle search mode input
		if m.searchMode {
			return m.updateSearchMode(msg)
		}

		action := resolveKey(msg)
		switch action {
		case keyQuit:
			m.quitting = true
			return m, tea.Quit

		case keySearch:
			m.searchMode = true
			m.searchQuery = ""
			m.searchResults = nil
			m.cursor = 0
			m.viewportOffset = 0
			return m, nil

		case keyCursorDown:
			if m.cursor < len(m.filteredEntries)-1 {
				m.cursor++
				m.clampCursor()
			}
			return m, nil

		case keyCursorUp:
			if m.cursor > 0 {
				m.cursor--
				m.clampCursor()
			}
			return m, nil

		case keyJumpTop:
			m.cursor = 0
			m.clampCursor()
			return m, nil

		case keyJumpBottom:
			if len(m.filteredEntries) > 0 {
				m.cursor = len(m.filteredEntries) - 1
			}
			m.clampCursor()
			return m, nil

		case keyEnterDir:
			if len(m.filteredEntries) == 0 {
				return m, nil
			}
			entry := m.filteredEntries[m.cursor]
			if entry.IsDir {
				// Navigate into directory
				oldCwd := m.cwd
				m.cwd = entry.Path
				m.cursor = 0
				m.viewportOffset = 0
				m.loadDir()
				// Switch watcher
				if m.watcher != nil {
					_ = switchWatch(m.watcher, oldCwd, m.cwd)
				}
				return m, nil
			}
			// File selected — print path and quit
			m.selected = entry.Path
			return m, tea.Quit

		case keyParentDir:
			parent := filepath.Dir(m.cwd)
			if parent == m.cwd {
				// Already at root
				return m, nil
			}
			oldCwd := m.cwd
			m.cwd = parent
			m.cursor = 0
			m.viewportOffset = 0
			m.loadDir()
			// Switch watcher
			if m.watcher != nil {
				_ = switchWatch(m.watcher, oldCwd, m.cwd)
			}
			return m, nil

		case keyToggleHidden:
			m.showHidden = !m.showHidden
			m.filteredEntries = filterHidden(m.entries, m.showHidden)
			m.clampCursor()
			return m, nil

		case keyCycleSort:
			m.sortMode = NextSortMode(m.sortMode)
			sortEntries(m.entries, m.sortMode, m.dirsFirst)
			m.filteredEntries = filterHidden(m.entries, m.showHidden)
			m.clampCursor()
			return m, nil
		}
	}

	return m, nil
}

// updateSearchMode handles key input while in search mode.
func (m *AppModel) updateSearchMode(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.Key()

	// Ctrl+C always quits
	if key.Code == 'c' && key.Mod == tea.ModCtrl {
		m.quitting = true
		return m, tea.Quit
	}

	switch key.Code {
	case tea.KeyEsc:
		// Exit search mode, restore normal view
		m.searchMode = false
		m.searchQuery = ""
		m.searchResults = nil
		m.cursor = 0
		m.viewportOffset = 0
		return m, nil

	case tea.KeyEnter:
		// Select the highlighted search result
		entries := m.activeEntries()
		if len(entries) == 0 {
			return m, nil
		}
		entry := entries[m.cursor]
		if entry.IsDir {
			// Navigate into directory and exit search
			oldCwd := m.cwd
			m.cwd = entry.Path
			m.searchMode = false
			m.searchQuery = ""
			m.searchResults = nil
			m.cursor = 0
			m.viewportOffset = 0
			m.loadDir()
			if m.watcher != nil {
				_ = switchWatch(m.watcher, oldCwd, m.cwd)
			}
			return m, nil
		}
		// File selected — print path and quit
		m.selected = entry.Path
		return m, tea.Quit

	case tea.KeyBackspace:
		if len(m.searchQuery) > 0 {
			m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
			m.runSearch()
			m.cursor = 0
			m.viewportOffset = 0
		}
		if m.searchQuery == "" {
			// If query is empty after backspace, exit search mode
			m.searchMode = false
			m.searchResults = nil
			m.cursor = 0
			m.viewportOffset = 0
		}
		return m, nil

	case tea.KeyDown:
		entries := m.activeEntries()
		if m.cursor < len(entries)-1 {
			m.cursor++
			m.clampCursor()
		}
		return m, nil

	case tea.KeyUp:
		if m.cursor > 0 {
			m.cursor--
			m.clampCursor()
		}
		return m, nil

	default:
		// Printable character — append to search query
		s := msg.String()
		if len(s) == 1 && s[0] >= 32 && s[0] <= 126 {
			m.searchQuery += s
			m.runSearch()
			m.cursor = 0
			m.viewportOffset = 0
		}
		return m, nil
	}
}

// View implements tea.Model.
func (m *AppModel) View() tea.View {
	v := tea.NewView(renderView(
		m.cwd,
		m.activeEntries(),
		m.cursor,
		m.viewportOffset,
		m.termWidth,
		m.termHeight,
		m.sortMode,
		m.noIcons,
		m.err,
		m.styles,
		m.searchMode,
		m.searchQuery,
	))
	v.AltScreen = true
	return v
}

// SetProgram sets the tea.Program reference for the watcher.
func (m *AppModel) SetProgram(p *tea.Program) {
	m.program = p
}

// StartWatcher starts the filesystem watcher.
func (m *AppModel) StartWatcher() error {
	if m.program == nil {
		return fmt.Errorf("program not set")
	}
	watcher, err := startWatcher(m.cwd, m.program)
	if err != nil {
		// Fall back gracefully — watcher is optional
		fmt.Fprintf(os.Stderr, "fp: warning: filesystem watcher unavailable: %s\n", err)
		return nil
	}
	m.watcher = watcher
	return nil
}

// Close cleans up resources.
func (m *AppModel) Close() {
	if m.watcher != nil {
		m.watcher.Close()
	}
}
