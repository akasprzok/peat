# CLAUDE.md

## Project Overview

Peat is a TUI (Terminal User Interface) for querying Prometheus metrics with terminal-native visualizations. Built with Go, Bubble Tea, and ntcharts.

## Commands

```bash
make build      # Build binary
make test       # Run tests with race detection and coverage
make lint       # Run golangci-lint (50+ linters)
make fmt        # Format code
make check      # Run all checks (fmt, vet, lint, test)
make install    # Install to GOPATH/bin
```

## Architecture

### Package Structure

```
internal/
├── commands/          # TUI application (~1,500 LOC)
│   ├── model.go       # TUIModel state management
│   ├── view.go        # Rendering logic
│   ├── update.go      # Message handling
│   ├── types.go       # QueryMode, TUIState, Result messages
│   ├── queries.go     # Query execution
│   ├── mode.go        # Mode interface definition
│   ├── mode_instant.go   # /query mode
│   ├── mode_range.go     # /query_range mode
│   ├── mode_series.go    # /series mode
│   ├── mode_labels.go    # /labels mode
│   ├── constants.go   # All value constants (see convention note)
│   └── shared.go      # ErrorStyle, WarningStyle, spinner
├── prometheus/        # API client wrapper
│   ├── prometheus.go  # Client interface: Query, QueryRange, Series, LabelNames, LabelValues
│   └── mock_client.go # MockClient for testing
├── charts/            # Terminal visualization
│   ├── barchart.go    # Instant query results
│   ├── timeseries.go  # Range query results
│   ├── constants.go   # All value constants (see convention note)
│   └── colors.go      # Paul Tol's colorblind-accessible palette
└── tables/            # Interactive table display
    └── tables.go      # Series and label tables
```

### Mode Interface Pattern

The app uses a Mode interface for extensibility. Each query mode implements:

```go
type Mode interface {
    Name() string
    Placeholder() string
    Execute(client prometheus.Client, query string, ...) tea.Cmd
    Render(m *TUIModel) string
}
```

Per-mode state is stored in parallel arrays indexed by QueryMode:
- `modeQueries[4]string` - Query text per mode
- `modeStates[4]TUIState` - State (Idle/Loading/Success/Error)
- `modeErrors[4]error` - Error per mode
- `modeDurations[4]time.Duration` - Query duration

### Constants Convention

All non-enum value constants belong in the per-package `constants.go` file. Enum constants (iota) stay in `types.go` alongside their type definitions.

### Adding a New Mode

1. Add constant to `QueryMode` enum in `types.go`
2. Create `mode_<name>.go` implementing the Mode interface
3. Register in the mode registry map in `mode.go`
4. Update array sizes if needed

## Query Modes

| Mode | Key | Description |
|------|-----|-------------|
| `/query` | Tab | Bar chart for instant queries |
| `/query_range` | Tab | Time series graph with interactive legend |
| `/series` | Tab | Interactive table for series by label matchers |
| `/labels` | Tab | Browse label names and values |

## Key Bindings

| Key | Mode | Action |
|-----|------|--------|
| `Tab` | Any | Cycle through query modes |
| `Enter` | Any | Execute query |
| `Esc` | Insert | Exit insert mode |
| `/` | Normal | Enter insert mode |
| `f` | Normal | Format PromQL query |
| `i` | Normal | Toggle interactive mode |
| `j/k` | Interactive | Navigate up/down |
| `h/l` | Interactive | Page up/down |
| `q` | Normal | Quit |

## Testing

**Pattern:** Table-driven tests throughout.

```go
tests := []struct {
    name    string
    input   string
    want    string
    wantErr bool
}{...}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {...})
}
```

**Mock client:** Use `MockClient` in `internal/prometheus/mock_client.go` for testing without a real Prometheus.

## Error Handling

- Wrap errors with context: `fmt.Errorf("doing X: %w", err)`
- Use `ErrorStyle` (red, bold) and `WarningStyle` (orange) from `shared.go`
- Store per-mode errors in `modeErrors` array

## Styling

- Colors defined in `internal/charts/colors.go` using Paul Tol's colorblind-accessible palette
- UI styling via lipgloss throughout
- Shared styles in `internal/commands/shared.go`

## Development Setup

**Requirements:** Go 1.25+

**Environment variables:**
- `PEAT_PROMETHEUS_URL` - Prometheus server URL
- `PEAT_PROMETHEUS_TIMEOUT` - Query timeout (default: 60s)

**Local development:** Use `.envrc` with direnv for automatic env setup.

## CI/CD

- **Pre-commit hooks:** `.pre-commit-config.yaml` runs linting and formatting
- **GitHub Actions:** CI runs on push/PR (tests on Linux, macOS, Windows)
- **Release:** Push a git tag to trigger GoReleaser (`.goreleaser.yaml`)
- **Security:** Weekly govulncheck and CodeQL scans

## Configuration Files

| File | Purpose |
|------|---------|
| `.golangci.yaml` | Linter config (50+ linters, strict settings) |
| `.goreleaser.yaml` | Multi-platform release builds |
| `.pre-commit-config.yaml` | Git hooks for quality checks |

## Key Dependencies

- `github.com/charmbracelet/bubbletea` - Terminal UI framework
- `github.com/charmbracelet/lipgloss` - Terminal styling
- `github.com/NimbleMarkets/ntcharts` - Terminal charts
- `github.com/evertras/bubble-table` - Interactive tables
- `github.com/alecthomas/kong` - CLI argument parsing
- `github.com/prometheus/client_golang` - Prometheus API client
- `github.com/prometheus/prometheus/promql/parser` - PromQL formatting
