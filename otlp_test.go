package trace

import (
	"context"
	"testing"
)

func TestOTLPProviderCreation(t *testing.T) {
	// Test that OTLP HTTP provider can be created (without network connection)
	tp, err := OTLPProvider("http://localhost:4318",
		ServiceName("otlp-test"),
		Environment("go-test"),
	)
	if err != nil {
		t.Fatalf("Failed to create OTLP provider: %v", err)
	}

	if tp == nil {
		t.Fatal("OTLP provider should not be nil")
	}

	// Test basic functionality
	tracer := tp.Tracer("test-tracer")
	if tracer == nil {
		t.Fatal("Tracer should not be nil")
	}

	// Cleanup
	tp.Shutdown(context.Background())
}

func TestOTLPGRPCProviderCreation(t *testing.T) {
	// Test that OTLP gRPC provider can be created (without network connection)
	tp, err := OTLPGRPCProvider("localhost:4317",
		ServiceName("otlp-grpc-test"),
		Environment("go-test"),
	)
	if err != nil {
		t.Fatalf("Failed to create OTLP gRPC provider: %v", err)
	}

	if tp == nil {
		t.Fatal("OTLP gRPC provider should not be nil")
	}

	// Test basic functionality
	tracer := tp.Tracer("test-tracer")
	if tracer == nil {
		t.Fatal("Tracer should not be nil")
	}

	// Cleanup
	tp.Shutdown(context.Background())
}

func TestJaegerProviderBackwardCompatibility(t *testing.T) {
	// Test that JaegerProvider still works and creates OTLP provider
	tp, err := JaegerProvider("http://localhost:14268/api/traces",
		ServiceName("jaeger-compat-test"),
		Environment("go-test"),
	)
	if err != nil {
		t.Fatalf("Failed to create Jaeger-compatible provider: %v", err)
	}

	if tp == nil {
		t.Fatal("Jaeger-compatible provider should not be nil")
	}

	// Test basic functionality
	tracer := tp.Tracer("test-tracer")
	if tracer == nil {
		t.Fatal("Tracer should not be nil")
	}

	// Cleanup
	tp.Shutdown(context.Background())
}

func TestConvertJaegerToOTLP(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Jaeger HTTP collector",
			input:    "http://localhost:14268/api/traces",
			expected: "http://localhost:4318",
		},
		{
			name:     "Jaeger gRPC collector",
			input:    "http://localhost:14250",
			expected: "http://localhost:4317",
		},
		{
			name:     "Remote Jaeger HTTP",
			input:    "http://jaeger.example.com:14268/api/traces",
			expected: "http://jaeger.example.com:4318",
		},
		{
			name:     "HTTPS Jaeger",
			input:    "https://secure.jaeger.com:14268/api/traces",
			expected: "https://secure.jaeger.com:4318",
		},
		{
			name:     "No port specified",
			input:    "http://localhost",
			expected: "http://localhost:4318",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := convertJaegerToOTLP(tc.input)
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestProviderCreationWithInvalidEndpoint(t *testing.T) {
	// Test with invalid endpoint
	_, err := OTLPProvider(":::invalid-url",
		ServiceName("invalid-test"),
	)
	if err == nil {
		t.Error("Expected error with invalid endpoint, got nil")
	}
}
