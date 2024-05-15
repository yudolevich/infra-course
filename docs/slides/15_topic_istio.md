## Istio

```{image} ../img/istio.svg
:width: 200px
```

### Service Mesh

```{image} ../img/istio1-slides.svg
:width: 700px
```

### Service Mesh

```{image} ../img/istio2-slides.svg
:width: 700px
```

### Concepts
```{revealjs-fragments}
* Traffic managment
* Observability
* Security
* Extensibility
```

### Traffic managment

```{image} ../img/istio3-slides.svg
:width: 700px
```

### Traffic managment
```{revealjs-fragments}
* VirtualService
* DestinationRule
* Gateway
* ServiceEntry
```

### Traffic managment
```{revealjs-fragments}
* Timeouts
* Retries
* Circuit breakers
* Fault injection
```

### Security

```{image} ../img/istio4-slides.svg
:width: 700px
```

### Observability

```{image} ../img/istio5-slides.png
:width: 700px
```

### Metrics
```bash
envoy_cluster_internal_upstream_rq{code="2xx",_name="xds-grpc"} 7163
envoy_cluster_upstream_rq_completed{cluster_name="xds-grpc"} 7164
envoy_cluster_ssl_connection_error{cluster_name="xds-grpc"} 0
envoy_cluster_lb_subsets_removed{cluster_name="xds-grpc"} 0
envoy_cluster_internal_upstream_rq_code="503",_name="xds-grpc"} 1
```

### Logs
```yaml
[2020-11-25T21:26:18.409Z] "GET /status/418 HTTP/1.1" 418 -
via_upstream - "-" 0 135 4 4 "-" "curl/7.73.0-DEV" 
"84961386-6d84-929d-98bd-c5aee93b5c88" "httpbin:8000" 
"10.44.1.27:80" outbound|8000||httpbin.foo.svc.cluster.local 
10.44.1.23:37652 10.0.45.184:8000 10.44.1.23:46520 - default
```

### Traces

```{image} ../img/istio6-slides.png
:width: 700px
```

### Kiali

```{image} ../img/istio7-slides.png
:width: 700px
```

### Kiali

```{image} ../img/istio8-slides.png
:width: 700px
```
