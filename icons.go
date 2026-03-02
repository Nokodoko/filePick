package main

import (
	"image/color"
	"path/filepath"
	"strings"

	"charm.land/lipgloss/v2"
)

// IconRule defines how to match a filename to an icon.
type IconRule struct {
	Match   string      // "extension", "exact", "exact-dir", "prefix"
	Pattern string      // ".go", "Makefile", ".git"
	Icon    string      // Nerd Font glyph
	Color   color.Color // lipgloss color for the icon
}

var (
	// Directory icons matched by exact name
	dirIcons = []IconRule{
		{Match: "exact-dir", Pattern: ".git", Icon: "\U000f02a2", Color: lipgloss.ANSIColor(208)},
		{Match: "exact-dir", Pattern: ".github", Icon: "\uf408", Color: lipgloss.ANSIColor(7)},
		{Match: "exact-dir", Pattern: "node_modules", Icon: "\ue718", Color: lipgloss.ANSIColor(2)},
		{Match: "exact-dir", Pattern: ".claude", Icon: "\U000f1719", Color: lipgloss.ANSIColor(5)},
	}

	// Default directory icon
	defaultDirIcon = IconRule{Match: "dir", Pattern: "*", Icon: "\U000f024b", Color: lipgloss.ANSIColor(4)}

	// File icons matched by exact filename
	exactFileIcons = []IconRule{
		{Match: "exact", Pattern: "Makefile", Icon: "\ue779", Color: lipgloss.ANSIColor(3)},
		{Match: "exact", Pattern: "Dockerfile", Icon: "\U000f0868", Color: lipgloss.ANSIColor(4)},
		{Match: "exact", Pattern: "docker-compose.yml", Icon: "\U000f0868", Color: lipgloss.ANSIColor(4)},
		{Match: "exact", Pattern: ".gitignore", Icon: "\U000f02a2", Color: lipgloss.ANSIColor(242)},
		{Match: "exact", Pattern: ".gitmodules", Icon: "\U000f02a2", Color: lipgloss.ANSIColor(242)},
		{Match: "exact", Pattern: ".gitattributes", Icon: "\U000f02a2", Color: lipgloss.ANSIColor(242)},
		{Match: "exact", Pattern: "LICENSE", Icon: "\U000f0124", Color: lipgloss.ANSIColor(3)},     // md-certificate (was v2 F718)
		{Match: "exact", Pattern: "LICENSE.md", Icon: "\U000f0124", Color: lipgloss.ANSIColor(3)}, // md-certificate (was v2 F718)
		{Match: "exact", Pattern: "CHANGELOG.md", Icon: "\U000f02da", Color: lipgloss.ANSIColor(4)}, // md-history (was v2 F7D9)
		{Match: "exact", Pattern: "README.md", Icon: "\U000f00ba", Color: lipgloss.ANSIColor(3)},
		{Match: "exact", Pattern: "go.mod", Icon: "\U000f07d3", Color: lipgloss.ANSIColor(6)},
		{Match: "exact", Pattern: "go.sum", Icon: "\U000f07d3", Color: lipgloss.ANSIColor(242)},
		{Match: "exact", Pattern: "Cargo.toml", Icon: "\ue7a8", Color: lipgloss.ANSIColor(1)},
		{Match: "exact", Pattern: "Cargo.lock", Icon: "\ue7a8", Color: lipgloss.ANSIColor(242)},
		{Match: "exact", Pattern: "package.json", Icon: "\ue718", Color: lipgloss.ANSIColor(2)},
		{Match: "exact", Pattern: "package-lock.json", Icon: "\ue718", Color: lipgloss.ANSIColor(242)},
		{Match: "exact", Pattern: "tsconfig.json", Icon: "\U000f06e6", Color: lipgloss.ANSIColor(4)},
		{Match: "exact", Pattern: ".env", Icon: "\uf462", Color: lipgloss.ANSIColor(3)},
		{Match: "exact", Pattern: "CLAUDE.md", Icon: "\U000f1719", Color: lipgloss.ANSIColor(5)},
	}

	// File icons matched by extension
	extIcons = []IconRule{
		{Match: "extension", Pattern: ".go", Icon: "\U000f07d3", Color: lipgloss.ANSIColor(6)},
		{Match: "extension", Pattern: ".mod", Icon: "\U000f07d3", Color: lipgloss.ANSIColor(6)},
		{Match: "extension", Pattern: ".sum", Icon: "\U000f07d3", Color: lipgloss.ANSIColor(242)},
		{Match: "extension", Pattern: ".rs", Icon: "\ue7a8", Color: lipgloss.ANSIColor(1)},
		{Match: "extension", Pattern: ".py", Icon: "\ue73c", Color: lipgloss.ANSIColor(3)},
		{Match: "extension", Pattern: ".js", Icon: "\ue74e", Color: lipgloss.ANSIColor(3)},
		{Match: "extension", Pattern: ".ts", Icon: "\U000f06e6", Color: lipgloss.ANSIColor(4)},
		{Match: "extension", Pattern: ".tsx", Icon: "\U000f06e6", Color: lipgloss.ANSIColor(6)},
		{Match: "extension", Pattern: ".jsx", Icon: "\ue7ba", Color: lipgloss.ANSIColor(6)},
		{Match: "extension", Pattern: ".c", Icon: "\ue61e", Color: lipgloss.ANSIColor(4)},
		{Match: "extension", Pattern: ".h", Icon: "\ue61e", Color: lipgloss.ANSIColor(5)},
		{Match: "extension", Pattern: ".cpp", Icon: "\ue61d", Color: lipgloss.ANSIColor(4)},
		{Match: "extension", Pattern: ".hpp", Icon: "\ue61d", Color: lipgloss.ANSIColor(5)},
		{Match: "extension", Pattern: ".java", Icon: "\ue738", Color: lipgloss.ANSIColor(1)},
		{Match: "extension", Pattern: ".rb", Icon: "\ue739", Color: lipgloss.ANSIColor(1)},
		{Match: "extension", Pattern: ".php", Icon: "\ue73d", Color: lipgloss.ANSIColor(5)},
		{Match: "extension", Pattern: ".swift", Icon: "\U000f06e5", Color: lipgloss.ANSIColor(208)},
		{Match: "extension", Pattern: ".kt", Icon: "\ue634", Color: lipgloss.ANSIColor(5)},
		{Match: "extension", Pattern: ".lua", Icon: "\ue620", Color: lipgloss.ANSIColor(4)},
		{Match: "extension", Pattern: ".sh", Icon: "\uf489", Color: lipgloss.ANSIColor(2)},
		{Match: "extension", Pattern: ".bash", Icon: "\uf489", Color: lipgloss.ANSIColor(2)},
		{Match: "extension", Pattern: ".zsh", Icon: "\uf489", Color: lipgloss.ANSIColor(2)},
		{Match: "extension", Pattern: ".fish", Icon: "\uf489", Color: lipgloss.ANSIColor(2)},
		{Match: "extension", Pattern: ".md", Icon: "\ue73e", Color: lipgloss.ANSIColor(7)},
		{Match: "extension", Pattern: ".txt", Icon: "\U000f0219", Color: lipgloss.ANSIColor(7)},
		{Match: "extension", Pattern: ".json", Icon: "\ue60b", Color: lipgloss.ANSIColor(3)},
		{Match: "extension", Pattern: ".jsonl", Icon: "\ue60b", Color: lipgloss.ANSIColor(3)},
		{Match: "extension", Pattern: ".yaml", Icon: "\ue6a8", Color: lipgloss.ANSIColor(1)},
		{Match: "extension", Pattern: ".yml", Icon: "\ue6a8", Color: lipgloss.ANSIColor(1)},
		{Match: "extension", Pattern: ".toml", Icon: "\ue6b2", Color: lipgloss.ANSIColor(242)},
		{Match: "extension", Pattern: ".xml", Icon: "\U000f15c0", Color: lipgloss.ANSIColor(208)},
		{Match: "extension", Pattern: ".html", Icon: "\ue736", Color: lipgloss.ANSIColor(208)},
		{Match: "extension", Pattern: ".css", Icon: "\ue749", Color: lipgloss.ANSIColor(4)},
		{Match: "extension", Pattern: ".scss", Icon: "\ue749", Color: lipgloss.ANSIColor(5)},
		{Match: "extension", Pattern: ".sql", Icon: "\ue706", Color: lipgloss.ANSIColor(4)},
		{Match: "extension", Pattern: ".graphql", Icon: "\U000f0877", Color: lipgloss.ANSIColor(5)},
		{Match: "extension", Pattern: ".proto", Icon: "\ue6b1", Color: lipgloss.ANSIColor(4)},
		{Match: "extension", Pattern: ".lock", Icon: "\uf023", Color: lipgloss.ANSIColor(242)},
		{Match: "extension", Pattern: ".log", Icon: "\U000f0219", Color: lipgloss.ANSIColor(242)},
		{Match: "extension", Pattern: ".env", Icon: "\uf462", Color: lipgloss.ANSIColor(3)},
		{Match: "extension", Pattern: ".gitignore", Icon: "\U000f02a2", Color: lipgloss.ANSIColor(242)},
		{Match: "extension", Pattern: ".dockerignore", Icon: "\U000f0868", Color: lipgloss.ANSIColor(242)},
		{Match: "extension", Pattern: ".png", Icon: "\U000f021f", Color: lipgloss.ANSIColor(5)},  // md-file_image (was v2 F71E)
		{Match: "extension", Pattern: ".jpg", Icon: "\U000f021f", Color: lipgloss.ANSIColor(5)},  // md-file_image (was v2 F71E)
		{Match: "extension", Pattern: ".jpeg", Icon: "\U000f021f", Color: lipgloss.ANSIColor(5)}, // md-file_image (was v2 F71E)
		{Match: "extension", Pattern: ".gif", Icon: "\U000f021f", Color: lipgloss.ANSIColor(5)},  // md-file_image (was v2 F71E)
		{Match: "extension", Pattern: ".svg", Icon: "\U000f0721", Color: lipgloss.ANSIColor(3)},
		{Match: "extension", Pattern: ".ico", Icon: "\U000f021f", Color: lipgloss.ANSIColor(3)},  // md-file_image (was v2 F71E)
		{Match: "extension", Pattern: ".zip", Icon: "\uf410", Color: lipgloss.ANSIColor(1)},
		{Match: "extension", Pattern: ".tar", Icon: "\uf410", Color: lipgloss.ANSIColor(1)},
		{Match: "extension", Pattern: ".gz", Icon: "\uf410", Color: lipgloss.ANSIColor(1)},
		{Match: "extension", Pattern: ".bz2", Icon: "\uf410", Color: lipgloss.ANSIColor(1)},
		{Match: "extension", Pattern: ".xz", Icon: "\uf410", Color: lipgloss.ANSIColor(1)},
		{Match: "extension", Pattern: ".7z", Icon: "\uf410", Color: lipgloss.ANSIColor(1)},
		{Match: "extension", Pattern: ".pdf", Icon: "\U000f0226", Color: lipgloss.ANSIColor(1)}, // md-file_pdf_box (was v2 F724)
		{Match: "extension", Pattern: ".doc", Icon: "\U000f022c", Color: lipgloss.ANSIColor(4)},
		{Match: "extension", Pattern: ".docx", Icon: "\U000f022c", Color: lipgloss.ANSIColor(4)},
		{Match: "extension", Pattern: ".xls", Icon: "\U000f021b", Color: lipgloss.ANSIColor(2)},
		{Match: "extension", Pattern: ".xlsx", Icon: "\U000f021b", Color: lipgloss.ANSIColor(2)},
		{Match: "extension", Pattern: ".csv", Icon: "\U000f021b", Color: lipgloss.ANSIColor(2)},
		{Match: "extension", Pattern: ".wasm", Icon: "\ue6a1", Color: lipgloss.ANSIColor(5)},
		{Match: "extension", Pattern: ".nix", Icon: "\uf313", Color: lipgloss.ANSIColor(4)},
	}

	// Special type icons
	symlinkIcon    = IconRule{Match: "type", Pattern: "symlink", Icon: "\U000f0337", Color: lipgloss.ANSIColor(6)}
	executableIcon = IconRule{Match: "type", Pattern: "executable", Icon: "\uf489", Color: lipgloss.ANSIColor(2)}
	defaultIcon    = IconRule{Match: "type", Pattern: "default", Icon: "\uf15b", Color: lipgloss.ANSIColor(7)}

	// ASCII fallback icons (when --no-icons is set)
	asciiDirIcon  = ">"
	asciiFileIcon = "-"

	// Lookup maps for fast resolution
	dirIconMap  map[string]IconRule
	exactMap    map[string]IconRule
	extMap      map[string]IconRule
)

func init() {
	dirIconMap = make(map[string]IconRule, len(dirIcons))
	for _, rule := range dirIcons {
		dirIconMap[rule.Pattern] = rule
	}

	exactMap = make(map[string]IconRule, len(exactFileIcons))
	for _, rule := range exactFileIcons {
		exactMap[rule.Pattern] = rule
	}

	extMap = make(map[string]IconRule, len(extIcons))
	for _, rule := range extIcons {
		extMap[rule.Pattern] = rule
	}
}

// resolveIcon returns the icon and icon color for a DirEntry.
// Resolution order:
// 1. Exact directory name match
// 2. Generic directory icon
// 3. Exact filename match
// 4. Extension match
// 5. Executable bit set
// 6. Symlink
// 7. Default file icon
func resolveIcon(entry *DirEntry, noIcons bool) (string, color.Color) {
	if noIcons {
		if entry.IsDir {
			return asciiDirIcon, defaultDirIcon.Color
		}
		return asciiFileIcon, defaultIcon.Color
	}

	// 1. Exact directory name match
	if entry.IsDir {
		if rule, ok := dirIconMap[entry.Name]; ok {
			return rule.Icon, rule.Color
		}
		// 2. Generic directory icon
		return defaultDirIcon.Icon, defaultDirIcon.Color
	}

	// 3. Exact filename match
	if rule, ok := exactMap[entry.Name]; ok {
		return rule.Icon, rule.Color
	}

	// 4. Extension match
	ext := strings.ToLower(filepath.Ext(entry.Name))
	if rule, ok := extMap[ext]; ok {
		return rule.Icon, rule.Color
	}

	// 5. Executable
	if entry.IsExecutable {
		return executableIcon.Icon, executableIcon.Color
	}

	// 6. Symlink (as fallback — the base icon is preferred if found above)
	if entry.IsSymlink {
		return symlinkIcon.Icon, symlinkIcon.Color
	}

	// 7. Default
	return defaultIcon.Icon, defaultIcon.Color
}
