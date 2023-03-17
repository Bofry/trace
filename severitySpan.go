package trace

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type SeveritySpan struct {
	span      trace.Span
	ctx       context.Context
	replyCode ReplyCode
	err       error

	events []SpanEvent
}

// Context implements SpanContext
func (s *SeveritySpan) Context() context.Context {
	return s.ctx
}

// HasSpanID implements SpanContext
func (s *SeveritySpan) HasSpanID() bool {
	return s.span.SpanContext().HasSpanID()
}

// HasTraceID implements SpanContext
func (s *SeveritySpan) HasTraceID() bool {
	return s.span.SpanContext().HasTraceID()
}

// SpanID implements SpanContext
func (s *SeveritySpan) SpanID() trace.SpanID {
	return s.span.SpanContext().SpanID()
}

// TraceFlags implements SpanContext
func (s *SeveritySpan) TraceFlags() trace.TraceFlags {
	return s.span.SpanContext().TraceFlags()
}

// TraceID implements SpanContext
func (s *SeveritySpan) TraceID() trace.TraceID {
	return s.span.SpanContext().TraceID()
}

// TraceState implements SpanContext
func (s *SeveritySpan) TraceState() trace.TraceState {
	return s.span.SpanContext().TraceState()
}

// Inject implements SeveritySpanContext
func (s *SeveritySpan) Inject(p propagation.TextMapPropagator, c propagation.TextMapCarrier) {
	if p == nil {
		p = otel.GetTextMapPropagator()
	}
	p.Inject(s.ctx, c)
}

// Link implements SpanContext
func (s *SeveritySpan) Link() Link {
	return Link(trace.LinkFromContext(s.ctx))
}

// End implements SpanContext
func (s *SeveritySpan) End(opts ...trace.SpanEndOption) {
	for _, e := range s.events {
		e.Flush()
	}

	if s.err != nil {
		s.span.SetAttributes(
			__ATTR_ERROR.Bool(true),
			__ATTR_EVENT_STATUS_CODE.String(__STATUS_CODE_ERROR),
			__ATTR_EVENT_STATUS_DESCRIPTION.String(s.err.Error()),
		)
	} else if len(s.replyCode) > 0 {
		s.span.SetAttributes(
			__ATTR_EVENT_STATUS_CODE.String(string(s.replyCode)),
		)
	}
	s.span.End(opts...)
}

// Tags implements SpanContext
func (s *SeveritySpan) Tags(tags ...KeyValue) {
	s.span.SetAttributes(tags...)
}

// Argv implements SpanContext
func (s *SeveritySpan) Argv(v interface{}) {
	kvset := expandObject(string(__ATTR_ARGV), v)
	s.span.SetAttributes(
		kvset...,
	)
}

// Reply implements SpanContext
func (s *SeveritySpan) Reply(code ReplyCode, v interface{}) {
	kvset := expandObject(string(__ATTR_REPLY), v)
	s.span.SetAttributes(
		kvset...,
	)

	// output later
	s.replyCode = code
}

// Err implements SpanContext
func (s *SeveritySpan) Err(err error) {
	if s.span.IsRecording() {
		s.span.RecordError(err, trace.WithAttributes(
			__ATTR_EVENT_SEVERITY.String(ERR.Name()),
		), trace.WithStackTrace(true))

		// output later
		s.err = err
	}
}

// Debug implements SpanContext
func (s *SeveritySpan) Debug(message string, v ...interface{}) SpanEvent {
	if s.span.IsRecording() {
		event := &SeverityEvent{
			timestamp: time.Now(),
			span:      s.span,
			severity:  DEBUG,
			message:   fmt.Sprintf(message, v...),
		}
		s.events = append(s.events, event)
		return event
	}
	return nopEventInstance
}

// Info implements SpanContext
func (s *SeveritySpan) Info(message string, v ...interface{}) SpanEvent {
	if s.span.IsRecording() {
		event := &SeverityEvent{
			timestamp: time.Now(),
			span:      s.span,
			severity:  INFO,
			message:   fmt.Sprintf(message, v...),
		}
		s.events = append(s.events, event)
		return event
	}
	return nopEventInstance
}

// Notice implements SpanContext
func (s *SeveritySpan) Notice(message string, v ...interface{}) SpanEvent {
	if s.span.IsRecording() {
		event := &SeverityEvent{
			timestamp: time.Now(),
			span:      s.span,
			severity:  NOTICE,
			message:   fmt.Sprintf(message, v...),
		}
		s.events = append(s.events, event)
		return event
	}
	return nopEventInstance
}

// Warning implements SpanContext
func (s *SeveritySpan) Warning(reason string, v ...interface{}) SpanEvent {
	if s.span.IsRecording() {
		event := &SeverityEvent{
			timestamp: time.Now(),
			span:      s.span,
			severity:  WARN,
			message:   fmt.Sprintf(reason, v...),
		}
		s.events = append(s.events, event)
		return event
	}
	return nopEventInstance
}

// Crit implements SpanContext
func (s *SeveritySpan) Crit(reason string, v ...interface{}) SpanEvent {
	if s.span.IsRecording() {
		event := &SeverityEvent{
			timestamp: time.Now(),
			span:      s.span,
			severity:  CRIT,
			message:   fmt.Sprintf(reason, v...),
		}
		s.events = append(s.events, event)
		return event
	}
	return nopEventInstance
}

// Alert implements SpanContext
func (s *SeveritySpan) Alert(reason string, v ...interface{}) SpanEvent {
	if s.span.IsRecording() {
		event := &SeverityEvent{
			timestamp: time.Now(),
			span:      s.span,
			severity:  ALERT,
			message:   fmt.Sprintf(reason, v...),
		}
		s.events = append(s.events, event)
		return event
	}
	return nopEventInstance
}

// Emerg implements SpanContext
func (s *SeveritySpan) Emerg(reason string, v ...interface{}) SpanEvent {
	if s.span.IsRecording() {
		event := &SeverityEvent{
			timestamp: time.Now(),
			span:      s.span,
			severity:  EMERG,
			message:   fmt.Sprintf(reason, v...),
		}
		s.events = append(s.events, event)
		return event
	}
	return nopEventInstance
}

// otelSpan implements SeveritySpanContext
func (s *SeveritySpan) otelSpan() trace.Span {
	return s.span
}
