package trace

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
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

// convertJaegerToOTLP converts Jaeger endpoint to OTLP endpoint
func convertJaegerToOTLP(jaegerURL string) string {
	parsed, err := url.Parse(jaegerURL)
	if err != nil {
		// Fallback to default OTLP endpoint
		return "http://localhost:4318"
	}

	// Convert known Jaeger ports to OTLP ports
	host := parsed.Hostname()
	port := parsed.Port()

	var newPort string
	switch port {
	case "14268": // Jaeger HTTP collector
		newPort = "4318"
	case "14250": // Jaeger gRPC collector
		newPort = "4317"
	default:
		// Default to OTLP HTTP port if no specific port mapping
		newPort = "4318"
	}

	// Reconstruct URL with OTLP endpoint
	var otlpURL strings.Builder
	otlpURL.WriteString(parsed.Scheme)
	otlpURL.WriteString("://")
	otlpURL.WriteString(host)
	otlpURL.WriteString(":")
	otlpURL.WriteString(newPort)

	return otlpURL.String()
}

// JaegerProvider creates a provider compatible with Jaeger endpoints using OTLP
// Deprecated: Use OTLPProvider or OTLPGRPCProvider instead
// Note: This function now uses OTLP internally. Requires Jaeger v1.35+ with OTLP support.
func JaegerProvider(jaegerURL string, attrs ...KeyValue) (*SeverityTracerProvider, error) {
	// Convert Jaeger endpoint to OTLP endpoint
	otlpEndpoint := convertJaegerToOTLP(jaegerURL)

	ctx := context.Background()
	parsed, parseErr := url.Parse(otlpEndpoint)
	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse converted OTLP endpoint %q: %w (requires Jaeger v1.35+ with OTLP support)", otlpEndpoint, parseErr)
	}

	hostPort := parsed.Host
	if parsed.Port() == "" {
		hostPort = hostPort + ":4318"
	}

	options := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(hostPort),
		otlptracehttp.WithURLPath("/v1/traces"),
	}

	if parsed.Scheme == "http" {
		options = append(options, otlptracehttp.WithInsecure())
	}

	exp, err := otlptracehttp.New(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter for endpoint %q: %w (ensure Jaeger v1.35+ with OTLP enabled)", otlpEndpoint, err)
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

