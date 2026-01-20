# Peat

> A beautiful CLI for querying Prometheus with terminal-native visualizations

[![CI](https://github.com/akasprzok/peat/workflows/CI/badge.svg)](https://github.com/akasprzok/peat/actions?query=workflow%3ACI)
[![Security](https://github.com/akasprzok/peat/workflows/Security/badge.svg)](https://github.com/akasprzok/peat/actions?query=workflow%3ASecurity)
[![Code Quality](https://github.com/akasprzok/peat/workflows/Code%20Quality/badge.svg)](https://github.com/akasprzok/peat/actions?query=workflow%3A%22Code+Quality%22)
[![Go Report Card](https://goreportcard.com/badge/github.com/akasprzok/peat)](https://goreportcard.com/report/github.com/akasprzok/peat)

Peat is a command-line tool that makes querying Prometheus metrics easy and visually appealing. It provides instant visualization of metrics with bar charts and time series graphs rendered directly in your terminal.

## Features

- 📊 **Terminal-native visualizations** - Beautiful bar charts and time series graphs using ntcharts
- 📋 **Interactive tables** - Browse query results with an interactive table interface
- 🎨 **Multiple output formats** - Choose between graph, table, JSON, or YAML output
- ⚡ **Fast & lightweight** - Written in Go for performance
- 🔍 **Query formatting** - Built-in PromQL query formatter
- 🌐 **Flexible configuration** - Use CLI flags or environment variables

## Installation

### Pre-built binaries

Download the latest release from the [releases page](https://github.com/akasprzok/peat/releases).

### From source

```bash
go install github.com/akasprzok/peat@latest
```

## Quick Start

Set your Prometheus URL as an environment variable:

```bash
export PEAT_PROMETHEUS_URL=http://localhost:9090
```

Run an instant query with a bar chart:

```bash
peat query 'sum(up) by (job)'
```

Query a time range with a time series graph:

```bash
peat query-range 'rate(http_requests_total[5m])'
```

## Commands

### `query` - Instant Query

Execute an instant Prometheus query and visualize the results.

```bash
peat query 'sum(rate(http_requests_total[5m])) by (method)'
```

**Options:**

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--prometheus-url` | `-p` | URL of the Prometheus endpoint | (from env) |
| `--timeout` | `-t` | Timeout for Prometheus query | `60s` |
| `--output` | `-o` | Output format: `graph`, `table`, `json`, `yaml` | `graph` |

**Examples:**

```bash
# Bar chart visualization (default)
peat query 'up'

# Interactive table
peat query 'up' --output table

# JSON output
peat query 'up' --output json

# YAML output
peat query 'up' --output yaml
```

### `query-range` - Range Query

Execute a range query over a specified time period and display as a time series.

```bash
peat query-range 'rate(cpu_usage[5m])' --range 6h
```

**Options:**

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--prometheus-url` | `-p` | URL of the Prometheus endpoint | (from env) |
| `--timeout` | `-t` | Timeout for Prometheus query | `60s` |
| `--range` | `-r` | Time range of query (e.g., `1h`, `6h`, `1d`) | `1h` |
| `--output` | `-o` | Output format: `graph`, `json`, `yaml` | `graph` |

**Examples:**

```bash
# Time series for the last hour (default)
peat query-range 'up'

# Time series for the last 24 hours
peat query-range 'memory_usage' --range 24h

# JSON output
peat query-range 'cpu_usage' --output json
```

### `series` - List Series

List all series matching a given label matcher.

```bash
peat series 'up'
peat series '{job="prometheus"}'
```

**Options:**

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--prometheus-url` | `-p` | URL of the Prometheus endpoint | (from env) |
| `--timeout` | `-t` | Timeout for Prometheus query | `60s` |
| `--limit` | `-l` | Limit the number of returned series | `100` |
| `--output` | `-o` | Output format: `json`, `yaml` | `json` |

**Examples:**

```bash
# List series for a metric
peat series 'http_requests_total'

# List with label matchers
peat series '{job="api",method="GET"}'

# Limit results
peat series 'up' --limit 10

# YAML output
peat series 'up' --output yaml
```

### `format-query` - Format Query

Format a PromQL query for better readability.

```bash
peat format-query 'sum(rate(http_requests_total[5m]))by(job)'
```

**Example:**

```bash
# Format a complex query
peat format-query 'histogram_quantile(0.95,sum(rate(http_request_duration_seconds_bucket[5m]))by(le,job))'
```

## Configuration

### Environment Variables

Configure peat using environment variables instead of command-line flags:

| Variable | Description | Default |
|----------|-------------|---------|
| `PEAT_PROMETHEUS_URL` | URL of the Prometheus endpoint | - |
| `PEAT_PROMETHEUS_TIMEOUT` | Prometheus query timeout | `60s` |

**Example:**

```bash
export PEAT_PROMETHEUS_URL=http://prometheus.example.com:9090
export PEAT_PROMETHEUS_TIMEOUT=30s

peat query 'up'
```

## Output Formats

Peat supports multiple output formats to suit different use cases:

- **`graph`** - Terminal-native visualizations (bar charts for instant queries, time series for range queries)
- **`table`** - Interactive table view for browsing results
- **`json`** - Machine-readable JSON format for scripting and automation
- **`yaml`** - Human-friendly YAML format

## License

See [LICENSE](LICENSE) file for details.
