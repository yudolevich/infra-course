## Opentelemetry

```{image} ../img/otel.svg
:width: 200px
```

### Observability

```{image} ../img/otel-slides1.svg
:width: 700px
```

### Signals
```{revealjs-fragments}
* Traces
* Metrics
* Logs
```

### Components
```{revealjs-fragments}
* Specification
* Instrumentation
* Collector
```

### Instrumentation
```{revealjs-fragments}
* Zero-code
* Code-based
```

### Collector

```{image} ../img/otel-slides2.svg
:width: 700px
```

### Collector
```{revealjs-code-block} yaml
---
data-line-numbers: 1-7|11-13|8-9|15-28
---
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318
processors:
  batch:

exporters:
  otlp:
    endpoint: otelcol:4317

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp]
    metrics:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp]
    logs:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp]
```
