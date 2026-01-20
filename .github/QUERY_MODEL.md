# Query Model - Bubbletea Implementation

## Overview

The `peat query` command has been wrapped in a Bubbletea model to provide an interactive, terminal-based UI with loading states and better error handling.

## Features

### 1. **Loading Indicator**
- Displays a spinner while the Prometheus query is executing
- Shows the query being executed

### 2. **State Management**
The model manages four states:
- `stateLoading` - Query is being executed
- `stateSuccess` - Query completed successfully
- `stateError` - Query failed with an error
- `stateShowingTable` - Interactive table view is active

### 3. **Multiple Output Formats**

#### Graph Output (default)
```bash
peat query 'up'
```
- Shows a loading spinner
- Displays warnings if any
- Renders a bar chart of the results

#### Table Output
```bash
peat query 'up' --output table
```
- Shows a loading spinner
- Displays warnings if any
- Opens an interactive table view
- Supports filtering, sorting, and navigation

#### JSON Output
```bash
peat query 'up' --output json
```
- Shows a loading spinner
- Outputs formatted JSON with data, warnings, and errors

#### YAML Output
```bash
peat query 'up' --output yaml
```
- Shows a loading spinner
- Outputs formatted YAML with data, warnings, and errors

## Implementation Details

### Files
- `internal/commands/query.go` - Main query command entry point
- `internal/commands/query_model.go` - Bubbletea model implementation

### Key Components

#### QueryModel Structure
```go
type QueryModel struct {
    promClient   prometheus.Client
    query        string
    timeout      time.Duration
    output       string
    state        queryState
    spinner      spinner.Model
    warnings     v1.Warnings
    vector       model.Vector
    err          error
    width        int
    height       int
    tableModel   *tables.Model
    chartContent string
    quitting     bool
}
```

#### Message Types
- `queryResultMsg` - Contains the query results (warnings, vector, error)
- `tea.WindowSizeMsg` - Terminal window size changes
- `tea.KeyMsg` - Keyboard input
- `spinner.TickMsg` - Spinner animation updates

### Complexity Management

The Update function has been split into smaller helper functions to maintain low cyclomatic complexity:
- `handleWindowSize` - Process window resize events
- `handleKeyMsg` - Process keyboard input
- `handleQueryResult` - Process query completion
- `handleOutputFormat` - Route to appropriate output handler
- `handleGraphOutput` - Prepare graph visualization
- `handleTableOutput` - Setup interactive table
- `handleSpinnerTick` - Update spinner animation
- `updateTableModel` - Delegate updates to table model

## User Experience

### Loading State
```
⣾ Executing query: sum(up) by (job)
```

### Success with Warnings
```
Warnings:
  • Some metric had missing data

[Bar chart visualization]
```

### Error State
```
Error: connection refused to Prometheus endpoint
```

### Interactive Table
- Press `/` + letters to filter
- Arrow keys to navigate
- `q` or `ctrl+c` to quit

## Benefits

1. **Responsive UI** - Users get immediate feedback that their query is being processed
2. **Better Error Handling** - Errors are displayed in a user-friendly format with styling
3. **Consistent Experience** - All output formats now use the same loading and error handling flow
4. **Keyboard Control** - Standard terminal navigation (`q`, `ctrl+c`)
5. **Visual Feedback** - Spinner animation shows the query is in progress
6. **Warning Display** - Prometheus warnings are clearly highlighted

## Dependencies

- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/bubbles/spinner` - Loading spinner component
- `github.com/charmbracelet/lipgloss` - Styling and formatting

## Future Enhancements

Potential improvements:
- Add query history navigation
- Support query cancellation
- Show query execution time
- Add progress indicators for long-running queries
- Cache recent query results
