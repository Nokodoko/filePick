package main

import (
	"testing"
)

func TestParseGitXY(t *testing.T) {
	tests := []struct {
		x, y   byte
		expect GitStatus
	}{
		{'?', '?', GitUntracked},
		{'!', '!', GitIgnored},
		{' ', 'M', GitModified},
		{'M', ' ', GitStaged},
		{'A', ' ', GitAdded},
		{'D', ' ', GitDeleted},
		{'R', ' ', GitRenamed},
		{' ', 'D', GitDeleted},
		{'U', 'U', GitConflict},
		{'A', 'A', GitConflict},
		{'D', 'D', GitConflict},
		{' ', ' ', GitNone},
	}

	for _, tt := range tests {
		result := parseGitXY(tt.x, tt.y)
		if result != tt.expect {
			t.Errorf("parseGitXY(%q, %q) = %d, want %d", tt.x, tt.y, result, tt.expect)
		}
	}
}

func TestMergeGitStatus(t *testing.T) {
	tests := []struct {
		a, b   GitStatus
		expect GitStatus
	}{
		{GitNone, GitModified, GitModified},
		{GitModified, GitUntracked, GitModified},
		{GitUntracked, GitStaged, GitUntracked},
		{GitConflict, GitModified, GitConflict},
		{GitStaged, GitConflict, GitConflict},
		{GitNone, GitNone, GitNone},
	}

	for _, tt := range tests {
		result := mergeGitStatus(tt.a, tt.b)
		if result != tt.expect {
			t.Errorf("mergeGitStatus(%d, %d) = %d, want %d", tt.a, tt.b, result, tt.expect)
		}
	}
}

func TestLookupGitStatus(t *testing.T) {
	info := &GitInfo{
		IsRepo: true,
		FileStatus: map[string]GitStatus{
			"main.go":  GitModified,
			"new.go":   GitUntracked,
			"staged.go": GitStaged,
		},
	}

	if s := lookupGitStatus(info, "main.go"); s != GitModified {
		t.Errorf("expected GitModified for main.go, got %d", s)
	}
	if s := lookupGitStatus(info, "new.go"); s != GitUntracked {
		t.Errorf("expected GitUntracked for new.go, got %d", s)
	}
	if s := lookupGitStatus(info, "clean.go"); s != GitNone {
		t.Errorf("expected GitNone for clean.go, got %d", s)
	}

	// Nil info should return GitNone
	if s := lookupGitStatus(nil, "any"); s != GitNone {
		t.Errorf("expected GitNone for nil info, got %d", s)
	}

	// Non-repo should return GitNone
	nonRepo := &GitInfo{IsRepo: false}
	if s := lookupGitStatus(nonRepo, "any"); s != GitNone {
		t.Errorf("expected GitNone for non-repo, got %d", s)
	}
}

func TestGitStatusIcons(t *testing.T) {
	// Verify all non-None statuses have icons
	statuses := []GitStatus{GitModified, GitStaged, GitUntracked, GitConflict, GitRenamed, GitDeleted, GitAdded}
	for _, s := range statuses {
		if _, ok := gitStatusIcons[s]; !ok {
			t.Errorf("missing icon for git status %d", s)
		}
	}
}

func TestDetectGitInfoInRepo(t *testing.T) {
	// This test runs in the filePick repo itself, so it should detect git
	info := detectGitInfo(".")
	if !info.IsRepo {
		t.Skip("not running in a git repository")
	}
	if info.Branch == "" {
		t.Error("expected non-empty branch name in git repo")
	}
}
