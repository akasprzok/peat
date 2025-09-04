# Peat

A CLI for querying Prometheus

## Examples

### Query

```
peat query 'sum(up) by (job)'
```

#### Options

| Long | Short | Description |
|------|-------|-------------|
| --timeout | -t | Timeout for Prometheus Query |
| --prometheus-url | -p | URL of the Prometheus endpoint |
| --output         | -o | Output format. Defaults to "graph". Other choices are "table", "json", and "yaml" |

### Query Range

```
peat query-range 'sum(up) by (job)'
```

#### Options

| Long | Short | Description |
|------|-------|-------------|
| --timeout | -t | Timeout for Prometheus Query |
| --prometheus-url | -p | URL of the Prometheus endpoint |
| --range          | -r | Time range of query. Defaults to "1h" |
| --output         | -o | Output format. Defaults to "graph". Other choices are "json" and "yaml" |

### Series

```
peat series 'up'
```

#### Options

| Long | Short | Description |
|------|-------|-------------|
| --timeout | -t | Timeout for Prometheus Query |
| --prometheus-url | -p | URL of the Prometheus endpoint |
| --limit          | -l | Limits the number of returned series. Defaults to "100" |
| --output         | -o | Output format. Defaults to "json". Other option is "yaml" |

## Environment Variables

Many CLI options can also be defined as environment variables

| Env Var             | Description                    |
|---------------------|--------------------------------|
| PEAT_PROMETHEUS_URL | URL of the Prometheus endpoint |
| PEAT_PROMETHEUS_TIMEOUT | Prometheus query timeout. Defaults to 60s |
