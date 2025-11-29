# Query Range Model - Bubbletea Implementation

## Overview

The `peat query-range` command has been wrapped in a Bubbletea model to provide an interactive, terminal-based UI with loading states and better error handling, similar to the `query` command.

## Features

### 1. **Loading Indicator**
- Displays a spinner while the Prometheus range query is executing
- Shows the query and time range being executed

### 2. **State Management**
The model manages three states:
- `stateRangeLoading` - Query is being executed
- `stateRangeSuccess` - Query completed successfully
- `stateRangeError` - Query failed with an error

### 3. **Multiple Output Formats**

#### Graph Output (default)
```bash
peat query-range 'up'
peat query-range 'rate(http_requests_total[5m])' --range 6h
```
- Shows a loading spinner with the query and time range
- Displays warnings if any
- Renders a time series chart of the results

#### JSON Output
```bash
peat query-range 'up' --output json
```
- Shows a loading spinner
- Outputs formatted JSON with time series data, warnings, and errors

#### YAML Output
```bash
peat query-range 'up' --output yaml
```
- Shows a loading spinner
- Outputs formatted YAML with time series data, warnings, and errors

## Implementation Details

### Files
- `internal/commands/query_range.go` - Main query-range command entry point
- `internal/commands/query_range_model.go` - Bubbletea model implementation

### Key Components

#### QueryRangeModel Structure
```go
type QueryRangeModel struct {
    promClient   prometheus.Client
    query        string
    timeRange    time.Duration
    timeout      time.Duration
    output       string
    state        queryRangeState
    spinner      spinner.Model
    matrix       model.Matrix
    warnings     v1.Warnings
    err          error
    width        int
    height       int
    chartContent string
    quitting     bool
}
```

#### Message Types
- `queryRangeResultMsg` - Contains the query results (matrix, warnings, error)
- `tea.WindowSizeMsg` - Terminal window size changes
- `tea.KeyMsg` - Keyboard input
- `spinner.TickMsg` - Spinner animation updates

### Complexity Management

The Update function has been split into smaller helper functions to maintain low cyclomatic complexity:
- `handleRangeWindowSize` - Process window resize events
- `handleRangeKeyMsg` - Process keyboard input
- `handleRangeQueryResult` - Process query completion
- `handleRangeOutputFormat` - Route to appropriate output handler
- `handleRangeGraphOutput` - Prepare time series visualization
- `handleRangeSpinnerTick` - Update spinner animation

## User Experience

### Loading State
```
⣾ Executing range query: rate(http_requests_total[5m]) (range: 6h)
```

### Success with Warnings
```
Warnings:
  • Some data points were missing in the time range

[Time series chart visualization]
```

### Error State
```
Error: connection refused to Prometheus endpoint
```

## Differences from Query Model

The query-range model differs from the query model in several ways:

1. **Data Type**
   - Uses `model.Matrix` (time series data) instead of `model.Vector` (instant values)
   - Each data point includes a timestamp

2. **Visualization**
   - Uses `Timeseries` chart instead of `Barchart`
   - Shows data over time with line graphs

3. **No Table Output**
   - Range queries return time series data which is not well-suited for the current table view
   - Supports: `graph`, `json`, `yaml` (no `table` option)

4. **Query Parameters**
   - Includes time range parameter (e.g., `1h`, `6h`, `24h`)
   - Automatically calculates start/end times based on range
   - Uses 1-minute step interval for data points

## Command Examples

### Basic Usage
```bash
# Query the last hour (default)
peat query-range 'up'

# Query a specific time range
peat query-range 'up' --range 24h
peat query-range 'rate(cpu_usage[5m])' -r 6h

# Different output formats
peat query-range 'memory_usage' --output json
peat query-range 'disk_usage' -o yaml
```

### Complex Queries
```bash
# Rate over time
peat query-range 'rate(http_requests_total[5m])' --range 12h

# Aggregations over time
peat query-range 'sum(rate(errors[1m])) by (service)' --range 3h

# Multiple series
peat query-range 'node_cpu_seconds_total' --range 1h
```

## Benefits

1. **Responsive UI** - Users get immediate feedback that their query is being processed
2. **Better Error Handling** - Errors are displayed in a user-friendly format with styling
3. **Consistent Experience** - All output formats now use the same loading and error handling flow
4. **Keyboard Control** - Standard terminal navigation (`q`, `ctrl+c`)
5. **Visual Feedback** - Spinner animation shows the query is in progress
6. **Warning Display** - Prometheus warnings are clearly highlighted
7. **Time Context** - Loading message shows the time range being queried

## Dependencies

- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/bubbles/spinner` - Loading spinner component

## Future Enhancements

Potential improvements:
- Add custom step interval option
- Support absolute time ranges (start/end timestamps)
- Add query history navigation
- Support query cancellation
- Show query execution time and data point count
- Add interactive time range selection
- Implement table view for time series data with pagination

