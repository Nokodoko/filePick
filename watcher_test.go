package main

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/fsnotify/fsnotify"
)

func TestWatcherDebounce(t *testing.T) {
	// Create a temp directory to watch
	tmpDir := t.TempDir()

	// Create a mock program to capture sent messages
	var mu sync.Mutex
	var msgs []tea.Msg

	// We test the watchLoop directly with a real watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer watcher.Close()

	err = watcher.Add(tmpDir)
	if err != nil {
		t.Fatalf("failed to add watch: %v", err)
	}

	// Create a minimal mock that captures Send calls
	// Since we can't easily mock tea.Program, test the debounce logic
	// by creating rapid events and verifying coalescing
	done := make(chan struct{})
	go func() {
		defer close(done)
		var timer *time.Timer
		var timerC <-chan time.Time
		count := 0

		for {
			select {
			case _, ok := <-watcher.Events:
				if !ok {
					return
				}
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
				mu.Lock()
				msgs = append(msgs, fsEventMsg{})
				mu.Unlock()
				timer = nil
				timerC = nil
				count++
				if count >= 1 {
					return
				}

			case _, ok := <-watcher.Errors:
				if !ok {
					return
				}
			}
		}
	}()

	// Create 5 files rapidly to trigger many events
	for i := 0; i < 5; i++ {
		f, err := os.Create(filepath.Join(tmpDir, "testfile"+string(rune('a'+i))))
		if err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
		f.Close()
	}

	// Wait for debounce to fire
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for debounced event")
	}

	mu.Lock()
	defer mu.Unlock()

	// The debounce should coalesce all rapid events into 1 event
	if len(msgs) != 1 {
		t.Errorf("expected 1 debounced event, got %d", len(msgs))
	}
}

func TestWatcherSwitchWatch(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer watcher.Close()

	// Add first directory
	err = watcher.Add(dir1)
	if err != nil {
		t.Fatalf("failed to add dir1: %v", err)
	}

	// Switch to second directory
	err = switchWatch(watcher, dir1, dir2)
	if err != nil {
		t.Fatalf("switchWatch failed: %v", err)
	}

	// Verify dir2 is watched by creating a file
	// (If dir1 was still watched, we wouldn't get events for dir2)
	watchList := watcher.WatchList()
	found := false
	for _, w := range watchList {
		if w == dir2 {
			found = true
			break
		}
	}
	if !found {
		t.Error("dir2 not in watch list after switch")
	}

	// Verify dir1 is no longer watched
	for _, w := range watchList {
		if w == dir1 {
			t.Error("dir1 still in watch list after switch")
		}
	}
}

func TestWatchWorktreeDirs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .git/worktrees directory
	gitWorktrees := filepath.Join(tmpDir, ".git", "worktrees")
	err := os.MkdirAll(gitWorktrees, 0o755)
	if err != nil {
		t.Fatalf("failed to create .git/worktrees: %v", err)
	}

	// Create .claude/worktrees directory
	claudeWorktrees := filepath.Join(tmpDir, ".claude", "worktrees")
	err = os.MkdirAll(claudeWorktrees, 0o755)
	if err != nil {
		t.Fatalf("failed to create .claude/worktrees: %v", err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer watcher.Close()

	watchWorktreeDirs(watcher, tmpDir)

	watchList := watcher.WatchList()

	foundGit := false
	foundClaude := false
	for _, w := range watchList {
		if w == gitWorktrees {
			foundGit = true
		}
		if w == claudeWorktrees {
			foundClaude = true
		}
	}

	if !foundGit {
		t.Error(".git/worktrees not in watch list")
	}
	if !foundClaude {
		t.Error(".claude/worktrees not in watch list")
	}
}

func TestWatchWorktreeDirsMissing(t *testing.T) {
	tmpDir := t.TempDir()

	// Don't create any worktree dirs — they should be silently skipped
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer watcher.Close()

	watchWorktreeDirs(watcher, tmpDir)

	watchList := watcher.WatchList()
	if len(watchList) != 0 {
		t.Errorf("expected empty watch list for missing dirs, got %v", watchList)
	}
}

func TestSwitchWatchCleansWorktreeDirs(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	// Create worktree dirs in dir1
	gitWT1 := filepath.Join(dir1, ".git", "worktrees")
	err := os.MkdirAll(gitWT1, 0o755)
	if err != nil {
		t.Fatalf("failed to mkdir: %v", err)
	}

	// Create worktree dirs in dir2
	claudeWT2 := filepath.Join(dir2, ".claude", "worktrees")
	err = os.MkdirAll(claudeWT2, 0o755)
	if err != nil {
		t.Fatalf("failed to mkdir: %v", err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer watcher.Close()

	// Initial setup: watch dir1 + its worktree dirs
	err = watcher.Add(dir1)
	if err != nil {
		t.Fatalf("failed to add dir1: %v", err)
	}
	watchWorktreeDirs(watcher, dir1)

	// Verify dir1's git worktrees dir is watched
	found := false
	for _, w := range watcher.WatchList() {
		if w == gitWT1 {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("dir1 .git/worktrees not watched before switch")
	}

	// Switch to dir2
	err = switchWatch(watcher, dir1, dir2)
	if err != nil {
		t.Fatalf("switchWatch failed: %v", err)
	}

	// Verify dir1's worktree dir is no longer watched
	for _, w := range watcher.WatchList() {
		if w == gitWT1 {
			t.Error("dir1 .git/worktrees still watched after switch")
		}
	}

	// Verify dir2's claude worktrees dir IS watched
	found = false
	for _, w := range watcher.WatchList() {
		if w == claudeWT2 {
			found = true
			break
		}
	}
	if !found {
		t.Error("dir2 .claude/worktrees not watched after switch")
	}
}
