package trace

import (
	"context"

	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

type SeverityTracerProvider struct {
	provider trace.TracerProvider
}

func (p *SeverityTracerProvider) TracerProvider() trace.TracerProvider {
	return p.provider
}

func (p *SeverityTracerProvider) Shutdown(ctx context.Context) error {
	switch v := p.provider.(type) {
	case *tracesdk.TracerProvider:
		return v.Shutdown(ctx)
	}
	return nil
}

func (p *SeverityTracerProvider) Tracer(name string, opts ...trace.TracerOption) *SeverityTracer {
	tr := p.provider.Tracer(name, opts...)
	return CreateSeverityTracer(tr)
}

func JaegerProvider(url string, attrs ...KeyValue) (*SeverityTracerProvider, error) {
	// Create the Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			attrs...,
		)),
	)

	stp := CreateSeverityTracerProvider(tp)
	return stp, nil
}
