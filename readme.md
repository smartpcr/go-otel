the library supports: logging, metrics and tracing.

- logger is composite of console, file and opentelemetry, 
  - file logger is configured of rolling logs that keeps most recent 10 files of 10MB each.
  - opentelemetry logger is configured to send logs to jaeger, it's integrated with span trace_id.
- metrics is configured to send metrics to either prometheus or jaeger collector.
  - when sent to prometheus, prometheus endpoint is read from its registry.
  - when sent to jaeger collector, it then forwards to prometheus via its exporter.
- tracing is configured to send traces to jaeger collector.

## How to use

### 1. Startup jaeger collector, prometheus and grafana

```bash
docker-compose up -d
```

### 2. Modify config file to point to correct endpoints

```bash
# config.yaml
logger:
  file:
    path: ./logs
  opentelemetry:
    endpoint: http://localhost:14268/api/traces
metrics:
    prometheus:
        endpoint: http://localhost:9090
    jaeger:
        endpoint: http://localhost:14250
tracing:
    jaeger:
        endpoint: http://localhost:14268/api/traces
```

### 3. Run the example

```bash
go run main.go
```
