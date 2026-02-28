package main

import (
	"strings"
	"testing"
)

func TestRenderPathBar(t *testing.T) {
	styles := NewStyles()
	result := renderPathBar("/home/user/Projects", 80, styles)
	if result == "" {
		t.Error("renderPathBar returned empty string")
	}
}

func TestRenderSeparator(t *testing.T) {
	styles := NewStyles()
	result := renderSeparator(40, styles)
	if result == "" {
		t.Error("renderSeparator returned empty string")
	}
}

func TestRenderEntryDir(t *testing.T) {
	styles := NewStyles()
	entry := DirEntry{Name: "src", IsDir: true}
	result := renderEntry(entry, false, 80, false, styles)
	if !strings.Contains(result, "src") {
		t.Errorf("expected entry to contain 'src', got %q", result)
	}
}

func TestRenderEntryFile(t *testing.T) {
	styles := NewStyles()
	entry := DirEntry{Name: "main.go", IsDir: false}
	result := renderEntry(entry, false, 80, false, styles)
	if !strings.Contains(result, "main.go") {
		t.Errorf("expected entry to contain 'main.go', got %q", result)
	}
}

func TestRenderEntrySelected(t *testing.T) {
	styles := NewStyles()
	entry := DirEntry{Name: "selected.txt", IsDir: false}
	result := renderEntry(entry, true, 80, false, styles)
	if !strings.Contains(result, "selected.txt") {
		t.Errorf("expected selected entry to contain 'selected.txt', got %q", result)
	}
}

func TestRenderEntryNoIcons(t *testing.T) {
	styles := NewStyles()
	entry := DirEntry{Name: "src", IsDir: true}
	result := renderEntry(entry, false, 80, true, styles)
	if !strings.Contains(result, ">") {
		t.Errorf("expected ASCII dir icon '>' in no-icons mode, got %q", result)
	}
}

func TestRenderEntrySymlink(t *testing.T) {
	styles := NewStyles()
	entry := DirEntry{Name: "link", IsDir: false, IsSymlink: true}
	result := renderEntry(entry, false, 80, false, styles)
	if !strings.Contains(result, "->") {
		t.Errorf("expected symlink indicator '->', got %q", result)
	}
}

func TestRenderFileListEmpty(t *testing.T) {
	styles := NewStyles()
	result := renderFileList(nil, 0, 0, 10, 80, false, styles)
	if !strings.Contains(result, "empty") {
		t.Errorf("expected '(empty)' for empty list, got %q", result)
	}
}

func TestRenderFileList(t *testing.T) {
	styles := NewStyles()
	entries := []DirEntry{
		{Name: "dir1", IsDir: true},
		{Name: "file1.go", IsDir: false},
		{Name: "file2.rs", IsDir: false},
	}
	result := renderFileList(entries, 0, 0, 10, 80, false, styles)
	if !strings.Contains(result, "dir1") {
		t.Error("expected file list to contain 'dir1'")
	}
	if !strings.Contains(result, "file1.go") {
		t.Error("expected file list to contain 'file1.go'")
	}
}

func TestRenderStatusBar(t *testing.T) {
	styles := NewStyles()
	result := renderStatusBar(14, SortByName, 80, styles)
	if !strings.Contains(result, "14 items") {
		t.Errorf("expected '14 items' in status bar, got %q", result)
	}
	if !strings.Contains(result, "name") {
		t.Errorf("expected 'name' sort indicator, got %q", result)
	}
}

func TestRenderViewComplete(t *testing.T) {
	styles := NewStyles()
	entries := []DirEntry{
		{Name: "src", IsDir: true},
		{Name: "main.go", IsDir: false},
	}
	result := renderView("/home/user/project", entries, 0, 0, 80, 24, SortByName, false, "", styles)
	if result == "" {
		t.Error("renderView returned empty string")
	}
	// Should contain all sections
	if !strings.Contains(result, "project") {
		t.Error("expected path bar in view")
	}
	if !strings.Contains(result, "src") {
		t.Error("expected file entries in view")
	}
	if !strings.Contains(result, "2 items") {
		t.Error("expected status bar in view")
	}
}

func TestRenderViewWithError(t *testing.T) {
	styles := NewStyles()
	result := renderView("/tmp", nil, 0, 0, 80, 24, SortByName, false, "permission denied", styles)
	if !strings.Contains(result, "permission denied") {
		t.Errorf("expected error message in view, got %q", result)
	}
}
