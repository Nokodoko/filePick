<!--
 ███████╗██╗██╗     ███████╗██████╗ ██╗ ██████╗██╗  ██╗
 ██╔════╝██║██║     ██╔════╝██╔══██╗██║██╔════╝██║ ██╔╝
 █████╗  ██║██║     █████╗  ██████╔╝██║██║     █████╔╝
 ██╔══╝  ██║██║     ██╔══╝  ██╔═══╝ ██║██║     ██╔═██╗
 ██║     ██║███████╗███████╗██║     ██║╚██████╗██║  ██╗
 ╚═╝     ╚═╝╚══════╝╚══════╝╚═╝     ╚═╝ ╚═════╝╚═╝  ╚═╝
-->

<img src="filepick_banner_wide.jpg" width="100%"/>

<div align="center">
  <img src="https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white"/>
  <img src="https://img.shields.io/badge/Bubbletea-FF75B5?style=for-the-badge&logo=go&logoColor=white"/>
  <img src="https://img.shields.io/badge/Linux-FCC624?style=for-the-badge&logo=linux&logoColor=black"/>
  <img src="https://img.shields.io/badge/License-MIT-00FF9F?style=for-the-badge"/>
</div>

<br>

<div align="center">
  <i>Single-pane file picker TUI for watching git worktrees spawn and disappear in real time.</i>
</div>

<br>

## > cat /etc/project.conf

```
╭──────────────────────────────────────────────────────────╮
│                                                          │
│   PROJECT="filePick"                                     │
│   VERSION="0.1.0"                                        │
│   LICENSE="MIT"                                          │
│   LANGUAGE="Go 1.25"                                     │
│   FRAMEWORK="bubbletea v2 + lipgloss v2"                 │
│   WATCHER="fsnotify (inotify)"                           │
│                                                          │
╰──────────────────────────────────────────────────────────╯
```

Built for one purpose: stare at a directory and see git worktrees appear as agents create them, then vanish as they merge and clean up. A live dashboard for swarm-driven development.

## > filePick --features

| Feature | Description |
|---------|-------------|
| Single-pane listing | Zero-chrome file listing with Nerd Font icons |
| Real-time updates | Filesystem events via fsnotify (inotify on Linux) |
| Git worktree awareness | Watches `.claude/worktrees/` directories |
| Fuzzy search | Press `/` for ripgrep-powered fuzzy file search |
| Vim keybindings | `hjkl` navigation, `g`/`G` jump, `.` toggle hidden |
| Dark palette | Terminal-optimized color scheme |
| Zero config | Flags and environment variables only |
| Shell composable | Prints selected path to stdout |

## > filePick architecture --map

```
+------------------------------------------------------------------------+
|                              filePick                                   |
|                                                                         |
|   +----------------+     +----------------+     +----------------+      |
|   |   main.go      |────>|   model.go     |────>|   render.go    |      |
|   |  (entry point) |     |  (bubbletea    |     |  (lipgloss     |      |
|   |                |     |   model/update) |     |   views)       |      |
|   +----------------+     +-------+--------+     +----------------+      |
|                                  |                                      |
|                    +-------------+-------------+                        |
|                    |             |             |                         |
|              +-----+----+ +-----+----+ +------+-----+                   |
|              | watcher  | |  git.go  | |  keys.go   |                   |
|              | .go      | | (worktree| | (keybinds) |                   |
|              | (fsnotify| |  detect) | +------------+                   |
|              |  events) | +----------+                                  |
|              +----------+                                               |
|                                                                         |
|   +----------------+     +----------------+     +----------------+      |
|   |   icons.go     |     |   sort.go      |     |  entries.go    |      |
|   |  (nerd font    |     |  (name/time/   |     |  (dir entry    |      |
|   |   codepoints)  |     |   size sort)   |     |   types)       |      |
|   +----------------+     +----------------+     +----------------+      |
+------------------------------------------------------------------------+
```

## > cat /etc/requirements

- Go 1.22+
- A terminal with a [Nerd Font](https://www.nerdfonts.com/) for icons (or use `--no-icons`)
- ripgrep (`rg`) for fuzzy search (optional)

## > ./install --quick

<details>
<summary><b>go install</b></summary>

```bash
go install github.com/n0ko/filePick@latest
```

</details>

<details>
<summary><b>Build from source</b></summary>

```bash
git clone https://github.com/n0ko/filePick.git
cd filePick
go build -o fp .
```

</details>

## > filePick --help

```
fp [path]                     Open TUI in [path] or current directory
  --hidden                    Show hidden files on startup
  --sort <mode>               name|time|size (default: name)
  --no-icons                  Disable Nerd Font icons (ASCII fallback)
  --version                   Print version and exit
  --help                      Print usage and exit
```

### Keybindings

| Key | Action |
|-----|--------|
| `j` / `Down` | Move cursor down |
| `k` / `Up` | Move cursor up |
| `g` / `Home` | Jump to first entry |
| `G` / `End` | Jump to last entry |
| `Enter` / `l` | Enter directory |
| `h` / `Backspace` | Go to parent directory |
| `.` | Toggle hidden files |
| `s` | Cycle sort mode (name -> time -> size) |
| `/` | Fuzzy file search (ripgrep) |
| `q` / `Esc` / `Ctrl-C` | Quit |

### Shell Integration

filePick prints the selected file path to stdout, enabling composition:

```bash
# Pick a file and open it
vim "$(fp ~/Projects)"

# Copy selected path to clipboard
fp | xclip -selection clipboard
```

## > filePick env --vars

| Variable | Effect |
|----------|--------|
| `NO_COLOR=1` | Disable all ANSI color output |
| `FP_ICONS=0` | Disable Nerd Font icons |

## > cat CONTRIBUTING.md

```bash
git clone https://github.com/n0ko/filePick.git
cd filePick
go build -o fp .
go test ./...
```

<img src="https://capsule-render.vercel.app/api?type=waving&color=0:00FF9F,50:00D4FF,100:0D1117&height=120&section=footer" width="100%"/>

<div align="center">
  <img src="https://img.shields.io/badge/License-MIT-00FF9F?style=flat-square"/>
  <br>
  <sub>Built with terminal aesthetics by <a href="https://github.com/n0ko">n0ko</a></sub>
</div>
