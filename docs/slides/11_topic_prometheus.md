## Prometheus

```{image} ../img/prometheus.svg
:width: 200px
```

### Concepts
```{revealjs-fragments}
* Metrics
* TSDB
* Exporters
* Service Discovery
* PromQL
* Visualization
* Alerting
```

### Architecture

```{image} ../img/prometheus-arch.png
:width: 700px
```

### Metrics

```{revealjs-fragments}
* `<metric_name>{<label>=<value>, ...}`
* Samples: [float64 value, millisecond timestamp]
* Time Series
```

### Metric Types

```{revealjs-fragments}
* Counter
* Gauge
* Histogram
* Summary
```

### PromQL

```{revealjs-code-block} bash
---
data-line-numbers: 1|2|3|4-6|7
---
http_requests_total
http_requests_total{job="apiserver", handler="/api/comments"}
rate(http_requests_total[5m])
sum by (job) (
  rate(http_requests_total[5m])
)
(instance_memory_limit - instance_memory_usage) / 1024 / 1024
```

### Visualization

```{image} ../img/grafana1.png
:width: 700px
```

### Visualization

```{image} ../img/grafana2.png
:width: 650px
```

### Alerting
```{revealjs-fragments}
* Rules
* Notifications
* Grouping
* Inhibition
* Silences
```

### Alerting Rules
```yaml
groups:
- name: example
  rules:
  - alert: HighRequestLatency
    expr: job:request_latency_seconds:mean5m{job="myjob"} > 0.5
    for: 10m
    labels:
      severity: page
    annotations:
      summary: High request latency
```
