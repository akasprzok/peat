# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Peat is a TUI (Terminal User Interface) for querying Prometheus metrics with terminal-native visualizations. It provides an interactive interface with mode switching, vim-style keybindings, and real-time query editing.

## Common Commands

```bash
# Build
make build

# Run tests with race detection and coverage
make test

# Lint (requires golangci-lint)
make lint

# Format code
make fmt

# Run all checks (fmt, vet, lint, test)
make check

# Install to GOPATH/bin
make install
```

## Architecture

The codebase follows a standard Go CLI structure with Kong for argument parsing and Bubble Tea for terminal UI:

- **main.go** - Entry point; parses CLI args via Kong and launches the TUI
- **internal/commands/tui_model.go** - Unified Bubble Tea TUI model managing all query modes
- **internal/commands/commands.go** - CLI struct with Kong bindings for URL, timeout, and initial query parameters
- **internal/prometheus/** - Prometheus API client wrapper
  - Exposes `Client` interface with `Query`, `QueryRange`, and `Series` methods
  - Also contains `FormatQuery` for PromQL formatting via prometheus/promql/parser
- **internal/charts/** - Terminal chart rendering using ntcharts library
  - `barchart.go` for instant query results
  - `timeseries.go` for range query results
- **internal/tables/** - Interactive table display using bubble-table

## TUI Overview

Peat provides three query modes, switchable via `Tab`:

- **/query** - Bar chart visualization for instant queries
- **/query_range** - Time series graph with interactive legend for range queries
- **/series** - Interactive table for browsing series by label matchers

## Key Bindings

Peat uses vim-style modal editing with **Insert Mode** (for editing queries) and **Normal Mode** (for navigation and commands). The app starts in insert mode so users can immediately type.

| Key | Mode | Action |
|-----|------|--------|
| `Tab` | Any | Cycle through query modes |
| `Enter` | Any | Execute query (exits insert mode) |
| `Esc` | Insert | Exit insert mode (return to normal mode) |
| `Esc` | Interactive | Exit interactive mode |
| `/` | Normal | Enter insert mode (edit query) |
| `f` | Normal | Format PromQL query |
| `i` | Normal | Toggle interactive mode (legend/table) |
| `j/k` | Interactive | Navigate up/down |
| `h/l` | Interactive | Page up/down |
| `q` | Normal | Quit |
| `Ctrl+C` | Any | Force quit |

## Key Dependencies

- `github.com/alecthomas/kong` - CLI argument parsing
- `github.com/charmbracelet/bubbletea` - Terminal UI framework
- `github.com/NimbleMarkets/ntcharts` - Terminal charts
- `github.com/evertras/bubble-table` - Interactive tables
- `github.com/prometheus/client_golang` - Prometheus API client
- `github.com/prometheus/prometheus/promql/parser` - PromQL parsing/formatting

## Environment Variables

- `PEAT_PROMETHEUS_URL` - Prometheus server URL (used by commands via Kong env binding)
- `PEAT_PROMETHEUS_TIMEOUT` - Query timeout (default: 60s)
