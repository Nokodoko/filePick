package main

import (
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

// startWatcher creates an fsnotify watcher for the given directory and returns
// a tea.Cmd that listens for events and sends fsEventMsg through the program.
// The watcher is non-recursive — it only watches the specified directory.
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

	go watchLoop(watcher, p)

	return watcher, nil
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
			// We care about create, delete, rename, and chmod events
			_ = event
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

// switchWatch removes the old watch and adds a new one for the given directory.
func switchWatch(watcher *fsnotify.Watcher, oldDir, newDir string) error {
	if oldDir != "" {
		_ = watcher.Remove(oldDir)
	}
	return watcher.Add(newDir)
}
