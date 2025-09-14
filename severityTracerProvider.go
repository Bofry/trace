package trace

import (
	"context"
	"net/url"
	"strings"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
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

// OTLPProvider creates a provider using OTLP HTTP exporter
func OTLPProvider(endpoint string, attrs ...KeyValue) (*SeverityTracerProvider, error) {
	// Parse URL to extract host:port and determine if secure
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	hostPort := parsed.Host
	if parsed.Port() == "" {
		// Add default port if not specified
		if parsed.Scheme == "https" {
			hostPort = hostPort + ":4318"
		} else {
			hostPort = hostPort + ":4318"
		}
	}

	ctx := context.Background()
	options := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(hostPort),
		otlptracehttp.WithURLPath("/v1/traces"),
	}

	// Set insecure if using HTTP
	if parsed.Scheme == "http" {
		options = append(options, otlptracehttp.WithInsecure())
	}

	exp, err := otlptracehttp.New(ctx, options...)
	if err != nil {
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			attrs...,
		)),
	)

	stp := CreateSeverityTracerProvider(tp)
	return stp, nil
}

// OTLPGRPCProvider creates a provider using OTLP gRPC exporter
func OTLPGRPCProvider(endpoint string, attrs ...KeyValue) (*SeverityTracerProvider, error) {
	ctx := context.Background()
	exp, err := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint(endpoint))
	if err != nil {
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			attrs...,
		)),
	)

	stp := CreateSeverityTracerProvider(tp)
	return stp, nil
}

// JaegerCompatibleProvider creates a provider that sends traces to Jaeger using OTLP
// This is a convenience function for testing with local Jaeger instances.
//
// For Jaeger v1.35+, it uses the OTLP endpoint (port 4318 for HTTP).
// For older Jaeger versions, you may need to use the legacy Jaeger collector ports.
//
// Usage:
//   tp, err := trace.JaegerCompatibleProvider("http://localhost:4318", attrs...)
//   // Or for Jaeger running on legacy port: "http://localhost:14268" (will auto-convert to OTLP)
func JaegerCompatibleProvider(endpoint string, attrs ...KeyValue) (*SeverityTracerProvider, error) {
	// Auto-detect and convert legacy Jaeger endpoints to OTLP
	if strings.Contains(endpoint, ":14268") {
		// Convert Jaeger HTTP collector (14268) to OTLP HTTP (4318)
		endpoint = strings.Replace(endpoint, ":14268", ":4318", 1)
		// Remove Jaeger-specific path
		endpoint = strings.Replace(endpoint, "/api/traces", "", 1)
	} else if strings.Contains(endpoint, ":14250") {
		// Convert Jaeger gRPC collector (14250) to OTLP gRPC (4317)
		endpoint = strings.Replace(endpoint, ":14250", ":4317", 1)
		return OTLPGRPCProvider(endpoint, attrs...)
	}

	// Use OTLP HTTP by default
	return OTLPProvider(endpoint, attrs...)
}


