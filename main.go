package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// Version is the application version.
const Version = "0.1.0"

func main() {
	// Flags
	showHidden := flag.Bool("hidden", false, "Show hidden files on startup")
	sortFlag := flag.String("sort", "name", "Sort mode: name|time|size")
	noIcons := flag.Bool("no-icons", false, "Disable Nerd Font icons (use ASCII fallback)")
	version := flag.Bool("version", false, "Print version and exit")

	// Custom usage
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: fp [flags] [path]\n\n")
		fmt.Fprintf(os.Stderr, "A single-pane file picker TUI with real-time filesystem watching.\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		fmt.Fprintf(os.Stderr, "  --hidden          Show hidden files on startup (default: off)\n")
		fmt.Fprintf(os.Stderr, "  --sort <mode>     Sort mode: name|time|size (default: name)\n")
		fmt.Fprintf(os.Stderr, "  --no-icons        Disable Nerd Font icons (use ASCII fallback)\n")
		fmt.Fprintf(os.Stderr, "  --version         Print version and exit\n")
		fmt.Fprintf(os.Stderr, "  --help            Print this help message\n")
		fmt.Fprintf(os.Stderr, "\nKeybindings:\n")
		fmt.Fprintf(os.Stderr, "  j/Down       Move cursor down\n")
		fmt.Fprintf(os.Stderr, "  k/Up         Move cursor up\n")
		fmt.Fprintf(os.Stderr, "  g/Home       Jump to first entry\n")
		fmt.Fprintf(os.Stderr, "  G/End        Jump to last entry\n")
		fmt.Fprintf(os.Stderr, "  Enter/l      Enter directory\n")
		fmt.Fprintf(os.Stderr, "  h/Backspace  Go to parent directory\n")
		fmt.Fprintf(os.Stderr, "  .            Toggle hidden files\n")
		fmt.Fprintf(os.Stderr, "  s            Cycle sort mode\n")
		fmt.Fprintf(os.Stderr, "  /            Fuzzy search (requires rg)\n")
		fmt.Fprintf(os.Stderr, "  q/Esc/Ctrl-C Quit\n")
		fmt.Fprintf(os.Stderr, "\nEnvironment:\n")
		fmt.Fprintf(os.Stderr, "  NO_COLOR=1   Disable all ANSI color output\n")
		fmt.Fprintf(os.Stderr, "  FP_ICONS=0   Disable Nerd Font icons\n")
		fmt.Fprintf(os.Stderr, "\nRequires a Nerd Font patched terminal for icons.\n")
	}

	flag.Parse()

	if *version {
		fmt.Printf("fp %s\n", Version)
		os.Exit(0)
	}

	// Check FP_ICONS environment variable
	if os.Getenv("FP_ICONS") == "0" {
		*noIcons = true
	}

	// Parse sort mode
	var sortMode SortMode
	switch strings.ToLower(*sortFlag) {
	case "name":
		sortMode = SortByName
	case "time":
		sortMode = SortByTime
	case "size":
		sortMode = SortBySize
	default:
		fmt.Fprintf(os.Stderr, "fp: invalid sort mode %q (use name, time, or size)\n", *sortFlag)
		os.Exit(1)
	}

	// Determine directory to open
	dir := "."
	if flag.NArg() > 0 {
		dir = flag.Arg(0)
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fp: %s\n", err)
		os.Exit(1)
	}

	// Verify directory exists
	info, err := os.Stat(absDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fp: %s\n", err)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "fp: %s is not a directory\n", absDir)
		os.Exit(1)
	}

	// Create model
	model := NewAppModel(absDir, *showHidden, sortMode, *noIcons)

	// Create program
	p := tea.NewProgram(&model)

	// Set program reference and start watcher
	model.SetProgram(p)
	if err := model.StartWatcher(); err != nil {
		fmt.Fprintf(os.Stderr, "fp: warning: %s\n", err)
	}

	// Run the program
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "fp: %s\n", err)
		os.Exit(1)
	}

	// Clean up
	model.Close()

	// Print selected file path to stdout if a file was selected
	if fm, ok := finalModel.(*AppModel); ok && fm.selected != "" {
		fmt.Println(fm.selected)
	}
}
