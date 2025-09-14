package trace

import (
	"context"
	"fmt"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func BenchmarkSeveritySpan_Debug(b *testing.B) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	tracer := Tracer("benchmark-tracer")
	span := tracer.Start(context.Background(), "benchmark-span")
	defer span.End()

	b.ResetTimer()
	b.ReportAllocs()

	i := 0
	for b.Loop() {
		event := span.Debug("Debug message %d", i)
		event.Tags(Key("iteration").Int(i))
		i++
	}
}

func BenchmarkSeveritySpan_Info(b *testing.B) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	tracer := Tracer("benchmark-tracer")
	span := tracer.Start(context.Background(), "benchmark-span")
	defer span.End()

	b.ResetTimer()
	b.ReportAllocs()

	i := 0
	for b.Loop() {
		event := span.Info("Info message %d", i)
		event.Tags(Key("iteration").Int(i))
		i++
	}
}

func BenchmarkSeveritySpan_Warning(b *testing.B) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	tracer := Tracer("benchmark-tracer")
	span := tracer.Start(context.Background(), "benchmark-span")
	defer span.End()

	b.ResetTimer()
	b.ReportAllocs()

	i := 0
	for b.Loop() {
		event := span.Warning("Warning message %d", i)
		event.Tags(Key("iteration").Int(i))
		i++
	}
}

func BenchmarkSeveritySpan_NoopSpan(b *testing.B) {
	// Use background context to get noop span
	span := CreateSeveritySpan(context.Background())

	b.ResetTimer()
	b.ReportAllocs()

	i := 0
	for b.Loop() {
		event := span.Debug("Debug message %d", i)
		event.Tags(Key("iteration").Int(i))
		i++
	}
}

func BenchmarkExpandObject_String(b *testing.B) {
	testString := "test string value"

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		result := expandObject("test", testString)
		_ = result
	}
}

func BenchmarkExpandObject_Map(b *testing.B) {
	testMap := map[string]any{
		"string_key": "string_value",
		"int_key":    42,
		"float_key":  3.14,
		"bool_key":   true,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		result := expandObject("test", testMap)
		_ = result
	}
}

func BenchmarkExpandObject_LargeMap(b *testing.B) {
	// Create a larger map to test performance
	testMap := make(map[string]any)
	for i := 0; i < 20; i++ {
		testMap[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("value_%d", i)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		result := expandObject("test", testMap)
		_ = result
	}
}

func BenchmarkExpandObject_Slice(b *testing.B) {
	testSlice := []string{"item1", "item2", "item3", "item4", "item5"}

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		result := expandObject("test", testSlice)
		_ = result
	}
}

func BenchmarkTracerTagBuilder(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		builder := &TracerTagBuilder{
			namespace: "benchmark",
			container: make([]KeyValue, 0, 8),
		}

		builder.String("string_field", "test_value")
		builder.Int("int_field", 42)
		builder.Float64("float_field", 3.14)
		builder.Bool("bool_field", true)

		result := builder.Result()
		_ = result
	}
}

func BenchmarkTracerTagBuilder_Large(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		builder := &TracerTagBuilder{
			namespace: "benchmark",
			container: make([]KeyValue, 0, 32),
		}

		// Add many fields to test scaling
		for j := 0; j < 10; j++ {
			builder.String(fmt.Sprintf("string_%d", j), fmt.Sprintf("value_%d", j))
			builder.Int(fmt.Sprintf("int_%d", j), j)
		}

		result := builder.Result()
		_ = result
	}
}

func BenchmarkSpanFromContext(b *testing.B) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	tracer := Tracer("benchmark-tracer")
	testSpan := tracer.Start(context.Background(), "test-span")
	defer testSpan.End()

	ctx := ContextWithSpan(context.Background(), testSpan)

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		span := SpanFromContext(ctx)
		_ = span
	}
}

func BenchmarkSpanFromContext_WithExtractor(b *testing.B) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	tracer := Tracer("benchmark-tracer")
	testSpan := tracer.Start(context.Background(), "test-span")
	defer testSpan.End()

	ctx := context.Background()
	extractor := &testSpanExtractor{span: testSpan}

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		span := SpanFromContext(ctx, extractor)
		_ = span
	}
}

func BenchmarkGlobalState_GetTracerProvider(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		provider := GetTracerProvider()
		_ = provider
	}
}

func BenchmarkGlobalState_GetSpanExtractor(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		extractor := GetSpanExtractor()
		_ = extractor
	}
}

func BenchmarkSpanEvent_Creation(b *testing.B) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	tracer := Tracer("benchmark-tracer")
	span := tracer.Start(context.Background(), "benchmark-span")
	defer span.End()

	b.ResetTimer()
	b.ReportAllocs()

	i := 0
	for b.Loop() {
		event := span.Debug("Benchmark message %d", i)
		_ = event
		i++
	}
}

func BenchmarkSpanEvent_WithTags(b *testing.B) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	tracer := Tracer("benchmark-tracer")
	span := tracer.Start(context.Background(), "benchmark-span")
	defer span.End()

	tags := []KeyValue{
		Key("benchmark.iteration").Int(0),
		Key("benchmark.type").String("performance"),
		Key("benchmark.enabled").Bool(true),
	}

	b.ResetTimer()
	b.ReportAllocs()

	i := 0
	for b.Loop() {
		event := span.Debug("Benchmark message %d", i)
		event.Tags(tags...)
		i++
	}
}

func BenchmarkSpanEvent_WithVars(b *testing.B) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	tracer := Tracer("benchmark-tracer")
	span := tracer.Start(context.Background(), "benchmark-span")
	defer span.End()

	vars := map[string]any{
		"iteration": 0,
		"type":      "benchmark",
		"enabled":   true,
		"rate":      1.5,
	}

	b.ResetTimer()
	b.ReportAllocs()

	i := 0
	for b.Loop() {
		event := span.Debug("Benchmark message %d", i)
		event.Vars(vars)
		i++
	}
}

func BenchmarkSpanEvent_FullWorkflow(b *testing.B) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	tracer := Tracer("benchmark-tracer")
	span := tracer.Start(context.Background(), "benchmark-span")
	defer span.End()

	b.ResetTimer()
	b.ReportAllocs()

	i := 0
	for b.Loop() {
		event := span.Debug("Benchmark message %d", i)
		event.Tags(Key("iteration").Int(i))
		event.Vars(map[string]any{"data": i})
		event.Flush()
		i++
	}
}

// Benchmark comparison: NoopEvent vs SeverityEvent
func BenchmarkNoopEvent_Operations(b *testing.B) {
	nop := nopEventInstance
	tags := []KeyValue{Key("test").String("value")}
	vars := map[string]any{"test": "value"}

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		nop.Tags(tags...)
		nop.Vars(vars)
		nop.Flush()
	}
}

