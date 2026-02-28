package main

import (
	"fmt"
	"os"
	"path/filepath"

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
	if len(m.filteredEntries) == 0 {
		m.cursor = 0
		m.viewportOffset = 0
		return
	}
	if m.cursor >= len(m.filteredEntries) {
		m.cursor = len(m.filteredEntries) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}

	// Adjust viewport
	listHeight := m.termHeight - 4
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
		m.loadDir()
		return m, nil

	case tea.KeyPressMsg:
		action := resolveKey(msg)
		switch action {
		case keyQuit:
			m.quitting = true
			return m, tea.Quit

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

// View implements tea.Model.
func (m *AppModel) View() tea.View {
	v := tea.NewView(renderView(
		m.cwd,
		m.filteredEntries,
		m.cursor,
		m.viewportOffset,
		m.termWidth,
		m.termHeight,
		m.sortMode,
		m.noIcons,
		m.err,
		m.styles,
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
