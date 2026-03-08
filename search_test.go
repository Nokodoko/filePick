package main

import (
	"strings"
	"testing"
)

func TestFuzzyMatchExact(t *testing.T) {
	if !fuzzyMatch("main.go", "main.go") {
		t.Error("exact match should succeed")
	}
}

func TestFuzzyMatchSubstring(t *testing.T) {
	if !fuzzyMatch("main.go", "main") {
		t.Error("substring match should succeed")
	}
}

func TestFuzzyMatchScattered(t *testing.T) {
	if !fuzzyMatch("model.go", "mdg") {
		t.Error("scattered chars 'mdg' should match 'model.go'")
	}
}

func TestFuzzyMatchNoMatch(t *testing.T) {
	if fuzzyMatch("main.go", "xyz") {
		t.Error("'xyz' should not match 'main.go'")
	}
}

func TestFuzzyMatchEmpty(t *testing.T) {
	if !fuzzyMatch("anything", "") {
		t.Error("empty pattern should match anything")
	}
}

func TestFuzzyMatchCaseInsensitive(t *testing.T) {
	// fuzzyMatch itself is case-sensitive; callers lowercase both
	if !fuzzyMatch("readme.md", "readme") {
		t.Error("lowercase match should succeed")
	}
}

func TestFuzzyMatchPath(t *testing.T) {
	if !fuzzyMatch("src/model/user.go", "modu") {
		t.Error("path fuzzy match should succeed")
	}
}

func TestRenderSearchBar(t *testing.T) {
	styles := NewStyles()
	result := renderSearchBar("test", 80, styles)
	if !strings.Contains(result, "/") {
		t.Error("search bar should contain '/' prompt")
	}
	if !strings.Contains(result, "test") {
		t.Error("search bar should contain query text")
	}
}

func TestRenderViewSearchMode(t *testing.T) {
	styles := NewStyles()
	entries := []DirEntry{
		{Name: "model.go", IsDir: false},
	}
	result := renderView("/home/user/project", entries, 0, 0, 80, 24, SortByName, false, "", styles, true, "model", nil)
	if !strings.Contains(result, "/") {
		t.Error("search mode view should contain search prompt")
	}
	if !strings.Contains(result, "model") {
		t.Error("search mode view should contain search query")
	}
}
