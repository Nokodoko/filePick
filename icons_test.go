package main

import (
	"testing"

	"charm.land/lipgloss/v2"
)

func TestIconResolveDirExact(t *testing.T) {
	entry := &DirEntry{Name: ".git", IsDir: true}
	icon, c := resolveIcon(entry, false)
	if icon != "\U000f02a2" {
		t.Errorf("expected .git icon \\U000f02a2, got %q", icon)
	}
	if c != lipgloss.ANSIColor(208) {
		t.Errorf("expected orange color for .git, got %v", c)
	}
}

func TestIconResolveDirGithub(t *testing.T) {
	entry := &DirEntry{Name: ".github", IsDir: true}
	icon, _ := resolveIcon(entry, false)
	if icon != "\uf408" {
		t.Errorf("expected .github icon \\uf408, got %q", icon)
	}
}

func TestIconResolveDirNodeModules(t *testing.T) {
	entry := &DirEntry{Name: "node_modules", IsDir: true}
	icon, _ := resolveIcon(entry, false)
	if icon != "\ue718" {
		t.Errorf("expected node_modules icon \\ue718, got %q", icon)
	}
}

func TestIconResolveDirClaude(t *testing.T) {
	entry := &DirEntry{Name: ".claude", IsDir: true}
	icon, _ := resolveIcon(entry, false)
	if icon != "\U000f1719" {
		t.Errorf("expected .claude icon \\U000f1719, got %q", icon)
	}
}

func TestIconResolveDirGeneric(t *testing.T) {
	entry := &DirEntry{Name: "src", IsDir: true}
	icon, c := resolveIcon(entry, false)
	if icon != "\U000f024b" {
		t.Errorf("expected generic dir icon \\U000f024b, got %q", icon)
	}
	if c != lipgloss.ANSIColor(4) {
		t.Errorf("expected blue color for generic dir, got %v", c)
	}
}

func TestIconResolveExactFile(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"Makefile", "\ue779"},
		{"Dockerfile", "\U000f0868"},
		{"go.mod", "\U000f07d3"},
		{"go.sum", "\U000f07d3"},
		{"README.md", "\U000f00ba"},
		{"LICENSE", "\uf718"},
		{".gitignore", "\U000f02a2"},
		{".env", "\uf462"},
		{"CLAUDE.md", "\U000f1719"},
		{"package.json", "\ue718"},
		{"Cargo.toml", "\ue7a8"},
		{"tsconfig.json", "\U000f06e6"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &DirEntry{Name: tt.name, IsDir: false}
			icon, _ := resolveIcon(entry, false)
			if icon != tt.expected {
				t.Errorf("expected icon for %s to be %q, got %q", tt.name, tt.expected, icon)
			}
		})
	}
}

func TestIconResolveExtension(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"main.go", "\U000f07d3"},
		{"lib.rs", "\ue7a8"},
		{"app.py", "\ue73c"},
		{"index.js", "\ue74e"},
		{"app.ts", "\U000f06e6"},
		{"comp.tsx", "\U000f06e6"},
		{"style.css", "\ue749"},
		{"page.html", "\ue736"},
		{"data.json", "\ue60b"},
		{"config.yaml", "\ue6a8"},
		{"config.yml", "\ue6a8"},
		{"doc.md", "\ue73e"},
		{"notes.txt", "\U000f0219"},
		{"script.sh", "\uf489"},
		{"photo.png", "\uf71e"},
		{"archive.zip", "\uf410"},
		{"report.pdf", "\uf724"},
		{"query.sql", "\ue706"},
		{"data.csv", "\U000f021b"},
		{"flake.nix", "\uf313"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &DirEntry{Name: tt.name, IsDir: false}
			icon, _ := resolveIcon(entry, false)
			if icon != tt.expected {
				t.Errorf("expected icon for %s to be %q, got %q", tt.name, tt.expected, icon)
			}
		})
	}
}

func TestIconResolveExecutable(t *testing.T) {
	entry := &DirEntry{Name: "mybin", IsDir: false, IsExecutable: true}
	icon, _ := resolveIcon(entry, false)
	if icon != "\uf489" {
		t.Errorf("expected executable icon \\uf489, got %q", icon)
	}
}

func TestIconResolveSymlink(t *testing.T) {
	entry := &DirEntry{Name: "link", IsDir: false, IsSymlink: true}
	icon, _ := resolveIcon(entry, false)
	if icon != "\U000f0337" {
		t.Errorf("expected symlink icon \\U000f0337, got %q", icon)
	}
}

func TestIconResolveDefault(t *testing.T) {
	entry := &DirEntry{Name: "unknown_file", IsDir: false}
	icon, _ := resolveIcon(entry, false)
	if icon != "\uf15b" {
		t.Errorf("expected default icon \\uf15b, got %q", icon)
	}
}

func TestIconResolveNoIcons(t *testing.T) {
	dirEntry := &DirEntry{Name: "src", IsDir: true}
	icon, _ := resolveIcon(dirEntry, true)
	if icon != ">" {
		t.Errorf("expected ASCII dir icon '>', got %q", icon)
	}

	fileEntry := &DirEntry{Name: "main.go", IsDir: false}
	icon, _ = resolveIcon(fileEntry, true)
	if icon != "-" {
		t.Errorf("expected ASCII file icon '-', got %q", icon)
	}
}

func TestIconResolutionOrder(t *testing.T) {
	// Exact name should take precedence over extension
	entry := &DirEntry{Name: "go.mod", IsDir: false, IsExecutable: true}
	icon, _ := resolveIcon(entry, false)
	// go.mod should match exact name, not .mod extension or executable
	if icon != "\U000f07d3" {
		t.Errorf("exact name should take precedence, expected \\U000f07d3, got %q", icon)
	}
}

func TestIconMappingCount(t *testing.T) {
	total := len(dirIcons) + len(exactFileIcons) + len(extIcons) + 3 // +3 for symlink, exec, default
	if total < 50 {
		t.Errorf("expected at least 50 icon mappings, got %d", total)
	}
}
