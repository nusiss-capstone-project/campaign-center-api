package telemetry

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/lianjin/campaign-center-api/server/http/data"
	appLog "github.com/lianjin/campaign-center-api/server/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

const disabledMessage = "OTLP endpoint not configured, telemetry export disabled"

type shutdownFunc func(context.Context) error

// Init configures OpenTelemetry from OTEL_* environment variables.
// Export is intentionally optional so local/dev startup is never blocked by telemetry.
func Init(ctx context.Context) shutdownFunc {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{}))

	if !exportEnabled() {
		appLog.Logger.Infow(disabledMessage,
			"otel_exporter_otlp_endpoint_set", os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "",
			"otel_exporter_otlp_headers_set", os.Getenv("OTEL_EXPORTER_OTLP_HEADERS") != "",
			"otel_exporter_otlp_protocol", os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL"),
		)
		return func(context.Context) error { return nil }
	}

	if protocol := strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL")); protocol != "http/protobuf" {
		appLog.Logger.Warnw("unsupported OTLP protocol, telemetry export disabled",
			"otel_exporter_otlp_protocol", protocol,
			"supported_protocol", "http/protobuf",
		)
		return func(context.Context) error { return nil }
	}

	res, err := newResource(ctx)
	if err != nil {
		appLog.Logger.Errorw("failed to create OpenTelemetry resource, telemetry export disabled", "error", err)
		return func(context.Context) error { return nil }
	}
	tp, err := initTracer(ctx, res)
	if err != nil {
		appLog.Logger.Errorw("failed to initialize trace exporter, telemetry export disabled", "error", err)
		return func(context.Context) error { return nil }
	}
	mp, err := initMetrics(ctx, res)
	if err != nil {
		appLog.Logger.Errorw("failed to initialize metric exporter, telemetry metrics disabled", "error", err)
	}
	appLog.Logger.Infow("OpenTelemetry export initialized",
		"otel_exporter_otlp_endpoint", os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		"otel_exporter_otlp_protocol", os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL"),
	)

	return func(ctx context.Context) error {
		var shutdownErr error
		if mp != nil {
			shutdownErr = mp.Shutdown(ctx)
		}
		if err := tp.Shutdown(ctx); err != nil && shutdownErr == nil {
			shutdownErr = err
		}
		return shutdownErr
	}
}

func initTracer(ctx context.Context, res *resource.Resource) (*sdktrace.TracerProvider, error) {
	exporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	return tp, nil
}

func initMetrics(ctx context.Context, res *resource.Resource) (*sdkmetric.MeterProvider, error) {
	exp, err := otlpmetrichttp.New(ctx)
	if err != nil {
		return nil, err
	}
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exp, sdkmetric.WithInterval(15*time.Second))),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(mp)
	return mp, nil
}

func newResource(ctx context.Context) (*resource.Resource, error) {
	attrs := []attribute.KeyValue{
		semconv.ServiceName(serviceName()),
		semconv.ServiceVersion("1.0.0"),
	}
	attrs = append(attrs, parseResourceAttributes(os.Getenv("OTEL_RESOURCE_ATTRIBUTES"))...)
	return resource.New(ctx, resource.WithAttributes(attrs...))
}

func serviceName() string {
	if v := strings.TrimSpace(os.Getenv("OTEL_SERVICE_NAME")); v != "" {
		return v
	}
	return data.ServiceName
}

func exportEnabled() bool {
	return strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")) != "" &&
		strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_HEADERS")) != "" &&
		strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL")) != ""
}

func parseResourceAttributes(raw string) []attribute.KeyValue {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	attrs := make([]attribute.KeyValue, 0, len(parts))
	for _, part := range parts {
		key, value, ok := strings.Cut(part, "=")
		key = strings.TrimSpace(key)
		if !ok || key == "" {
			continue
		}
		attrs = append(attrs, attribute.String(key, strings.TrimSpace(value)))
	}
	return attrs
}
