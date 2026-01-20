# Peat

> An interactive TUI for exploring Prometheus metrics

[![CI](https://github.com/akasprzok/peat/workflows/CI/badge.svg)](https://github.com/akasprzok/peat/actions?query=workflow%3ACI)
[![Security](https://github.com/akasprzok/peat/workflows/Security/badge.svg)](https://github.com/akasprzok/peat/actions?query=workflow%3ASecurity)
[![Code Quality](https://github.com/akasprzok/peat/workflows/Code%20Quality/badge.svg)](https://github.com/akasprzok/peat/actions?query=workflow%3A%22Code+Quality%22)
[![Go Report Card](https://goreportcard.com/badge/github.com/akasprzok/peat)](https://goreportcard.com/report/github.com/akasprzok/peat)

Peat is a terminal user interface for querying and visualizing Prometheus metrics. It provides bar charts, time series graphs, and interactive tables rendered directly in your terminal with vim-style navigation.

## Features

- **Terminal-native visualizations** - Bar charts and time series graphs using ntcharts
- **Mode-based interface** - Switch between Instant, Range, and Series modes with `Tab`
- **Vim-style navigation** - Navigate results with `j/k/h/l` keys
- **Interactive series highlighting** - Focus on individual series in charts and tables
- **Query formatting** - Format PromQL queries with `f` key
- **Fast & lightweight** - Written in Go for performance

## Installation

### Pre-built binaries

Download the latest release from the [releases page](https://github.com/akasprzok/peat/releases).

### From source

```bash
go install github.com/akasprzok/peat@latest
```

## Usage

### Quick Start

Set your Prometheus URL and launch Peat:

```bash
export PEAT_PROMETHEUS_URL=http://localhost:9090
peat
```

Or pass the URL directly:

```bash
peat --prometheus-url=http://localhost:9090
```

### The TUI Interface

Peat provides three query modes, accessible via `Tab`:

1. **Instant** - Execute instant queries and display results as a bar chart
2. **Range** - Execute range queries over time and display as a time series graph
3. **Series** - Browse series matching label selectors in an interactive table

### Workflow

1. Launch `peat`
2. Press `/` to focus the query input
3. Type your PromQL query
4. Press `Enter` to execute
5. Press `i` to enter interactive mode and navigate results
6. Press `Tab` to switch between modes
7. Press `q` to quit

## Key Bindings

| Key | Action |
|-----|--------|
| `Tab` | Cycle through query modes |
| `Enter` | Execute query |
| `/` | Focus query input |
| `f` | Format PromQL query |
| `i` | Enter interactive mode (legend/table navigation) |
| `Esc` | Exit interactive mode |
| `j/k` | Navigate up/down in interactive mode |
| `h/l` | Page up/down in interactive mode |
| `q` | Quit |
| `Ctrl+C` | Force quit |

## Configuration

Peat can be configured using environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `PEAT_PROMETHEUS_URL` | URL of the Prometheus endpoint | - |
| `PEAT_PROMETHEUS_TIMEOUT` | Prometheus query timeout | `60s` |

**Example:**

```bash
export PEAT_PROMETHEUS_URL=http://prometheus.example.com:9090
export PEAT_PROMETHEUS_TIMEOUT=30s

peat
```

## License

See [LICENSE](LICENSE) file for details.
