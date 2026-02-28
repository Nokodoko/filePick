# filePick

Single-pane file picker TUI for watching git worktrees spawn and disappear in real time. Go runtime, bubbletea framework, fsnotify filesystem events, Nerd Font icons.

Built for one purpose: stare at a directory and see git worktrees appear as agents create them, then vanish as they merge and clean up. A live dashboard for swarm-driven development.

## Why

Existing file managers (Yazi, lf, ranger) are general-purpose tools with features filePick does not need:

- **Multi-pane complexity.** Yazi shows parent directory, current directory, and file preview simultaneously. For worktree watching, only the current directory matters.
- **Configuration overhead.** Yazi requires `~/.config/yazi/` with multiple TOML files to customize. filePick has zero configuration — it reads `$CWD` and watches.
- **No live filesystem events.** Yazi polls or re-reads on focus. filePick uses inotify/fsnotify to reflect filesystem changes within milliseconds.
- **No git worktree awareness.** No existing file picker highlights `.claude/worktrees/` entries or understands the worktree lifecycle.
- **Heavy dependencies.** Yazi is a Rust binary with async runtime, image protocol support, and plugin system. filePick is a single Go binary with three dependencies.

filePick does exactly one thing: display a directory listing with icons that updates in real time via filesystem events.

## Design Principles

1. **Single-pane, zero-chrome.** One column of icons + filenames. No parent pane, no preview pane, no tabs, no splits. The highlighted selection bar and top path are the only UI beyond the file list.
2. **Real-time by default.** Every filesystem change (create, delete, rename, chmod) triggers an immediate re-render. No polling, no manual refresh. fsnotify delivers inotify events on Linux.
3. **Git worktree first-class.** When `$CWD` contains `.claude/worktrees/` or `.git/worktrees/`, worktree directories are visually distinguished. New worktrees appear instantly; deleted worktrees vanish instantly.
4. **Nerd Font native.** Uses the same glyph codepoints as Yazi for visual consistency. Users with a Nerd Font patched terminal see identical icons.
5. **Dark terminal assumption.** All colors are chosen for dark backgrounds (ANSI 256-color). No light-theme mode. Respects `NO_COLOR` for piped/redirected output.
6. **Zero configuration.** No config file, no dotfiles, no RC. Behavior is controlled entirely by CLI flags and environment variables (`NO_COLOR`, `TERM`).
7. **Stateless.** No history, no bookmarks, no session restore. Launch, look, quit. State lives in the filesystem, not in filePick.

## On-Disk Format

filePick has no persistent on-disk format. It is a read-only viewer of the filesystem. There are no config files, databases, or caches to manage.

The only "format" is the filesystem itself:

```
$CWD/
  .claude/
    worktrees/            # Watched for worktree spawn/delete events
      feature-auth/       # Active worktree (appears in real time)
      fix-bug-123/        # Active worktree (appears in real time)
  .git/
    worktrees/            # Standard git worktree metadata
  agents/                 # Regular directory
  cmd/                    # Regular directory
  go.mod                  # Regular file
  go.sum                  # Regular file
  Makefile                # Regular file
  README.md               # Regular file
```

### Filesystem Reading

filePick reads directory entries via `os.ReadDir()` on startup and after each fsnotify event. It stats each entry to determine:

- Type: directory, regular file, symlink, executable
- Name: used for icon mapping (extension, exact name match)
- ModTime: used for optional sort-by-time mode

No file contents are ever read. filePick never opens files — only directories.

## Data Model

### DirEntry

```typescript
interface DirEntry {
  // Identity
  name: string;           // "go.mod", "agents", "README.md"
  path: string;           // Full absolute path

  // Classification
  isDir: boolean;         // true for directories
  isSymlink: boolean;     // true for symlinks (resolved for icon, shown with indicator)
  isExecutable: boolean;  // true if any execute bit set (unix only)
  isHidden: boolean;      // true if name starts with "."

  // Display
  icon: string;           // Nerd Font glyph, e.g. "\uf115" for directory
  iconColor: ANSIColor;   // ANSI color code for the icon
  nameColor: ANSIColor;   // ANSI color code for the filename text

  // Metadata
  size: number;           // File size in bytes (for display, not sorting by default)
  modTime: string;        // ISO 8601, for sort-by-time mode
}
```

### AppModel (bubbletea Model)

```typescript
interface AppModel {
  // Directory state
  cwd: string;                    // Current working directory (absolute path)
  entries: DirEntry[];            // Sorted directory entries
  filteredEntries: DirEntry[];    // After hidden-file filter applied

  // UI state
  cursor: number;                 // Index of highlighted entry in filteredEntries
  viewportOffset: number;         // First visible row (for scrolling)
  termWidth: number;              // Terminal columns
  termHeight: number;             // Terminal rows

  // Settings
  showHidden: boolean;            // Toggle with '.' key (default: false)
  sortMode: SortMode;            // "name" | "time" | "size" (default: "name")
  dirsFirst: boolean;             // Directories sorted before files (default: true)

  // Filesystem watcher
  watcher: FSWatcher;             // fsnotify watcher instance

  // Error state
  err?: string;                   // Displayed in status bar if non-nil
}
```

### SortMode

```typescript
type SortMode = "name" | "time" | "size";
```

### ANSIColor

```typescript
type ANSIColor = string;  // lipgloss.Color value, e.g. "4" (blue), "7" (white), "2" (green)
```

### Icon Mapping

```typescript
interface IconRule {
  match: "extension" | "exact" | "prefix";  // How to match the filename
  pattern: string;                           // ".go", "Makefile", ".git"
  icon: string;                              // Nerd Font glyph
  color: ANSIColor;                          // Icon color
}
```

### Navigation Lifecycle

```
idle ──> keypress ──> update_model ──> render
  ^                                      |
  +--------------------------------------+

idle ──> fsnotify_event ──> reload_dir ──> re_sort ──> render
  ^                                                      |
  +------------------------------------------------------+
```

The bubbletea event loop handles both keyboard input and filesystem events through the same `Update()` function. fsnotify events are delivered as custom `tea.Msg` types via a goroutine that reads from the fsnotify channel and sends `tea.Msg` values into the bubbletea program.

## CLI

Binary name: `fp` (short for filePick).

No subcommands. `fp` launches the TUI. All configuration is via flags.

### Launch

```
fp [path]                              Open TUI in [path] or $CWD
  --hidden                             Show hidden files on startup (default: off)
  --sort <mode>                        name|time|size (default: name)
  --no-icons                           Disable Nerd Font icons (use ASCII fallback)
  --version                            Print version and exit
  --help                               Print usage and exit
```

### Keybindings (In-TUI)

```
j / Down        Move cursor down
k / Up          Move cursor up
g / Home        Jump to first entry
G / End         Jump to last entry
Enter / l       Enter directory (cd into selected dir)
h / Backspace   Go to parent directory
.               Toggle hidden files
s               Cycle sort mode (name -> time -> size -> name)
q / Esc / Ctrl-C  Quit
```

### Environment Variables

```
NO_COLOR=1       Disable all ANSI color output
TERM             Used by lipgloss for capability detection
FP_ICONS=0       Disable Nerd Font icons (same as --no-icons)
```

## JSON Output Format

filePick is a TUI application, not a CLI tool with structured output. There is no JSON mode.

However, filePick prints the selected entry's absolute path to stdout on quit-with-selection (pressing Enter on a file, not a directory):

```
# User selects a file and presses Enter (or a custom "select" binding)
/home/n0ko/Programs/filePick/go.mod
```

This enables shell integration:

```bash
# Open selected file in editor
vim $(fp)

# Or with command substitution
fp | xargs vim
```

Error output goes to stderr:

```
fp: cannot read directory '/root/private': permission denied
```

## Concurrency Model

filePick has two concurrent concerns: the TUI event loop and the filesystem watcher.

### Goroutine Architecture

```
+-------------------+      tea.Msg       +-------------------+
|  fsnotify watcher |  --------------->  |  bubbletea loop   |
|  (goroutine)      |                    |  (main goroutine) |
+-------------------+                    +-------------------+
        |                                         |
        | inotify fd                              | terminal I/O
        v                                         v
  [kernel inotify]                          [stdin/stdout]
```

### Implementation

1. On startup, create `fsnotify.Watcher` watching `cwd`
2. Spawn a goroutine that reads from `watcher.Events` channel
3. On each event, send a custom `fsEventMsg` into the bubbletea program via `p.Send()`
4. The `Update()` function handles `fsEventMsg` by calling `os.ReadDir()` and re-sorting
5. Debounce rapid events: accumulate events for 50ms before triggering a single re-read
6. On directory change (Enter/Backspace), remove old watch and add new watch

### Watcher Scope

- filePick watches **only the current directory** (non-recursive)
- When the user navigates into a subdirectory, the old watch is removed and a new watch is added
- No recursive watching — this keeps inotify descriptor usage at exactly 1

### Race Conditions

- `os.ReadDir()` may return entries that are deleted between read and render — handle gracefully (skip missing entries on stat failure)
- Multiple fsnotify events may fire in rapid succession (e.g., `git worktree add` creates multiple filesystem operations) — the 50ms debounce coalesces these into a single re-read
- Terminal resize events (`tea.WindowSizeMsg`) are handled by bubbletea natively

## Migration

Not applicable. filePick is a greenfield project with no predecessor to migrate from. There is no existing data, configuration, or state to port.

## Integration

### Shell Integration

filePick prints the selected file path to stdout, enabling composition with shell tools:

```bash
# Pick a file and open it
vim "$(fp /home/n0ko/Projects)"

# Pick a file and copy its path
fp | xclip -selection clipboard

# Use in a script
selected=$(fp /var/log)
if [ -n "$selected" ]; then
  less "$selected"
fi
```

### Agent Worktree Monitoring

The primary integration is visual monitoring of agent worktree activity:

```bash
# Watch worktrees directory for agent activity
fp /home/n0ko/Programs/myproject/.claude/worktrees

# Or from project root, navigate into .claude/worktrees/
fp /home/n0ko/Programs/myproject
```

Agents spawned by the supervisor create worktrees via `git worktree add`. filePick's fsnotify watcher detects the new directory within milliseconds and renders it with a folder icon. When the agent completes and the worktree is removed (`git worktree remove`), filePick detects the deletion and removes the entry from the list.

### No Programmatic API

filePick exposes no library, no IPC socket, no HTTP endpoint. It is a terminal application only. Integration is via stdout path output and visual monitoring.

## What It Does NOT Do

Explicitly out of scope (keep it minimal):

- **No file preview.** filePick never reads file contents. No syntax highlighting, no image preview, no hex dump. Use `cat`, `bat`, or your editor.
- **No file operations.** No copy, move, delete, rename, chmod. filePick is read-only. Use your shell or a file manager.
- **No multi-pane layout.** No parent directory pane, no preview pane, no splits, no tabs. One pane, one directory.
- **No search or filter.** No fuzzy finder, no regex filter, no type-to-search. Use `fzf` if you need search.
- **No bookmarks or history.** No saved locations, no recent directories, no session state. Launch fresh every time.
- **No configuration file.** No TOML, no YAML, no dotfile. Flags and environment variables only.
- **No remote filesystems.** No SSH, no SFTP, no S3, no FUSE awareness. Local filesystem only.
- **No plugin system.** No Lua, no shell hooks, no custom commands. The binary is the entire application.
- **No mouse support.** Keyboard only. No click-to-select, no scroll wheel (bubbletea supports mouse, but filePick opts out for simplicity).
- **No light theme.** Dark terminal backgrounds only. If `NO_COLOR` is set, all color is stripped.

## Tech Stack

| Concern | Choice | Rationale |
|---------|--------|-----------|
| Language | Go 1.22+ | Fast compilation, single static binary, excellent stdlib for filesystem ops |
| TUI Framework | charmbracelet/bubbletea v2 | The standard Go TUI framework, Elm architecture, testable |
| Styling | charmbracelet/lipgloss v2 | Companion to bubbletea, ANSI 256-color, respects NO_COLOR |
| FS Watching | fsnotify/fsnotify v1 | Cross-platform inotify wrapper, de facto Go standard |
| Testing | stdlib `testing` | No test framework dependency, `go test ./...` |
| Formatting | `gofmt` / `goimports` | Go standard, zero config |
| Linting | golangci-lint | Standard Go meta-linter |
| Distribution | `go install` / `go build` | Single static binary, no runtime dependencies |
| Icons | Nerd Font glyphs (hardcoded) | Same codepoints as Yazi, no icon font dependency in binary |

## Nerd Font Icon Mapping

The following table maps file types to Nerd Font glyphs. These are the same codepoints Yazi uses for visual consistency. The glyph column shows the Unicode codepoint; the terminal renders the actual icon when a Nerd Font is installed.

### Directory Icons

| Match Type | Pattern | Glyph | Codepoint | Color | Description |
|------------|---------|-------|-----------|-------|-------------|
| dir (default) | `*` | 󰉋 | `\U000f024b` | Blue (4) | Generic closed folder |
| exact-dir | `.git` | 󰊢 | `\U000f02a2` | Orange (208) | Git directory |
| exact-dir | `.github` |  | `\uf408` | White (7) | GitHub directory |
| exact-dir | `node_modules` |  | `\ue718` | Green (2) | Node modules |
| exact-dir | `.claude` | 󱜙 | `\U000f1719` | Magenta (5) | Claude config |

### File Icons by Exact Name

| Match Type | Pattern | Glyph | Codepoint | Color | Description |
|------------|---------|-------|-----------|-------|-------------|
| exact | `Makefile` |  | `\ue779` | Yellow (3) | Make |
| exact | `Dockerfile` | 󰡨 | `\U000f0868` | Blue (4) | Docker |
| exact | `docker-compose.yml` | 󰡨 | `\U000f0868` | Blue (4) | Docker Compose |
| exact | `.gitignore` | 󰊢 | `\U000f02a2` | Grey (242) | Git ignore |
| exact | `.gitmodules` | 󰊢 | `\U000f02a2` | Grey (242) | Git submodules |
| exact | `.gitattributes` | 󰊢 | `\U000f02a2` | Grey (242) | Git attributes |
| exact | `LICENSE` |  | `\uf718` | Yellow (3) | License |
| exact | `LICENSE.md` |  | `\uf718` | Yellow (3) | License |
| exact | `CHANGELOG.md` |  | `\uf7d9` | Blue (4) | Changelog |
| exact | `README.md` | 󰂺 | `\U000f00ba` | Yellow (3) | Readme |
| exact | `go.mod` | 󰟓 | `\U000f07d3` | Cyan (6) | Go module |
| exact | `go.sum` | 󰟓 | `\U000f07d3` | Grey (242) | Go checksum |
| exact | `Cargo.toml` |  | `\ue7a8` | Red (1) | Rust Cargo |
| exact | `Cargo.lock` |  | `\ue7a8` | Grey (242) | Rust lock |
| exact | `package.json` |  | `\ue718` | Green (2) | Node package |
| exact | `package-lock.json` |  | `\ue718` | Grey (242) | Node lockfile |
| exact | `tsconfig.json` | 󰛦 | `\U000f06e6` | Blue (4) | TypeScript config |
| exact | `.env` |  | `\uf462` | Yellow (3) | Environment |
| exact | `CLAUDE.md` | 󱜙 | `\U000f1719` | Magenta (5) | Claude instructions |

### File Icons by Extension

| Match Type | Pattern | Glyph | Codepoint | Color | Description |
|------------|---------|-------|-----------|-------|-------------|
| ext | `.go` | 󰟓 | `\U000f07d3` | Cyan (6) | Go |
| ext | `.mod` | 󰟓 | `\U000f07d3` | Cyan (6) | Go module |
| ext | `.sum` | 󰟓 | `\U000f07d3` | Grey (242) | Go checksum |
| ext | `.rs` |  | `\ue7a8` | Red (1) | Rust |
| ext | `.py` |  | `\ue73c` | Yellow (3) | Python |
| ext | `.js` |  | `\ue74e` | Yellow (3) | JavaScript |
| ext | `.ts` | 󰛦 | `\U000f06e6` | Blue (4) | TypeScript |
| ext | `.tsx` | 󰛦 | `\U000f06e6` | Cyan (6) | TypeScript React |
| ext | `.jsx` |  | `\ue7ba` | Cyan (6) | JavaScript React |
| ext | `.c` |  | `\ue61e` | Blue (4) | C |
| ext | `.h` |  | `\ue61e` | Magenta (5) | C header |
| ext | `.cpp` |  | `\ue61d` | Blue (4) | C++ |
| ext | `.hpp` |  | `\ue61d` | Magenta (5) | C++ header |
| ext | `.java` |  | `\ue738` | Red (1) | Java |
| ext | `.rb` |  | `\ue739` | Red (1) | Ruby |
| ext | `.php` |  | `\ue73d` | Magenta (5) | PHP |
| ext | `.swift` | 󰛥 | `\U000f06e5` | Orange (208) | Swift |
| ext | `.kt` |  | `\ue634` | Magenta (5) | Kotlin |
| ext | `.lua` |  | `\ue620` | Blue (4) | Lua |
| ext | `.sh` |  | `\uf489` | Green (2) | Shell script |
| ext | `.bash` |  | `\uf489` | Green (2) | Bash script |
| ext | `.zsh` |  | `\uf489` | Green (2) | Zsh script |
| ext | `.fish` |  | `\uf489` | Green (2) | Fish script |
| ext | `.md` |  | `\ue73e` | White (7) | Markdown |
| ext | `.txt` | 󰈙 | `\U000f0219` | White (7) | Plain text |
| ext | `.json` |  | `\ue60b` | Yellow (3) | JSON |
| ext | `.jsonl` |  | `\ue60b` | Yellow (3) | JSON Lines |
| ext | `.yaml` |  | `\ue6a8` | Red (1) | YAML |
| ext | `.yml` |  | `\ue6a8` | Red (1) | YAML |
| ext | `.toml` |  | `\ue6b2` | Grey (242) | TOML |
| ext | `.xml` | 󰗀 | `\U000f15c0` | Orange (208) | XML |
| ext | `.html` |  | `\ue736` | Orange (208) | HTML |
| ext | `.css` |  | `\ue749` | Blue (4) | CSS |
| ext | `.scss` |  | `\ue749` | Magenta (5) | SCSS |
| ext | `.sql` |  | `\ue706` | Blue (4) | SQL |
| ext | `.graphql` | 󰡷 | `\U000f0877` | Magenta (5) | GraphQL |
| ext | `.proto` |  | `\ue6b1` | Blue (4) | Protobuf |
| ext | `.lock` |  | `\uf023` | Grey (242) | Lock file |
| ext | `.log` | 󰈙 | `\U000f0219` | Grey (242) | Log file |
| ext | `.env` |  | `\uf462` | Yellow (3) | Environment |
| ext | `.gitignore` | 󰊢 | `\U000f02a2` | Grey (242) | Git |
| ext | `.dockerignore` | 󰡨 | `\U000f0868` | Grey (242) | Docker ignore |
| ext | `.png` |  | `\uf71e` | Magenta (5) | PNG image |
| ext | `.jpg` |  | `\uf71e` | Magenta (5) | JPEG image |
| ext | `.jpeg` |  | `\uf71e` | Magenta (5) | JPEG image |
| ext | `.gif` |  | `\uf71e` | Magenta (5) | GIF image |
| ext | `.svg` | 󰜡 | `\U000f0721` | Yellow (3) | SVG image |
| ext | `.ico` |  | `\uf71e` | Yellow (3) | Icon |
| ext | `.zip` |  | `\uf410` | Red (1) | Zip archive |
| ext | `.tar` |  | `\uf410` | Red (1) | Tar archive |
| ext | `.gz` |  | `\uf410` | Red (1) | Gzip |
| ext | `.bz2` |  | `\uf410` | Red (1) | Bzip2 |
| ext | `.xz` |  | `\uf410` | Red (1) | XZ |
| ext | `.7z` |  | `\uf410` | Red (1) | 7-Zip |
| ext | `.pdf` |  | `\uf724` | Red (1) | PDF |
| ext | `.doc` | 󰈬 | `\U000f022c` | Blue (4) | Word |
| ext | `.docx` | 󰈬 | `\U000f022c` | Blue (4) | Word |
| ext | `.xls` | 󰈛 | `\U000f021b` | Green (2) | Excel |
| ext | `.xlsx` | 󰈛 | `\U000f021b` | Green (2) | Excel |
| ext | `.csv` | 󰈛 | `\U000f021b` | Green (2) | CSV |
| ext | `.wasm` |  | `\ue6a1` | Magenta (5) | WebAssembly |
| ext | `.nix` |  | `\uf313` | Blue (4) | Nix |

### Special Type Icons

| Match Type | Pattern | Glyph | Codepoint | Color | Description |
|------------|---------|-------|-----------|-------|-------------|
| type | symlink | 󰌷 | `\U000f0337` | Cyan (6) | Symbolic link |
| type | executable |  | `\uf489` | Green (2) | Executable file |
| type | default |  | `\uf15b` | White (7) | Unknown file type |

### Icon Resolution Order

1. Exact directory name match (e.g., `.git`, `node_modules`)
2. If directory: generic folder icon
3. Exact filename match (e.g., `Makefile`, `Dockerfile`, `go.mod`)
4. Extension match (e.g., `.go`, `.md`, `.rs`)
5. Executable bit set: executable icon
6. Symlink: symlink icon (overlay — the base icon is resolved first, symlink indicator appended)
7. Default: generic file icon

## UI Layout

### ASCII Mockup — Normal State

```
┌─────────────────────────────────────────────────────────┐
│  ~/Programs/ai/computeCommander                         │
├─────────────────────────────────────────────────────────┤
│  󰉋  .claude                                            │
│  󰉋  agents                                             │
│  󰉋  cmd                                                │
│  󰉋  internal                                           │
│  󰉋  migrations                                         │
│  󰉋  pkg                                                │
│  󰉋  templates                                          │
│ ████████████████████████████████████████████████████████│
│ █ 󰟓  go.mod                                           █│
│ ████████████████████████████████████████████████████████│
│  󰟓  go.sum                                             │
│    Makefile                                             │
│    README.md                                            │
│  󰊢  .gitignore                                         │
├─────────────────────────────────────────────────────────┤
│  14 items                                      name  │
└─────────────────────────────────────────────────────────┘
```

### ASCII Mockup — Worktree Watching

```
┌─────────────────────────────────────────────────────────┐
│  ~/Projects/myapp/.claude/worktrees                     │
├─────────────────────────────────────────────────────────┤
│ ████████████████████████████████████████████████████████│
│ █ 󰉋  feature-auth                                     █│
│ ████████████████████████████████████████████████████████│
│  󰉋  fix-validation-bug                                 │
│  󰉋  refactor-db-layer                                  │
│                                                         │
│  (watching for changes...)                              │
│                                                         │
├─────────────────────────────────────────────────────────┤
│  3 items                                       name  │
└─────────────────────────────────────────────────────────┘
```

### Layout Specification

```
+-----------------------------------------------------------+
|  PATH BAR          (1 line, top)                          |
|  Truncated from left: "~/Programs/ai/computeCo..."        |
|  Style: dim white on default background                   |
+-----------------------------------------------------------+
|  SEPARATOR          (1 line, thin horizontal rule)        |
+-----------------------------------------------------------+
|                                                           |
|  FILE LIST          (terminal height - 4 lines)           |
|  Each row: "  {icon}  {name}"                             |
|  Selected row: full-width highlight bar (bg #3b3b3b)      |
|  Directory names: blue foreground                         |
|  Symlink names: cyan foreground with " -> target" suffix  |
|  Executable names: green foreground                       |
|  Regular file names: white foreground                     |
|  Hidden entries: dimmed                                   |
|                                                           |
+-----------------------------------------------------------+
|  SEPARATOR          (1 line, thin horizontal rule)        |
+-----------------------------------------------------------+
|  STATUS BAR         (1 line, bottom)                      |
|  Left: "{count} items"                                    |
|  Right: sort mode indicator                               |
|  Style: dim                                               |
+-----------------------------------------------------------+
```

### Color Palette

| Element | Foreground | Background | Style |
|---------|-----------|------------|-------|
| Path bar | White (7) | Default | Dim, bold |
| Separator | Grey (240) | Default | Dim |
| Directory name | Blue (4) | Default | Bold |
| Regular file name | White (7) | Default | Normal |
| Executable name | Green (2) | Default | Bold |
| Symlink name | Cyan (6) | Default | Italic |
| Hidden entry | Grey (242) | Default | Dim |
| Selection bar (text) | White (15) | Grey (#3b3b3b) | Bold |
| Selection bar (bg) | — | `#3b3b3b` | Full width |
| Status bar | Grey (242) | Default | Dim |
| Error message | Red (1) | Default | Bold |

### Path Truncation

The path bar shows the current directory path, truncated from the left if it exceeds terminal width minus 4 characters:

- Full: `~/Programs/ai/computeCommander`
- Truncated: `...ai/computeCommander` (at component boundaries, not mid-word)
- Home directory replaced with `~`

## Project Infrastructure

### Directory Structure

```
filePick/
  go.mod                              # Module: github.com/n0ko/filePick
  go.sum                              # Dependency checksums
  main.go                             # Entry point: flag parsing, model init, tea.NewProgram
  model.go                            # bubbletea Model: Init, Update, View
  entries.go                          # DirEntry construction, os.ReadDir wrapper, sorting
  icons.go                            # Icon mapping table, resolution logic
  icons_test.go                       # Icon resolution tests
  watcher.go                          # fsnotify setup, event debouncing, tea.Msg bridge
  watcher_test.go                     # Watcher debounce tests
  styles.go                           # lipgloss style definitions, color palette
  render.go                           # View rendering: path bar, file list, status bar
  render_test.go                      # Render output tests (snapshot-style)
  keys.go                             # Key binding definitions
  sort.go                             # Sort implementations (name, time, size)
  sort_test.go                        # Sort tests
  truncate.go                         # Path truncation logic
  truncate_test.go                    # Truncation tests
  .gitignore                          # Ignore binary, IDE files
  .github/
    workflows/
      ci.yml                          # lint + test on push/PR
  SPEC.md                             # This specification
  CLAUDE.md                           # Agent instructions
  README.md                           # User-facing documentation
  Makefile                            # build, test, lint, install targets
```

### Version Management

Version is a `const` in `main.go`:

```go
const Version = "0.1.0"
```

Displayed via `fp --version`. Bumped manually before tagging releases. Tags follow `vX.Y.Z` format.

### Makefile

```makefile
.PHONY: build test lint install clean

BINARY := fp
VERSION := $(shell grep 'const Version' main.go | cut -d'"' -f2)

build:
	go build -o $(BINARY) .

test:
	go test ./... -v

lint:
	golangci-lint run

install: build
	cp $(BINARY) $(GOPATH)/bin/

clean:
	rm -f $(BINARY)
```

### CI Workflow (`.github/workflows/ci.yml`)

```yaml
name: CI
on:
  pull_request:
    branches: [main]
  push:
    branches: [main]
jobs:
  ci:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: go test ./... -v
      - name: Lint
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest
```

### CLAUDE.md

Agent instructions for Claude Code sessions:

```markdown
# filePick

Single-pane file picker TUI. Go + bubbletea + fsnotify.

## Build & Test
- `go build -o fp .` — build binary
- `go test ./...` — run all tests
- `golangci-lint run` — lint

## Conventions
- gofmt formatting (tabs, no config)
- All files in package `main` (single-binary, no internal packages)
- Icons are hardcoded in icons.go, not loaded from config
- No CGO, no external runtime dependencies
- Tests use stdlib `testing` only
```

## Estimated Size

| Area | Files | LOC |
|------|-------|-----|
| Core TUI (model, render, keys, styles) | 5 | ~450 |
| Filesystem (entries, watcher, sort) | 3 | ~250 |
| Icons (mapping table + resolution) | 1 | ~350 |
| Utilities (truncate) | 1 | ~40 |
| Entry point (main) | 1 | ~60 |
| Tests | 5 | ~400 |
| Infrastructure (Makefile, CI, CLAUDE.md) | 4 | ~100 |
| **Total** | **20** | **~1,650** |

---

## 15. Task Manifest

| ID | Agent | Description | File Scope (read) | File Scope (write) | Depends On | Verify Command |
|----|-------|-------------|--------------------|--------------------|------------|----------------|
| T1 | unix-coder | Initialize Go module with go.mod and dependency declarations (bubbletea v2, lipgloss v2, fsnotify v1) | SPEC.md | go.mod, go.sum | — | `cd /home/n0ko/Programs/filePick && go mod tidy && echo OK` |
| T2 | unix-coder | Implement icon mapping table and resolution logic in icons.go with full Nerd Font glyph table from SPEC | SPEC.md | icons.go, icons_test.go | T1 | `cd /home/n0ko/Programs/filePick && go test -run TestIcon -v` |
| T3 | unix-coder | Implement lipgloss style definitions in styles.go (color palette, selection bar, path bar, status bar styles) | SPEC.md | styles.go | T1 | `cd /home/n0ko/Programs/filePick && go build .` |
| T4 | unix-coder | Implement DirEntry construction, os.ReadDir wrapper, and sorting logic in entries.go and sort.go | SPEC.md | entries.go, sort.go, sort_test.go | T1 | `cd /home/n0ko/Programs/filePick && go test -run TestSort -v` |
| T5 | unix-coder | Implement path truncation logic in truncate.go | SPEC.md | truncate.go, truncate_test.go | T1 | `cd /home/n0ko/Programs/filePick && go test -run TestTruncate -v` |
| T6 | unix-coder | Implement fsnotify watcher with debouncing and tea.Msg bridge in watcher.go | SPEC.md | watcher.go, watcher_test.go | T1 | `cd /home/n0ko/Programs/filePick && go test -run TestWatcher -v` |
| T7 | unix-coder | Implement keybinding definitions in keys.go | SPEC.md | keys.go | T1 | `cd /home/n0ko/Programs/filePick && go build .` |
| T8 | unix-coder | Implement View rendering (path bar, file list with icons, selection bar, status bar) in render.go | SPEC.md, styles.go, icons.go, truncate.go | render.go, render_test.go | T2, T3, T5 | `cd /home/n0ko/Programs/filePick && go test -run TestRender -v` |
| T9 | unix-coder | Implement bubbletea Model (Init, Update, View) in model.go wiring together entries, watcher, render, and keys | SPEC.md, entries.go, watcher.go, render.go, keys.go, sort.go | model.go | T4, T6, T7, T8 | `cd /home/n0ko/Programs/filePick && go build .` |
| T10 | unix-coder | Implement main.go entry point with flag parsing, model initialization, and tea.NewProgram launch | SPEC.md, model.go | main.go | T9 | `cd /home/n0ko/Programs/filePick && go build -o fp . && ./fp --version` |
| T11 | unix-coder | Create .gitignore, Makefile, CLAUDE.md, README.md, and CI workflow | SPEC.md | .gitignore, Makefile, CLAUDE.md, README.md, .github/workflows/ci.yml | T10 | `cd /home/n0ko/Programs/filePick && make build && make test` |
| T12 | code-review | Review complete implementation for correctness, style consistency, race conditions, and adherence to SPEC | All .go files, SPEC.md | — | T11 | `cd /home/n0ko/Programs/filePick && go vet ./... && golangci-lint run` |

## 16. Dependency Graph

```
Phase 1 (single): [T1]
  └── Initialize Go module

Phase 2 (parallel, after Phase 1): [T2, T3, T4, T5, T6, T7]
  ├── T2: Icon mapping table
  ├── T3: Style definitions
  ├── T4: DirEntry + sorting
  ├── T5: Path truncation
  ├── T6: Filesystem watcher
  └── T7: Key bindings

Phase 3 (after T2, T3, T5): [T8]
  └── T8: View rendering

Phase 4 (after T4, T6, T7, T8): [T9]
  └── T9: bubbletea Model

Phase 5 (after T9): [T10]
  └── T10: main.go entry point

Phase 6 (after T10): [T11]
  └── T11: Infrastructure files

Phase 7 — Review (after T11): [T12]
  └── T12: Code review
```

## 17. Target State

Files created:

- `/home/n0ko/Programs/filePick/go.mod`
- `/home/n0ko/Programs/filePick/go.sum`
- `/home/n0ko/Programs/filePick/main.go`
- `/home/n0ko/Programs/filePick/model.go`
- `/home/n0ko/Programs/filePick/entries.go`
- `/home/n0ko/Programs/filePick/icons.go`
- `/home/n0ko/Programs/filePick/icons_test.go`
- `/home/n0ko/Programs/filePick/watcher.go`
- `/home/n0ko/Programs/filePick/watcher_test.go`
- `/home/n0ko/Programs/filePick/styles.go`
- `/home/n0ko/Programs/filePick/render.go`
- `/home/n0ko/Programs/filePick/render_test.go`
- `/home/n0ko/Programs/filePick/keys.go`
- `/home/n0ko/Programs/filePick/sort.go`
- `/home/n0ko/Programs/filePick/sort_test.go`
- `/home/n0ko/Programs/filePick/truncate.go`
- `/home/n0ko/Programs/filePick/truncate_test.go`
- `/home/n0ko/Programs/filePick/.gitignore`
- `/home/n0ko/Programs/filePick/Makefile`
- `/home/n0ko/Programs/filePick/CLAUDE.md`
- `/home/n0ko/Programs/filePick/README.md`
- `/home/n0ko/Programs/filePick/.github/workflows/ci.yml`

Files modified:

- None (greenfield project)

Files deleted:

- None

## 18. Verification Plan

### Per-Task Checks

| Task | Verify Command | Pass Criteria |
|------|---------------|---------------|
| T1 | `go mod tidy` | Exit code 0, go.mod lists bubbletea, lipgloss, fsnotify |
| T2 | `go test -run TestIcon -v` | All icon resolution tests pass |
| T3 | `go build .` | Compiles without errors |
| T4 | `go test -run TestSort -v` | Name/time/size sort tests pass, dirs-first verified |
| T5 | `go test -run TestTruncate -v` | Truncation at component boundaries verified |
| T6 | `go test -run TestWatcher -v` | Debounce coalescing verified |
| T7 | `go build .` | Compiles without errors |
| T8 | `go test -run TestRender -v` | Render output matches expected snapshots |
| T9 | `go build .` | Compiles without errors |
| T10 | `go build -o fp . && ./fp --version` | Prints version string |
| T11 | `make build && make test` | Build succeeds, all tests pass |
| T12 | `go vet ./... && golangci-lint run` | No warnings or errors |

### Integration Check

```bash
cd /home/n0ko/Programs/filePick && \
  go build -o fp . && \
  go test ./... -v && \
  go vet ./... && \
  ./fp --version | grep -q "0.1.0"
```

This verifies the complete application builds, all tests pass, no vet warnings exist, and the version flag works.

### Manual Smoke Test (post-build)

```bash
# Launch in current directory and verify TUI renders
cd /home/n0ko/Programs/filePick && ./fp

# Launch in a specific directory
./fp /tmp

# Verify fsnotify: in one terminal run fp, in another create/delete files
# and observe real-time updates in the TUI
```

### Rollback

```bash
cd /home/n0ko/Programs/filePick && git stash
```

If the project has not been committed yet, remove all generated files:

```bash
cd /home/n0ko/Programs/filePick && \
  rm -f go.mod go.sum main.go model.go entries.go icons.go icons_test.go \
    watcher.go watcher_test.go styles.go render.go render_test.go \
    keys.go sort.go sort_test.go truncate.go truncate_test.go \
    .gitignore Makefile CLAUDE.md README.md && \
  rm -rf .github
```

## 19. Success Criteria (Machine-Verifiable)

- [ ] `cd /home/n0ko/Programs/filePick && go build -o fp .` exits 0
- [ ] `cd /home/n0ko/Programs/filePick && go test ./...` exits 0
- [ ] `cd /home/n0ko/Programs/filePick && go vet ./...` exits 0
- [ ] `cd /home/n0ko/Programs/filePick && ./fp --version` outputs `fp 0.1.0`
- [ ] `cd /home/n0ko/Programs/filePick && ./fp --help` outputs usage text containing `--hidden`, `--sort`, `--no-icons`
- [ ] File `/home/n0ko/Programs/filePick/icons.go` contains at least 50 icon mapping entries
- [ ] File `/home/n0ko/Programs/filePick/watcher.go` imports `github.com/fsnotify/fsnotify`
- [ ] File `/home/n0ko/Programs/filePick/model.go` imports `github.com/charmbracelet/bubbletea`
- [ ] File `/home/n0ko/Programs/filePick/styles.go` imports `github.com/charmbracelet/lipgloss`
- [ ] File `/home/n0ko/Programs/filePick/go.mod` contains `module github.com/n0ko/filePick`
- [ ] `cd /home/n0ko/Programs/filePick && grep -c 'func Test' *_test.go | awk -F: '{s+=$2}END{print s}'` outputs a number >= 10
- [ ] File `/home/n0ko/Programs/filePick/.github/workflows/ci.yml` exists
- [ ] File `/home/n0ko/Programs/filePick/Makefile` contains targets `build`, `test`, `lint`, `install`

## Agent Assignments

| Task | Agent | Rationale |
|------|-------|-----------|
| Go module initialization | `unix-coder` | Boilerplate Go project setup |
| Icon mapping table (350 LOC) | `unix-coder` | Data-heavy file, mechanical translation from spec table |
| Style definitions | `unix-coder` | lipgloss API calls, direct from spec color palette |
| DirEntry + sorting | `unix-coder` | Filesystem stdlib usage, sort implementations |
| Path truncation | `unix-coder` | String manipulation utility |
| Filesystem watcher | `unix-coder` | fsnotify integration, goroutine bridge to bubbletea |
| Key bindings | `unix-coder` | bubbletea key map definition |
| View rendering | `unix-coder` | lipgloss composition, the visual core of the app |
| bubbletea Model | `unix-coder` | Elm architecture wiring, the behavioral core |
| Entry point | `unix-coder` | Flag parsing, program launch |
| Infrastructure | `unix-coder` | Makefile, CI, docs — boilerplate |
| Code review | `code-review` | Final quality gate before merge |

## Execution Order

```
Phase 1: Project Bootstrap
  └── T1: Initialize Go module (agent: unix-coder)

Phase 2: Parallel Component Implementation [blocked by Phase 1]
  ├── T2: Icon mapping table (agent: unix-coder)
  ├── T3: Style definitions (agent: unix-coder)
  ├── T4: DirEntry + sorting (agent: unix-coder)    [parallel]
  ├── T5: Path truncation (agent: unix-coder)        [parallel]
  ├── T6: Filesystem watcher (agent: unix-coder)     [parallel]
  └── T7: Key bindings (agent: unix-coder)           [parallel]

Phase 3: Rendering [blocked by T2, T3, T5]
  └── T8: View rendering (agent: unix-coder)

Phase 4: Model Assembly [blocked by T4, T6, T7, T8]
  └── T9: bubbletea Model (agent: unix-coder)

Phase 5: Entry Point [blocked by Phase 4]
  └── T10: main.go (agent: unix-coder)

Phase 6: Infrastructure [blocked by Phase 5]
  └── T11: Docs, CI, Makefile (agent: unix-coder)

Phase 7: Review [blocked by Phase 6]
  └── T12: Code review (agent: code-review)
```

Recommended directive: `/pai` — plan-then-implement pipeline, as this is a single-binary greenfield project where phases naturally serialize.

## Failure Modes

| Failure | Detection | Recovery |
|---------|-----------|----------|
| fsnotify fails to initialize (no inotify support) | `watcher.Add()` returns error | Print warning to stderr, fall back to 1-second polling loop |
| Terminal too narrow for path bar | `termWidth < 20` in WindowSizeMsg | Truncate path to `...` only, skip status bar if height < 5 |
| Permission denied on directory read | `os.ReadDir()` returns EACCES | Display error in status bar, keep last known entry list |
| Nerd Font not installed | User sees broken glyphs | `--no-icons` flag or `FP_ICONS=0` switches to ASCII: `>` for dirs, `-` for files |
| Rapid filesystem events (e.g., `rm -rf`) | Hundreds of fsnotify events per second | 50ms debounce coalesces to single re-read |
| Symlink cycle | `os.Stat` follows symlinks infinitely | Use `os.Lstat` for entry listing, resolve symlink target for display only (one level) |
| Directory deleted while viewing | fsnotify DELETE_SELF event | Navigate to parent directory automatically |
| Empty directory | `os.ReadDir()` returns empty slice | Render "(empty)" centered in file list area |
| Binary built without Nerd Font support in terminal | Glyphs render as boxes/tofu | Document Nerd Font requirement in README and --help output |

## Open Questions

No open questions remain. All design decisions have been made in this specification.
