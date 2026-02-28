# filePick

Single-pane file picker TUI for watching git worktrees spawn and disappear in real time.

Built for one purpose: stare at a directory and see git worktrees appear as agents create them, then vanish as they merge and clean up. A live dashboard for swarm-driven development.

## Features

- Single-pane, zero-chrome file listing with Nerd Font icons
- Real-time filesystem updates via fsnotify (inotify on Linux)
- Git worktree awareness for `.claude/worktrees/` directories
- Vim-style keybindings
- Dark terminal optimized color palette
- Zero configuration -- flags and environment variables only

## Install

```bash
go install github.com/n0ko/filePick@latest
```

Or build from source:

```bash
git clone https://github.com/n0ko/filePick.git
cd filePick
go build -o fp .
```

## Usage

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
| j / Down | Move cursor down |
| k / Up | Move cursor up |
| g / Home | Jump to first entry |
| G / End | Jump to last entry |
| Enter / l | Enter directory |
| h / Backspace | Go to parent directory |
| . | Toggle hidden files |
| s | Cycle sort mode (name -> time -> size) |
| q / Esc / Ctrl-C | Quit |

### Shell Integration

filePick prints the selected file path to stdout, enabling composition:

```bash
# Pick a file and open it
vim "$(fp ~/Projects)"

# Copy selected path to clipboard
fp | xclip -selection clipboard
```

### Environment Variables

| Variable | Effect |
|----------|--------|
| `NO_COLOR=1` | Disable all ANSI color output |
| `FP_ICONS=0` | Disable Nerd Font icons |

## Requirements

- Go 1.22+
- A terminal with a [Nerd Font](https://www.nerdfonts.com/) for icons (or use `--no-icons`)

## License

MIT
