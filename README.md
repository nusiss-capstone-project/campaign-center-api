# campaign-center-api

Backend API for campaign-center.

## Modules

- `common`: shared protobuf and middleware helpers
- `client`: gRPC client bootstrap
- `server`: gin/http, MySQL, Redis, gRPC, and OpenTelemetry bootstrap

## Observability

The server always writes structured JSON logs locally and in Railway. OpenTelemetry
export is optional: when OTLP settings are missing, the app logs the missing
variables and starts normally. If the endpoint itself is absent, it logs
`OTLP endpoint not configured, telemetry export disabled`.

### Local Development

Run without Grafana Cloud export:

```bash
cd server
go run .
```

Example access log:

```json
{"level":"info","time":"2026-05-16T15:40:00+08:00","msg":"http request completed","service":"campaign-center-api","env":"local","request_id":"c1fd0f6e5dbf45ad9fd03d4e8ed12f65","method":"GET","path":"/campaign-center-api/v1/ping","route":"/campaign-center-api/v1/ping","status":200,"duration_ms":2.31,"trace_id":"4bf92f3577b34da6a3ce929d0e0e4736","span_id":"00f067aa0ba902b7"}
```

Every HTTP response includes `X-Request-ID`. If the request already has this
header, the server propagates it; otherwise it generates one.

### Railway + Grafana Cloud OTLP

Configure these Railway variables to enable direct OTLP HTTP/protobuf export to
Grafana Cloud:

```bash
APP_ENV=production
OTEL_SERVICE_NAME=campaign-center-api
OTEL_EXPORTER_OTLP_ENDPOINT=https://otlp-gateway-<region>.grafana.net/otlp
OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf
OTEL_EXPORTER_OTLP_HEADERS="Authorization=Basic <base64-instance-id-api-token>"
OTEL_RESOURCE_ATTRIBUTES=deployment.environment=railway,service.namespace=campaign-center
OTEL_TRACES_SAMPLER_ARG=1.0
```

Notes:

- `OTEL_EXPORTER_OTLP_ENDPOINT`, `OTEL_EXPORTER_OTLP_HEADERS`, and
  `OTEL_EXPORTER_OTLP_PROTOCOL` must all be present before export is enabled.
- Only `http/protobuf` is supported by this lightweight setup.
- Trace sampling is parent-based. `OTEL_TRACES_SAMPLER_ARG` controls sampling
  for new root traces (`1.0` = 100%, `0.1` = 10%) while respecting upstream
  sampled/unsampled `traceparent` decisions.
- Go runtime metrics are collected every 15 seconds when metrics export is
  enabled, using the same OTLP MeterProvider as HTTP metrics.
- Do not configure a local OpenTelemetry Collector for Railway; the app exports
  directly to Grafana Cloud OTLP.

### Verify in Grafana Cloud

Traces:

1. Open Grafana Cloud Tempo.
2. Search for service name `campaign-center-api`.
3. Trigger an API call, for example `GET /campaign-center-api/v1/ping`.
4. Confirm spans include HTTP method, route, status, and duration.

Metrics:

1. Open Grafana Cloud metrics / dashboards.
2. Filter by `service.name="campaign-center-api"` or the configured
   `OTEL_SERVICE_NAME`.
3. Confirm HTTP server metrics appear after traffic reaches the Railway service.

For this deployment, Grafana Cloud Prometheus exposes the OTel Go runtime
metrics with Prometheus-style names. Search these exact metric identifiers in
Grafana Explore:

- `go_goroutine_count`
- `go_memory_used_bytes`
- `go_memory_limit_bytes`
- `go_memory_allocated_bytes_total`
- `go_memory_allocations_total`
- `go_memory_gc_goal_bytes`
- `go_processor_limit`
- `go_config_gogc_percent`

If `OTEL_GO_X_DEPRECATED_RUNTIME_METRICS=true` is set, the runtime package also
emits deprecated names such as `runtime_go_goroutines` and
`runtime_go_mem_heap_alloc_bytes`.

Logs:

1. Open Railway logs.
2. Filter for `request_id`, `trace_id`, or `http request completed`.
3. Use the `trace_id` from a log line to correlate with Tempo traces.
