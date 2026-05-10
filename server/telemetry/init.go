package telemetry

import (
	"context"
	stdlog "log"
	"os"
	"time"

	"github.com/lianjin/campaign-center-api/server/http/data"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func InitTracer() func() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		stdlog.Printf("OTEL_EXPORTER_OTLP_ENDPOINT is empty, telemetry disabled")
		return func() {}
	}

	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint(endpoint), otlptracegrpc.WithInsecure())
	if err != nil {
		stdlog.Printf("Could not set up telemetry: %v", err)
		return func() {}
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(data.ServiceName),
		)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	stdlog.Printf("Telemetry initialized, OTLP gRPC endpoint=%s", endpoint)

	return func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer shutdownCancel()
		if err := tp.Shutdown(shutdownCtx); err != nil {
			stdlog.Printf("Error shutting down tracer provider: %v", err)
		}
	}
}

func InitMetrics() func() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		stdlog.Printf("OTEL_EXPORTER_OTLP_ENDPOINT is empty, metrics disabled")
		return func() {}
	}

	res, err := resource.New(ctx, resource.WithAttributes(semconv.ServiceName(data.ServiceName), semconv.ServiceVersion("1.0.0")))
	if err != nil {
		stdlog.Printf("Could not create metrics resource: %v", err)
		return func() {}
	}

	exp, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithEndpoint(endpoint), otlpmetricgrpc.WithInsecure())
	if err != nil {
		stdlog.Printf("Could not set up metrics exporter: %v", err)
		return func() {}
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exp, sdkmetric.WithInterval(15*time.Second))),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(mp)
	stdlog.Printf("Metrics initialized, OTLP gRPC endpoint=%s", endpoint)

	return func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer shutdownCancel()
		if err := mp.Shutdown(shutdownCtx); err != nil {
			stdlog.Printf("Error shutting down meter provider: %v", err)
		}
	}
}
