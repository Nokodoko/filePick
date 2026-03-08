package main

import (
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/fsnotify/fsnotify"
)

// fsEventMsg is sent when a filesystem change is detected.
// The bubbletea Update function handles this by re-reading the directory.
type fsEventMsg struct{}

// watcherErrorMsg is sent when the watcher encounters an error.
type watcherErrorMsg struct {
	err error
}

// debounceInterval is the time to wait before coalescing rapid events.
const debounceInterval = 50 * time.Millisecond

// pollInterval is the fallback poll interval for catching events that
// inotify misses (e.g., changes in subdirectories not directly watched).
const pollInterval = 5 * time.Second

// worktreeDirs are subdirectory paths (relative to the watched root) that
// should be watched in addition to the current directory. These are where
// git worktrees and claude worktrees appear/disappear during swarm runs.
var worktreeDirs = []string{
	".git/worktrees",
	".claude/worktrees",
}

// startWatcher creates an fsnotify watcher for the given directory, adds
// targeted watches for worktree directories, starts a fallback poll timer,
// starts a SIGUSR1 listener, and returns the watcher.
func startWatcher(dir string, p *tea.Program) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	err = watcher.Add(dir)
	if err != nil {
		watcher.Close()
		return nil, err
	}

	// Watch worktree directories if they exist.
	watchWorktreeDirs(watcher, dir)

	go watchLoop(watcher, p)
	go pollLoop(p)
	go signalLoop(p)

	return watcher, nil
}

// watchWorktreeDirs adds inotify watches for known worktree directories
// relative to root. Missing directories are silently skipped — they may
// not exist yet and the fallback poll will catch later changes.
func watchWorktreeDirs(watcher *fsnotify.Watcher, root string) {
	for _, rel := range worktreeDirs {
		dir := filepath.Join(root, rel)
		info, err := os.Stat(dir)
		if err != nil || !info.IsDir() {
			continue
		}
		// Best effort — ignore errors for non-existent dirs.
		_ = watcher.Add(dir)
	}
}

// watchLoop reads from the watcher channels and sends debounced events
// to the bubbletea program. It coalesces rapid events using a timer.
func watchLoop(watcher *fsnotify.Watcher, p *tea.Program) {
	var timer *time.Timer
	var timerC <-chan time.Time

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			// On Create events for directories, try to add a watch on new
			// worktree directories that may have just appeared.
			if event.Has(fsnotify.Create) {
				info, err := os.Stat(event.Name)
				if err == nil && info.IsDir() {
					// Check if it's a worktree-related dir worth watching.
					for _, rel := range worktreeDirs {
						if filepath.Base(filepath.Dir(event.Name)) == filepath.Base(filepath.Dir(rel)) {
							_ = watcher.Add(event.Name)
							break
						}
					}
				}
			}
			// Reset or start the debounce timer
			if timer == nil {
				timer = time.NewTimer(debounceInterval)
				timerC = timer.C
			} else {
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				timer.Reset(debounceInterval)
			}

		case <-timerC:
			// Debounce timer fired — send a single event
			p.Send(fsEventMsg{})
			timer = nil
			timerC = nil

		case _, ok := <-watcher.Errors:
			if !ok {
				return
			}
			// We could send an error msg, but for robustness we just continue
		}
	}
}

// pollLoop sends a periodic fsEventMsg as a fallback for changes that
// inotify misses (non-recursive subdirectory changes, NFS, etc.).
func pollLoop(p *tea.Program) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for range ticker.C {
		p.Send(fsEventMsg{})
	}
}

// signalLoop listens for SIGUSR1 and sends an fsEventMsg when received.
// This allows external processes (e.g., cmdr's dashboard) to force an
// immediate refresh of the file listing.
func signalLoop(p *tea.Program) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGUSR1)
	for range sigCh {
		p.Send(fsEventMsg{})
	}
}

// switchWatch removes old watches (including worktree subdirs) and adds
// new ones for the given directory and its worktree subdirs.
func switchWatch(watcher *fsnotify.Watcher, oldDir, newDir string) error {
	// Remove old directory and its worktree subdirs.
	if oldDir != "" {
		_ = watcher.Remove(oldDir)
		for _, rel := range worktreeDirs {
			_ = watcher.Remove(filepath.Join(oldDir, rel))
		}
	}

	// Add new directory.
	err := watcher.Add(newDir)
	if err != nil {
		return err
	}

	// Watch worktree dirs under the new directory.
	watchWorktreeDirs(watcher, newDir)

	return nil
}
