package trace

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// Test TracerTagBuilder
func TestTracerTagBuilder_AllTypes(t *testing.T) {
	builder := &TracerTagBuilder{
		namespace: "test",
		container: make([]KeyValue, 0, 16),
	}

	// Test all type methods
	builder.String("string_field", "test_value")
	builder.StringSlice("string_slice", []string{"a", "b", "c"})
	builder.Int("int_field", 42)
	builder.IntSlice("int_slice", []int{1, 2, 3})
	builder.Int64("int64_field", int64(123456))
	builder.Int64Slice("int64_slice", []int64{4, 5, 6})
	builder.Float64("float64_field", 3.14)
	builder.Float64Slice("float64_slice", []float64{1.1, 2.2, 3.3})
	builder.Bool("bool_field", true)
	builder.BoolSlice("bool_slice", []bool{true, false, true})

	result := builder.Result()
	if len(result) != 10 {
		t.Errorf("Expected 10 key-value pairs, got %d", len(result))
	}

	// Check namespace is correctly applied
	for _, kv := range result {
		key := string(kv.Key)
		if len(key) < 5 || key[:5] != "test." {
			t.Errorf("Expected key to start with 'test.', got %s", key)
		}
	}
}

func TestTracerTagBuilder_EmptyName(t *testing.T) {
	builder := &TracerTagBuilder{
		namespace: "test",
		container: make([]KeyValue, 0, 4),
	}

	initialLen := len(builder.container)

	// Empty names should be ignored
	builder.String("", "value")
	builder.Int("", 42)
	builder.Bool("", true)

	if len(builder.container) != initialLen {
		t.Error("Empty names should not add to container")
	}
}

func TestTracerTagBuilder_Value(t *testing.T) {
	builder := &TracerTagBuilder{
		namespace: "test",
		container: make([]KeyValue, 0, 8),
	}

	testCases := []struct {
		name  string
		value any
	}{
		{"simple_string", "hello"},
		{"number", 42},
		{"map_value", map[string]any{"nested": "value", "count": 10}},
		{"slice_value", []string{"item1", "item2"}},
	}

	for _, tc := range testCases {
		builder.Value(tc.name, tc.value)
	}

	result := builder.Result()
	if len(result) == 0 {
		t.Error("Expected non-empty result from Value method")
	}
}

// Test expandObject function
func TestExpandObject_PrimitiveTypes(t *testing.T) {
	testCases := []struct {
		name     string
		input    any
		expected int // expected number of KeyValue pairs
	}{
		{"String", "test", 1},
		{"Int", 42, 1},
		{"Int64", int64(123), 1},
		{"Float64", 3.14, 1},
		{"Bool", true, 1},
		{"StringSlice", []string{"a", "b"}, 1},
		{"IntSlice", []int{1, 2, 3}, 1},
		{"Int64Slice", []int64{4, 5}, 1},
		{"Float64Slice", []float64{1.1, 2.2}, 1},
		{"BoolSlice", []bool{true, false}, 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := expandObject("test", tc.input)
			if len(result) != tc.expected {
				t.Errorf("Expected %d KeyValue pairs for %s, got %d", tc.expected, tc.name, len(result))
			}
		})
	}
}

func TestExpandObject_Map(t *testing.T) {
	testMap := map[string]any{
		"string_key": "string_value",
		"int_key":    42,
		"bool_key":   true,
		"nested_map": map[string]any{"inner": "value"},
	}

	result := expandObject("test", testMap)

	// Should have one KV pair for each map entry
	if len(result) < 4 {
		t.Errorf("Expected at least 4 KeyValue pairs for map, got %d", len(result))
	}

	// Check that keys have correct namespace
	for _, kv := range result {
		key := string(kv.Key)
		if len(key) < 5 || key[:5] != "test." {
			t.Errorf("Expected key to start with 'test.', got %s", key)
		}
	}
}

// Custom type that implements TracerTagMarshaler
type customMarshalerType struct {
	Value string
	Count int
}

func (c customMarshalerType) MarshalTracerTag(builder *TracerTagBuilder) error {
	builder.String("custom_value", c.Value)
	builder.Int("custom_count", c.Count)
	return nil
}

func TestExpandObject_TracerTagMarshaler(t *testing.T) {
	custom := customMarshalerType{
		Value: "test_value",
		Count: 42,
	}

	result := expandObject("custom", custom)

	if len(result) != 2 {
		t.Errorf("Expected 2 KeyValue pairs from marshaler, got %d", len(result))
	}
}

// Custom type that implements TracerTagMarshaler but returns error
type errorMarshalerType struct{}

func (e errorMarshalerType) MarshalTracerTag(builder *TracerTagBuilder) error {
	return errors.New("marshaling error")
}

func TestExpandObject_TracerTagMarshalerError(t *testing.T) {
	errorCustom := errorMarshalerType{}

	result := expandObject("error", errorCustom)

	if len(result) != 1 {
		t.Errorf("Expected 1 KeyValue pair for error case, got %d", len(result))
	}

	// Should have error key
	key := string(result[0].Key)
	if key != "error_error" {
		t.Errorf("Expected error key 'error_error', got %s", key)
	}
}

func TestExpandObject_UnsupportedType(t *testing.T) {
	// Test with unsupported type
	unsupported := make(chan int)

	result := expandObject("unsupported", unsupported)

	// Should fall back to stringer
	if len(result) != 1 {
		t.Errorf("Expected 1 KeyValue pair for unsupported type, got %d", len(result))
	}
}

// Test SpanEvent system
func TestSpanEvent_SeverityEvent(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	tracer := Tracer("test-tracer")
	span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	// Create a severity event
	event := span.Debug("Test debug message")

	if !event.IsRecording() {
		t.Error("Expected event to be recording initially")
	}

	// Add tags and vars
	event.Tags(Key("test.tag").String("test.value"))
	event.Vars(map[string]any{"var1": "value1", "var2": 42})

	// Should still be recording
	if !event.IsRecording() {
		t.Error("Expected event to be recording after adding tags/vars")
	}

	// Flush the event
	event.Flush()

	// Should not be recording after flush
	if event.IsRecording() {
		t.Error("Expected event to not be recording after flush")
	}

	// Operations after flush should be no-ops
	event.Tags(Key("after.flush").String("ignored"))
	event.Vars(map[string]any{"ignored": true})
}

func TestSpanEvent_SeverityEventWithError(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	tracer := Tracer("test-tracer")
	span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	// Create event and set error
	event := span.Warning("Test warning with error")
	testErr := errors.New("test error")
	event.Error(testErr)

	// Flush should handle error case
	event.Flush()

	if event.IsRecording() {
		t.Error("Expected event to not be recording after flush")
	}
}

func TestSpanEvent_NopEvent(t *testing.T) {
	nop := nopEventInstance

	if nop.IsRecording() {
		t.Error("NopEvent should not be recording")
	}

	// All operations should be no-ops
	nop.Flush()
	nop.Error(errors.New("test"))
	nop.Tags(Key("test").String("value"))
	nop.Vars(map[string]any{"test": "value"})

	// Should still not be recording
	if nop.IsRecording() {
		t.Error("NopEvent should still not be recording after operations")
	}
}

// Test CompositeSpanExtractor
func TestCompositeSpanExtractor(t *testing.T) {
	// Create test span
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	tracer := Tracer("test-tracer")
	testSpan := tracer.Start(context.Background(), "test-span")
	defer testSpan.End()

	// Create extractors
	extractor1 := &testSpanExtractor{span: nil} // Returns nil
	extractor2 := &testSpanExtractor{span: testSpan} // Returns testSpan

	composite := NewCompositeSpanExtractor(extractor1, extractor2)

	ctx := context.Background()
	extracted := composite.Extract(ctx)

	// Should return testSpan from extractor2
	if extracted != testSpan {
		t.Error("Expected composite extractor to return testSpan from second extractor")
	}
}

func TestCompositeSpanExtractor_AllNil(t *testing.T) {
	extractor1 := &testSpanExtractor{span: nil}
	extractor2 := &testSpanExtractor{span: nil}

	composite := NewCompositeSpanExtractor(extractor1, extractor2)

	ctx := context.Background()
	extracted := composite.Extract(ctx)

	if extracted != nil {
		t.Error("Expected composite extractor to return nil when all extractors return nil")
	}
}

func TestCompositeSpanExtractor_Empty(t *testing.T) {
	composite := NewCompositeSpanExtractor()

	ctx := context.Background()
	extracted := composite.Extract(ctx)

	if extracted != nil {
		t.Error("Expected empty composite extractor to return nil")
	}
}

// Test infer and stringer functions
func TestInferFunction(t *testing.T) {
	testCases := []struct {
		name     string
		input    any
		expected bool // whether it should return valid KeyValue
	}{
		{"String", "test", true},
		{"Int", 42, true},
		{"Int64", int64(123), true},
		{"Float64", 3.14, true},
		{"Bool", true, true},
		{"StringSlice", []string{"a", "b"}, true},
		{"IntSlice", []int{1, 2}, true},
		{"Int64Slice", []int64{3, 4}, true},
		{"Float64Slice", []float64{1.1, 2.2}, true},
		{"BoolSlice", []bool{true, false}, true},
		{"Unsupported", make(chan int), false},
	}

	key := Key("test.key")
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := infer(key, tc.input)
			if tc.expected && !result.Valid() {
				t.Errorf("Expected valid KeyValue for %s", tc.name)
			}
			if !tc.expected && result.Valid() {
				t.Errorf("Expected invalid KeyValue for %s", tc.name)
			}
		})
	}
}

// Custom type that implements fmt.Stringer
type stringerType struct {
	value string
}

func (s stringerType) String() string {
	return fmt.Sprintf("stringer:%s", s.value)
}

func TestStringerFunction(t *testing.T) {
	key := Key("test.key")

	// Test with fmt.Stringer
	stringerObj := stringerType{value: "test"}
	result := stringer(key, stringerObj)

	if !result.Valid() {
		t.Error("Expected valid KeyValue for Stringer type")
	}

	// Test with non-Stringer
	nonStringerObj := struct{ value string }{value: "test"}
	result = stringer(key, nonStringerObj)

	if !result.Valid() {
		t.Error("Expected valid KeyValue for non-Stringer type (should use fmt.Sprintf)")
	}
}