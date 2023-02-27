package trace

import (
	"time"

	"go.opentelemetry.io/otel/trace"
)

type SpanEvent interface {
	IsRecording() bool
	Flush()
	Error(err error)
	Tags(tags ...KeyValue)
	Vars(v interface{})
}

var (
	_ SpanEvent = SeverityEvent{}
	_ SpanEvent = NopEvent(0)
)

var (
	nopEventInstance = NopEvent(0)
)

type NopEvent int

// IsRecording implements SpanEvent
func (NopEvent) IsRecording() bool {
	return false
}

// Flush implements SpanEvent
func (NopEvent) Flush() {
	// do nothing
}

// Error implements SpanEvent
func (NopEvent) Error(err error) {
	// do nothing
}

// Tags implements SpanEvent
func (NopEvent) Tags(tags ...KeyValue) {
	// do nothing
}

// Vars implements SpanEvent
func (NopEvent) Vars(v interface{}) {
	// do nothing
}

type SeverityEvent struct {
	span trace.Span

	timestamp time.Time
	severity  Severity
	message   string
	flushed   bool
	tags      []KeyValue
	err       error
}

// IsRecording implements SpanEvent
func (s SeverityEvent) IsRecording() bool {
	return !s.flushed
}

// Flush implements SpanEvent
func (s SeverityEvent) Flush() {
	if !s.flushed {
		s.flushed = true

		s.tags = append(s.tags,
			__ATTR_EVENT_MESSAGE.String(s.message),
			__ATTR_EVENT_SEVERITY.String(s.severity.Name()),
		)

		if s.err != nil {
			err := s.err
			s.span.RecordError(err, trace.WithAttributes(
				s.tags...,
			), trace.WithTimestamp(s.timestamp))
		} else {
			s.span.AddEvent(string(__ATTR_EVENT), trace.WithAttributes(
				s.tags...,
			), trace.WithTimestamp(s.timestamp))
		}

	}
}

// Error implements SpanEvent
func (s SeverityEvent) Error(err error) {
	if !s.flushed {
		s.flushed = true

		s.err = err
	}
}

// Tags implements SpanEvent
func (s SeverityEvent) Tags(tags ...KeyValue) {
	if !s.flushed {
		s.flushed = true

		s.tags = append(s.tags, tags...)
	}
}

// Vars implements SpanEvent
func (s SeverityEvent) Vars(v interface{}) {
	if !s.flushed {
		s.flushed = true

		tags := expandObject(string(__ATTR_VARS), v)
		s.tags = append(s.tags, tags...)
	}
}
