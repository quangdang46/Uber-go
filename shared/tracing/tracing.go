package tracing

import (
	"context"
	"fmt"

	"log"

	"go.opentelemetry.io/otel"

	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

type Config struct {
	ServiceName    string
	Environment    string
	JaegerEndpoint string
}

func InitTracer(config Config) (func(context.Context) error, error) {

	traceExporter, err := newJaegerExporter(config.JaegerEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create jaeger exporter: %w", err)
	}

	traceProvider, err := newTraceProvider(config, traceExporter)
	if err != nil {
		return nil, fmt.Errorf("failed to create tracer provider: %w", err)
	}

	otel.SetTracerProvider(traceProvider)

	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	return traceProvider.Shutdown, nil

}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTraceProvider(config Config, exporter sdktrace.SpanExporter) (*sdktrace.TracerProvider, error) {
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(config.ServiceName),
			semconv.DeploymentEnvironmentKey.String(config.Environment),
		))
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create tracer provider: %w", err)
	}
	return traceProvider, nil

}

func newJaegerExporter(endpoint string) (sdktrace.SpanExporter, error) {
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(endpoint)))
	if err != nil {
		log.Fatalf("failed to create jaeger exporter: %v", err)
	}
	return exporter, nil
}

func GetTracer(name string) trace.Tracer {

	return otel.GetTracerProvider().Tracer(name)
}
