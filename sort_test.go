package main

import (
	"testing"
	"time"
)

func TestSortByName(t *testing.T) {
	entries := []DirEntry{
		{Name: "zebra.txt", IsDir: false},
		{Name: "alpha.go", IsDir: false},
		{Name: "middle.rs", IsDir: false},
	}
	sortEntries(entries, SortByName, false)

	expected := []string{"alpha.go", "middle.rs", "zebra.txt"}
	for i, e := range entries {
		if e.Name != expected[i] {
			t.Errorf("position %d: expected %s, got %s", i, expected[i], e.Name)
		}
	}
}

func TestSortByNameCaseInsensitive(t *testing.T) {
	entries := []DirEntry{
		{Name: "Zebra.txt", IsDir: false},
		{Name: "alpha.go", IsDir: false},
		{Name: "Beta.rs", IsDir: false},
	}
	sortEntries(entries, SortByName, false)

	expected := []string{"alpha.go", "Beta.rs", "Zebra.txt"}
	for i, e := range entries {
		if e.Name != expected[i] {
			t.Errorf("position %d: expected %s, got %s", i, expected[i], e.Name)
		}
	}
}

func TestSortByTime(t *testing.T) {
	now := time.Now()
	entries := []DirEntry{
		{Name: "old.txt", IsDir: false, ModTime: now.Add(-2 * time.Hour)},
		{Name: "new.txt", IsDir: false, ModTime: now},
		{Name: "mid.txt", IsDir: false, ModTime: now.Add(-1 * time.Hour)},
	}
	sortEntries(entries, SortByTime, false)

	// Newest first
	expected := []string{"new.txt", "mid.txt", "old.txt"}
	for i, e := range entries {
		if e.Name != expected[i] {
			t.Errorf("position %d: expected %s, got %s", i, expected[i], e.Name)
		}
	}
}

func TestSortBySize(t *testing.T) {
	entries := []DirEntry{
		{Name: "small.txt", IsDir: false, Size: 100},
		{Name: "large.txt", IsDir: false, Size: 10000},
		{Name: "medium.txt", IsDir: false, Size: 1000},
	}
	sortEntries(entries, SortBySize, false)

	// Largest first
	expected := []string{"large.txt", "medium.txt", "small.txt"}
	for i, e := range entries {
		if e.Name != expected[i] {
			t.Errorf("position %d: expected %s, got %s", i, expected[i], e.Name)
		}
	}
}

func TestSortDirsFirst(t *testing.T) {
	entries := []DirEntry{
		{Name: "file_a.txt", IsDir: false},
		{Name: "dir_b", IsDir: true},
		{Name: "file_c.go", IsDir: false},
		{Name: "dir_a", IsDir: true},
	}
	sortEntries(entries, SortByName, true)

	expected := []string{"dir_a", "dir_b", "file_a.txt", "file_c.go"}
	for i, e := range entries {
		if e.Name != expected[i] {
			t.Errorf("position %d: expected %s, got %s", i, expected[i], e.Name)
		}
	}
}

func TestSortDirsFirstDisabled(t *testing.T) {
	entries := []DirEntry{
		{Name: "file_a.txt", IsDir: false},
		{Name: "dir_b", IsDir: true},
		{Name: "file_c.go", IsDir: false},
		{Name: "dir_a", IsDir: true},
	}
	sortEntries(entries, SortByName, false)

	expected := []string{"dir_a", "dir_b", "file_a.txt", "file_c.go"}
	for i, e := range entries {
		if e.Name != expected[i] {
			t.Errorf("position %d: expected %s, got %s", i, expected[i], e.Name)
		}
	}
}

func TestSortModeString(t *testing.T) {
	tests := []struct {
		mode     SortMode
		expected string
	}{
		{SortByName, "name"},
		{SortByTime, "time"},
		{SortBySize, "size"},
	}
	for _, tt := range tests {
		if tt.mode.String() != tt.expected {
			t.Errorf("expected %s, got %s", tt.expected, tt.mode.String())
		}
	}
}

func TestNextSortMode(t *testing.T) {
	if NextSortMode(SortByName) != SortByTime {
		t.Error("expected name -> time")
	}
	if NextSortMode(SortByTime) != SortBySize {
		t.Error("expected time -> size")
	}
	if NextSortMode(SortBySize) != SortByName {
		t.Error("expected size -> name")
	}
}

func TestSortTimeTiebreaker(t *testing.T) {
	now := time.Now()
	entries := []DirEntry{
		{Name: "beta.txt", IsDir: false, ModTime: now},
		{Name: "alpha.txt", IsDir: false, ModTime: now},
	}
	sortEntries(entries, SortByTime, false)

	// Same time, fall back to name
	if entries[0].Name != "alpha.txt" {
		t.Errorf("expected alpha.txt first (name tiebreaker), got %s", entries[0].Name)
	}
}

func TestSortSizeTiebreaker(t *testing.T) {
	entries := []DirEntry{
		{Name: "beta.txt", IsDir: false, Size: 100},
		{Name: "alpha.txt", IsDir: false, Size: 100},
	}
	sortEntries(entries, SortBySize, false)

	// Same size, fall back to name
	if entries[0].Name != "alpha.txt" {
		t.Errorf("expected alpha.txt first (name tiebreaker), got %s", entries[0].Name)
	}
}
