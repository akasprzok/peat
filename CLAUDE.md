# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Peat is a CLI tool for querying Prometheus metrics with terminal-native visualizations. It provides bar charts for instant queries and time series graphs for range queries, along with table, JSON, and YAML output formats.

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

- **main.go** - Entry point; parses CLI args via Kong and dispatches to commands
- **internal/commands/** - CLI command implementations using Kong's command pattern
  - Each command struct (e.g., `QueryCmd`, `QueryRangeCmd`) has a `Run(*Context)` method
  - Commands use Bubble Tea models for interactive output (query_model.go, query_range_model.go)
- **internal/prometheus/** - Prometheus API client wrapper
  - Exposes `Client` interface with `Query`, `QueryRange`, and `Series` methods
  - Also contains `FormatQuery` for PromQL formatting via prometheus/promql/parser
- **internal/charts/** - Terminal chart rendering using ntcharts library
  - `barchart.go` for instant query results
  - `timeseries.go` for range query results
- **internal/tables/** - Interactive table display using bubble-table

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
