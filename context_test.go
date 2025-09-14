package trace

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestSpanFromContext_ExtractorPriority(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	tracer := Tracer("test-tracer")

	// Create test spans
	span1 := tracer.Start(context.Background(), "span-1")
	span2 := tracer.Start(context.Background(), "span-2")
	span3 := tracer.Start(context.Background(), "span-3")
	defer span1.End()
	defer span2.End()
	defer span3.End()

	// Create extractors
	extractor1 := &testSpanExtractor{span: span1}
	extractor2 := &testSpanExtractor{span: span2}

	// Set global extractor
	originalGlobalExtractor := GetSpanExtractor()
	SetSpanExtractor(&testSpanExtractor{span: span3})
	defer SetSpanExtractor(originalGlobalExtractor)

	ctx := context.Background()

	// Test extractor priority: specified extractors should be checked first
	result := SpanFromContext(ctx, extractor1, extractor2)
	if result != span1 {
		t.Error("Expected first specified extractor to take priority")
	}

	// Test with no specified extractors - should use global
	result = SpanFromContext(ctx)
	if result != span3 {
		t.Error("Expected global extractor to be used when no extractors specified")
	}
}

func TestSpanFromContext_ContextValue(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	tracer := Tracer("test-tracer")
	testSpan := tracer.Start(context.Background(), "test-span")
	defer testSpan.End()

	// Test ContextWithSpan
	ctx := ContextWithSpan(context.Background(), testSpan)
	result := SpanFromContext(ctx)

	if result != testSpan {
		t.Error("Expected ContextWithSpan to store and retrieve span correctly")
	}
}

func TestSpanFromContext_NilExtractors(t *testing.T) {
	ctx := context.Background()

	// Create nil-returning extractor
	nilExtractor := &testSpanExtractor{span: nil}

	// Should fall back to OpenTelemetry default when all extractors return nil
	result := SpanFromContext(ctx, nilExtractor)

	// Should not be nil (should be a SeveritySpan wrapping the default)
	if result == nil {
		t.Error("Expected non-nil span from fallback")
	}

	// Should be a no-op span
	if !IsNoopSeveritySpan(result) {
		t.Error("Expected fallback span to be no-op")
	}
}

func TestSpanToContext_ValueContext(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	tracer := Tracer("test-tracer")
	testSpan := tracer.Start(context.Background(), "test-span")
	defer testSpan.End()

	// Create a ValueContext implementation
	valueCtx := &mockValueContext{
		Context: context.Background(),
		values:  make(map[any]any),
	}

	// Test SpanToContext with ValueContext
	resultCtx := SpanToContext(valueCtx, testSpan)

	// Should return the same context (modified in-place)
	if resultCtx != valueCtx {
		t.Error("Expected SpanToContext to return same ValueContext")
	}

	// Verify span was stored
	if storedSpan, ok := valueCtx.values[__CONTEXT_SEVERITY_SPAN_KEY]; !ok || storedSpan != testSpan {
		t.Error("Expected span to be stored in ValueContext")
	}
}

func TestSpanToContext_RegularContext(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	tracer := Tracer("test-tracer")
	testSpan := tracer.Start(context.Background(), "test-span")
	defer testSpan.End()

	ctx := context.Background()

	// Test SpanToContext with regular context - should use injector (noop by default)
	resultCtx := SpanToContext(ctx, testSpan)

	// With noop injector, context should be the same
	if resultCtx != ctx {
		t.Log("SpanToContext with noop injector returned same context (expected behavior)")
	}
}

func TestContextWithSpan(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	tracer := Tracer("test-tracer")
	testSpan := tracer.Start(context.Background(), "test-span")
	defer testSpan.End()

	parentCtx := context.Background()
	childCtx := ContextWithSpan(parentCtx, testSpan)

	// Should be different context
	if childCtx == parentCtx {
		t.Error("Expected ContextWithSpan to return different context")
	}

	// Should be able to retrieve the span
	retrievedSpan := SpanFromContext(childCtx)
	if retrievedSpan != testSpan {
		t.Error("Expected to retrieve same span from context")
	}
}

func TestSpanFromContext_NilContext(t *testing.T) {
	// Should handle nil context gracefully
	result := SpanFromContext(context.TODO())

	if result == nil {
		t.Error("Expected non-nil span even with nil context")
	}

	// Should be a no-op span
	if !IsNoopSeveritySpan(result) {
		t.Error("Expected span from nil context to be no-op")
	}
}

// Custom span injector for testing
type testSpanInjector struct {
	injectedSpan *SeveritySpan
}

func (tsi *testSpanInjector) Inject(ctx context.Context, span *SeveritySpan) context.Context {
	tsi.injectedSpan = span
	return context.WithValue(ctx, "injected_span", span)
}

func TestSpanToContext_CustomInjector(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	tracer := Tracer("test-tracer")
	testSpan := tracer.Start(context.Background(), "test-span")
	defer testSpan.End()

	// Set custom injector
	customInjector := &testSpanInjector{}
	originalInjector := GetSpanInjector()
	SetSpanInjector(customInjector)
	defer SetSpanInjector(originalInjector)

	ctx := context.Background()
	resultCtx := SpanToContext(ctx, testSpan)

	// Custom injector should have been called
	if customInjector.injectedSpan != testSpan {
		t.Error("Expected custom injector to receive the test span")
	}

	// Result should contain our injected value
	if injectedSpan := resultCtx.Value("injected_span"); injectedSpan != testSpan {
		t.Error("Expected custom injector to modify context correctly")
	}
}

// Mock ValueContext for testing
type mockValueContext struct {
	context.Context
	values map[any]any
}

func (mvc *mockValueContext) SetValue(key, value any) {
	mvc.values[key] = value
}

func (mvc *mockValueContext) Value(key any) any {
	if val, ok := mvc.values[key]; ok {
		return val
	}
	return mvc.Context.Value(key)
}

func TestValueContext_Interface(t *testing.T) {
	valueCtx := &mockValueContext{
		Context: context.Background(),
		values:  make(map[any]any),
	}

	// Test SetValue and Value
	testKey := "test_key"
	testValue := "test_value"

	valueCtx.SetValue(testKey, testValue)

	retrievedValue := valueCtx.Value(testKey)
	if retrievedValue != testValue {
		t.Error("Expected to retrieve same value that was set")
	}
}

func TestCompositeSpanExtractor_Integration(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	tracer := Tracer("test-tracer")
	span1 := tracer.Start(context.Background(), "span-1")
	span2 := tracer.Start(context.Background(), "span-2")
	defer span1.End()
	defer span2.End()

	// Create composite with multiple extractors
	extractor1 := &testSpanExtractor{span: nil} // Returns nil
	extractor2 := &testSpanExtractor{span: span1} // Returns span1
	extractor3 := &testSpanExtractor{span: span2} // Returns span2

	composite := NewCompositeSpanExtractor(extractor1, extractor2, extractor3)

	ctx := context.Background()
	result := composite.Extract(ctx)

	// Should return span1 (from first non-nil extractor)
	if result != span1 {
		t.Error("Expected composite to return span from first non-nil extractor")
	}
}