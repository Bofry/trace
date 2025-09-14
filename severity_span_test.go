package trace

import (
	"context"
	"errors"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestSeveritySpan_AllSeverityLevels(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	tracer := Tracer("test-tracer")
	span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	// Test all severity levels
	testCases := []struct {
		name     string
		severity Severity
		method   func(string, ...any) SpanEvent
	}{
		{"Debug", DEBUG, span.Debug},
		{"Info", INFO, span.Info},
		{"Notice", NOTICE, span.Notice},
		{"Warning", WARN, span.Warning},
		{"Crit", CRIT, span.Crit},
		{"Alert", ALERT, span.Alert},
		{"Emerg", EMERG, span.Emerg},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := tc.method("Test %s message", tc.name)
			if event == nil {
				t.Errorf("Expected non-nil event for %s", tc.name)
			}
			if !event.IsRecording() {
				t.Errorf("Expected event to be recording for %s", tc.name)
			}
		})
	}
}

func TestSeveritySpan_Argv(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	tracer := Tracer("test-tracer")
	span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	testCases := []struct {
		name  string
		input any
	}{
		{"String", "test-string"},
		{"Int", 42},
		{"Float", 3.14},
		{"Bool", true},
		{"Map", map[string]any{"key": "value", "num": 123}},
		{"Nil", nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			span.Argv(tc.input)
		})
	}
}

func TestSeveritySpan_Reply(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	tracer := Tracer("test-tracer")
	span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	testCases := []struct {
		name string
		code ReplyCode
		data any
	}{
		{"Pass", PASS, "success"},
		{"Fail", FAIL, "error occurred"},
		{"Unset", UNSET, nil},
		{"PassWithData", PASS, map[string]any{"result": "ok", "count": 10}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			span.Reply(tc.code, tc.data)
		})
	}
}

func TestSeveritySpan_Err(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	tracer := Tracer("test-tracer")
	span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	testErr := errors.New("test error")
	span.Err(testErr)

	// The error should be recorded internally
	if span.err != testErr {
		t.Errorf("Expected span.err to be %v, got %v", testErr, span.err)
	}
}

func TestSeveritySpan_Tags(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	tracer := Tracer("test-tracer")
	span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	tags := []KeyValue{
		Key("string.tag").String("value"),
		Key("int.tag").Int(42),
		Key("bool.tag").Bool(true),
	}

	span.Tags(tags...)
}

func TestSeveritySpan_Disable(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	tracer := Tracer("test-tracer")
	span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	// Initially not disabled
	if span.IsDisabled() {
		t.Error("Expected span to not be disabled initially")
	}

	// Disable span
	span.Disable(true)
	if !span.IsDisabled() {
		t.Error("Expected span to be disabled after Disable(true)")
	}

	// Re-enable span
	span.Disable(false)
	if span.IsDisabled() {
		t.Error("Expected span to not be disabled after Disable(false)")
	}
}

func TestSeveritySpan_ContextAndSpanInfo(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	tracer := Tracer("test-tracer")
	ctx := context.Background()
	span := tracer.Start(ctx, "test-span")
	defer span.End()

	// Test context
	spanCtx := span.Context()
	if spanCtx == nil {
		t.Error("Expected non-nil context")
	}

	// Test span info methods
	if !span.HasSpanID() {
		t.Error("Expected span to have SpanID")
	}

	if !span.HasTraceID() {
		t.Error("Expected span to have TraceID")
	}

	spanID := span.SpanID()
	if !spanID.IsValid() {
		t.Error("Expected valid SpanID")
	}

	traceID := span.TraceID()
	if !traceID.IsValid() {
		t.Error("Expected valid TraceID")
	}

	traceFlags := span.TraceFlags()
	_ = traceFlags // Just ensure it doesn't panic

	traceState := span.TraceState()
	_ = traceState // Just ensure it doesn't panic
}

func TestSeveritySpan_Link(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	tracer := Tracer("test-tracer")
	span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	link := span.Link()
	if !link.SpanContext.IsValid() {
		t.Error("Expected valid link SpanContext")
	}
}

func TestSeveritySpan_DisabledBehavior(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	tracer := Tracer("test-tracer")
	span := tracer.Start(context.Background(), "test-span")

	// Disable the span
	span.Disable(true)

	// End should be no-op when disabled
	span.End()

	// Span should still be disabled after End
	if !span.IsDisabled() {
		t.Error("Expected span to remain disabled after End")
	}
}

func TestSeveritySpan_EventChaining(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	tracer := Tracer("test-tracer")
	span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	// Create event and chain operations
	event := span.Debug("Test message")
	event.Tags(Key("test.key").String("test.value"))
	event.Vars(map[string]any{"var1": "value1", "var2": 42})
	event.Error(errors.New("test error"))

	if !event.IsRecording() {
		t.Error("Expected event to be recording")
	}

	// Flush the event
	event.Flush()

	// After flush, should not be recording
	if event.IsRecording() {
		t.Error("Expected event to not be recording after flush")
	}
}