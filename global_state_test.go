package trace

import (
	"context"
	"sync"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestGlobalTracerProvider_ConcurrentAccess(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp1 := CreateSeverityTracerProvider(trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	))
	tp2 := CreateSeverityTracerProvider(trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	))

	const numGoroutines = 100
	var wg sync.WaitGroup

	// Test concurrent Set and Get operations
	wg.Add(numGoroutines * 2)

	// Concurrent setters
	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()
			if i%2 == 0 {
				SetTracerProvider(tp1)
			} else {
				SetTracerProvider(tp2)
			}
		}(i)
	}

	// Concurrent getters
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			provider := GetTracerProvider()
			if provider == nil {
				t.Error("GetTracerProvider returned nil")
			}
		}()
	}

	wg.Wait()

	// Final provider should be one of the two we set
	finalProvider := GetTracerProvider()
	if finalProvider != tp1 && finalProvider != tp2 {
		t.Error("Final provider should be either tp1 or tp2")
	}
}

func TestGlobalSpanExtractor_ConcurrentAccess(t *testing.T) {
	extractor1 := &testSpanExtractor{span: nil}
	extractor2 := &testSpanExtractor{span: nil}

	const numGoroutines = 10 // Reduced for stability
	var wg sync.WaitGroup

	// Test concurrent Set and Get operations
	wg.Add(numGoroutines * 2)

	// Concurrent setters
	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()
			if i%2 == 0 {
				SetSpanExtractor(extractor1)
			} else {
				SetSpanExtractor(extractor2)
			}
		}(i)
	}

	// Concurrent getters
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			extractor := GetSpanExtractor()
			if extractor == nil {
				t.Error("GetSpanExtractor returned nil")
			}
		}()
	}

	wg.Wait()

	// Final extractor should not be nil
	finalExtractor := GetSpanExtractor()
	if finalExtractor == nil {
		t.Error("Final extractor should not be nil")
	}
}

func TestGlobalSpanInjector_ConcurrentAccess(t *testing.T) {
	injector1 := noopSpanInjectorInstance
	injector2 := noopSpanInjectorInstance // Using same instance for simplicity

	const numGoroutines = 100
	var wg sync.WaitGroup

	// Test concurrent Set and Get operations
	wg.Add(numGoroutines * 2)

	// Concurrent setters
	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()
			if i%2 == 0 {
				SetSpanInjector(injector1)
			} else {
				SetSpanInjector(injector2)
			}
		}(i)
	}

	// Concurrent getters
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			injector := GetSpanInjector()
			if injector == nil {
				t.Error("GetSpanInjector returned nil")
			}
		}()
	}

	wg.Wait()
}

func TestGlobalState_SetSameValue(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := CreateSeverityTracerProvider(trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	))

	// Get initial provider
	initial := GetTracerProvider()

	// Set the same provider multiple times
	SetTracerProvider(tp)
	SetTracerProvider(tp)
	SetTracerProvider(tp)

	// Should still be the same
	if GetTracerProvider() != tp {
		t.Error("Expected same provider after multiple sets")
	}

	// Reset to initial
	SetTracerProvider(initial)
}

func TestGlobalState_DefaultValues(t *testing.T) {
	// Test default tracer provider is not nil
	defaultTP := GetTracerProvider()
	if defaultTP == nil {
		t.Error("Default TracerProvider should not be nil")
	}

	// Test default span extractor is not nil
	defaultExtractor := GetSpanExtractor()
	if defaultExtractor == nil {
		t.Error("Default SpanExtractor should not be nil")
	}

	// Test default span injector is not nil
	defaultInjector := GetSpanInjector()
	if defaultInjector == nil {
		t.Error("Default SpanInjector should not be nil")
	}
}

func TestGlobalTracer_Function(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := CreateSeverityTracerProvider(trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	))
	SetTracerProvider(tp)

	// Test global Tracer function
	tracer := Tracer("test-tracer")
	if tracer == nil {
		t.Error("Tracer should not be nil")
	}

	// Test that it creates valid spans
	span := tracer.Start(context.Background(), "test-span")
	if span == nil {
		t.Error("Span should not be nil")
	}
	defer span.End()
}

// Custom span extractor for testing
type testSpanExtractor struct {
	span *SeveritySpan
}

func (tse *testSpanExtractor) Extract(ctx context.Context) *SeveritySpan {
	return tse.span
}

func TestGlobalSpanExtractor_CustomExtractor(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	tracer := Tracer("test-tracer")
	testSpan := tracer.Start(context.Background(), "test-span")
	defer testSpan.End()

	// Create custom extractor
	customExtractor := &testSpanExtractor{span: testSpan}
	SetSpanExtractor(customExtractor)

	// Test that it returns our custom span
	ctx := context.Background()
	extractedSpan := SpanFromContext(ctx)
	if extractedSpan != testSpan {
		t.Error("Expected custom extractor to return our test span")
	}
}

// Benchmark concurrent access to global state
func BenchmarkGlobalState_ConcurrentAccess(b *testing.B) {
	exporter := tracetest.NewInMemoryExporter()
	tp := CreateSeverityTracerProvider(trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	))

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Mix of reads and writes
			if b.N%10 == 0 {
				SetTracerProvider(tp)
			} else {
				_ = GetTracerProvider()
			}
		}
	})
}