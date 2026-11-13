# Fly Grafana PromQL

Queries abaixo assumem prefixo default `go_fiber_api` e app `go-fiber-api`.

## HTTP

Request rate (RPS):

```promql
sum(rate(go_fiber_api_http_requests_total{app="go-fiber-api"}[5m]))
```

Error rate (5xx):

```promql
sum(rate(go_fiber_api_http_requests_total{app="go-fiber-api",status=~"5.."}[5m]))
/
sum(rate(go_fiber_api_http_requests_total{app="go-fiber-api"}[5m]))
```

p99 latency:

```promql
histogram_quantile(
  0.99,
  sum by (le) (
    rate(go_fiber_api_http_request_duration_seconds_bucket{app="go-fiber-api"}[5m])
  )
)
```

In-flight requests:

```promql
go_fiber_api_http_requests_in_flight{app="go-fiber-api"}
```

## Runtime

Goroutines:

```promql
go_goroutines{app="go-fiber-api"}
```

Resident memory (bytes):

```promql
process_resident_memory_bytes{app="go-fiber-api"}
```

Heap alloc (bytes):

```promql
go_memstats_heap_alloc_bytes{app="go-fiber-api"}
```
