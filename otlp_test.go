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



func TestProviderCreationWithInvalidEndpoint(t *testing.T) {
	// Test with invalid endpoint
	_, err := OTLPProvider(":::invalid-url",
		ServiceName("invalid-test"),
	)
	if err == nil {
		t.Error("Expected error with invalid endpoint, got nil")
	}
}

func TestJaegerCompatibleProvider(t *testing.T) {
	// Test modern OTLP endpoint
	tp1, err := JaegerCompatibleProvider("http://localhost:4318",
		ServiceName("jaeger-compatible-test-otlp"),
		Environment("go-test"),
	)
	if err != nil {
		t.Fatalf("Failed to create Jaeger-compatible provider with OTLP endpoint: %v", err)
	}
	if tp1 == nil {
		t.Fatal("Jaeger-compatible provider should not be nil")
	}
	tp1.Shutdown(context.Background())

	// Test legacy Jaeger HTTP collector endpoint (should auto-convert)
	tp2, err := JaegerCompatibleProvider("http://localhost:14268/api/traces",
		ServiceName("jaeger-compatible-test-legacy"),
		Environment("go-test"),
	)
	if err != nil {
		t.Fatalf("Failed to create Jaeger-compatible provider with legacy endpoint: %v", err)
	}
	if tp2 == nil {
		t.Fatal("Jaeger-compatible provider should not be nil")
	}
	tp2.Shutdown(context.Background())

	// Test legacy Jaeger gRPC collector endpoint (should auto-convert)
	tp3, err := JaegerCompatibleProvider("localhost:14250",
		ServiceName("jaeger-compatible-test-grpc"),
		Environment("go-test"),
	)
	if err != nil {
		t.Fatalf("Failed to create Jaeger-compatible provider with gRPC endpoint: %v", err)
	}
	if tp3 == nil {
		t.Fatal("Jaeger-compatible provider should not be nil")
	}
	tp3.Shutdown(context.Background())
}
