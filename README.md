# Peat

> An interactive TUI for exploring Prometheus metrics

[![CI](https://github.com/akasprzok/peat/workflows/CI/badge.svg)](https://github.com/akasprzok/peat/actions?query=workflow%3ACI)
[![Security](https://github.com/akasprzok/peat/workflows/Security/badge.svg)](https://github.com/akasprzok/peat/actions?query=workflow%3ASecurity)
[![Go Report Card](https://goreportcard.com/badge/github.com/akasprzok/peat)](https://goreportcard.com/report/github.com/akasprzok/peat)

Peat is a terminal user interface for querying and visualizing Prometheus metrics. It provides bar charts, time series graphs, and interactive tables rendered directly in your terminal with vim-style navigation.

## Features

- **Terminal-native visualizations** - Bar charts and time series graphs using ntcharts
- **Mode-based interface** - Switch between /query, /query_range, and /series modes with `Tab`
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

1. **/query** - Execute instant queries and display results as a bar chart
2. **/query_range** - Execute range queries over time and display as a time series graph
3. **/series** - Browse series matching label selectors in an interactive table

### Workflow

1. Launch `peat` (starts in insert mode)
2. Type your PromQL query
3. Press `Enter` to execute (exits to normal mode)
4. Press `i` to enter interactive mode and navigate results
5. Press `Esc` to exit interactive mode
6. Press `/` to edit your query again
7. Press `Tab` to switch between modes
8. Press `q` to quit

## Key Bindings

Peat uses vim-style modal editing with **Insert Mode** (for editing queries) and **Normal Mode** (for navigation and commands).

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
