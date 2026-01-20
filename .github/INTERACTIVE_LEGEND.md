# Interactive Mode Feature

## Overview

The `peat query-range` command now features an interactive mode that allows you to focus on individual time series by greying out all others. This makes it easier to analyze specific metrics in complex charts with multiple series.

## How to Use

### Basic Usage

1. **Run a range query**:
   ```bash
   peat query-range 'up' --range 6h
   ```

2. **Press `i` to enter interactive mode**:
   - The legend border changes from blue to pink
   - The table becomes navigable

3. **Navigate through the legend**:
   - `j` or `down`: Move down one row
   - `k` or `up`: Move up one row
   - `l` or `right`: Move forward one page
   - `h` or `left`: Move backward one page
   - As you navigate, the selected series is highlighted in the chart
   - All other series are greyed out

4. **Press `i` again to exit interactive mode**:
   - All series return to their original colors
   - Legend border returns to blue

5. **Press `q` to quit**

## Visual States

### Normal State (All Series Visible)
```
┌────────────────────────────────────────┐
│ [All series shown in different colors] │
│ Series 1: Blue                          │
│ Series 2: Green                         │
│ Series 3: Yellow                        │
└────────────────────────────────────────┘

┌───┬─────────────────────────────┐ (Blue border)
│   │ Metric                      │
├───┼─────────────────────────────┤
│ █ │ metric_1                    │
│ █ │ metric_2                    │
│ █ │ metric_3                    │
└───┴─────────────────────────────┘

Press i for interactive mode, q to quit
```

### Interactive Mode (Series Selected)
```
┌────────────────────────────────────────┐
│ Series 1: Bold White (highlighted)      │
│ Series 2: Grey (dimmed)                 │
│ Series 3: Grey (dimmed)                 │
└────────────────────────────────────────┘

┌───┬─────────────────────────────┐ (Pink border)
│   │ Metric                      │
├───┼─────────────────────────────┤
│ █ │ metric_1                    │ ← Selected/highlighted
│ █ │ metric_2                    │
│ █ │ metric_3                    │
└───┴─────────────────────────────┘

Interactive mode • j/k: row • h/l: page • i: exit • q: quit
```

## Keyboard Controls

| Key | Action |
|-----|--------|
| `i` | Toggle interactive mode on/off |
| `j` or `↓` | Move down one row (when in interactive mode) |
| `k` or `↑` | Move up one row (when in interactive mode) |
| `l` or `→` | Move forward one page (when in interactive mode) |
| `h` or `←` | Move backward one page (when in interactive mode) |
| `q` or `ctrl+c` | Quit |

## Implementation Details

### Chart Re-rendering

When navigating through the legend:
1. The `selectedIndex` is updated based on the highlighted row
2. The chart is regenerated using `TimeseriesSplitWithSelection(matrix, width, selectedIndex)`
3. Non-selected series are rendered in grey (color "240")
4. **Selected series is rendered in bold, bright white** (color "231") for high contrast
5. **Selected series is drawn a second time** at the end to ensure it appears on top for better visibility

### Legend Table

- Built using `github.com/evertras/bubble-table`
- Shows up to 5 rows at a time
- Displays:
  - Color indicator (█)
  - Metric name
- Supports pagination if more than 5 series

### Files Modified

1. `internal/charts/timeseries.go`:
   - Added `TimeseriesSplitWithSelection(matrix, width, selectedIndex)` function
   - Applies grey color to non-selected series
   - Renders selected series in bold, bright white for high contrast
   - Draws selected series a second time at the end (on top) for better visibility

2. `internal/commands/query_range_model.go`:
   - Added `legendFocused` and `selectedIndex` fields
   - Enhanced keyboard navigation:
     - `i`: Toggle interactive mode
     - `j`/`k`: Row-by-row navigation
     - `h`/`l`: Page-by-page navigation
   - Added `updateSelectedFromTable()` method
   - Added `regenerateChart()` method
   - Dynamic help text based on state

## Use Cases

### Analyzing Specific Metrics in Busy Charts

When you have many overlapping time series:
```bash
peat query-range 'node_cpu_seconds_total' --range 1h
```

This might show 8+ CPU cores. You can:
1. Press `i` to enter interactive mode
2. Use `j`/`k` to navigate through individual cores (row by row)
3. Use `l`/`h` to jump forward/backward through pages (5 cores at a time)
4. Analyze each core's behavior in isolation
5. Press `i` to exit and see all cores together again

### Comparing Series

1. Focus on Series A (press `i`, navigate to it)
2. Observe its pattern
3. Navigate to Series B with `j`
4. Compare the pattern
5. Exit interactive mode to see the relationship

## Benefits

1. **Reduced Visual Clutter** - Focus on one series at a time
2. **Better Analysis** - Easier to spot patterns in individual metrics
3. **Interactive Exploration** - Quick navigation between series
4. **Reversible** - Easily return to full view
5. **Visual Feedback** - Border color and help text show current state

## Future Enhancements

Potential improvements:
- Multi-select support (select multiple series to compare)
- Color customization for selected vs dimmed series
- Zoom to selected series time range
- Export selected series data
- Search/filter in legend
