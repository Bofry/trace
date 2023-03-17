package trace

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

func Test__InvalidKeyValue(t *testing.T) {
	ok := __InvalidKeyValue.Valid()
	expectedOk := false
	if ok != expectedOk {
		t.Errorf("__InvalidKeyValue.Valid(): expect %v, but got %v.", expectedOk, ok)
	}
}

func TestIsOtelNoopSpan(t *testing.T) {
	span := trace.SpanFromContext(context.Background())
	ok := IsOtelNoopSpan(span)
	expectedOk := true
	if ok != expectedOk {
		t.Errorf("IsOtelNoopSpan(): expect %v, but got %v.", expectedOk, ok)
	}
}

func TestIsNoopSeveritySpan(t *testing.T) {
	severitySpan := CreateSeveritySpan(context.Background())
	ok := IsNoopSeveritySpan(severitySpan)
	expectedOk := true
	if ok != expectedOk {
		t.Errorf("IsNoopSeveritySpan(): expect %v, but got %v.", expectedOk, ok)
	}
}
