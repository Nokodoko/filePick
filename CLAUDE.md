# filePick

Single-pane file picker TUI. Go + bubbletea + fsnotify.

## Build & Test
- `go build -o fp .` -- build binary
- `go test ./...` -- run all tests
- `golangci-lint run` -- lint

## Conventions
- gofmt formatting (tabs, no config)
- All files in package `main` (single-binary, no internal packages)
- Icons are hardcoded in icons.go, not loaded from config
- No CGO, no external runtime dependencies
- Tests use stdlib `testing` only
