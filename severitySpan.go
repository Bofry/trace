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

	disabled bool
}

func (s *SeveritySpan) Disable(disabled bool) {
	s.disabled = disabled
}

func (s *SeveritySpan) IsDisabled() bool {
	return s.disabled
}

func (s *SeveritySpan) Context() context.Context {
	return s.ctx
}

func (s *SeveritySpan) HasSpanID() bool {
	return s.span.SpanContext().HasSpanID()
}

func (s *SeveritySpan) HasTraceID() bool {
	return s.span.SpanContext().HasTraceID()
}

func (s *SeveritySpan) SpanID() trace.SpanID {
	return s.span.SpanContext().SpanID()
}

func (s *SeveritySpan) TraceFlags() trace.TraceFlags {
	return s.span.SpanContext().TraceFlags()
}

func (s *SeveritySpan) TraceID() trace.TraceID {
	return s.span.SpanContext().TraceID()
}

func (s *SeveritySpan) TraceState() trace.TraceState {
	return s.span.SpanContext().TraceState()
}

func (s *SeveritySpan) Inject(p propagation.TextMapPropagator, c propagation.TextMapCarrier) {
	if p == nil {
		p = otel.GetTextMapPropagator()
	}
	p.Inject(s.ctx, c)
}

func (s *SeveritySpan) Link() Link {
	return Link(trace.LinkFromContext(s.ctx))
}

func (s *SeveritySpan) End(opts ...trace.SpanEndOption) {
	if s.disabled {
		return
	}

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

func (s *SeveritySpan) Tags(tags ...KeyValue) {
	if !s.span.IsRecording() {
		return
	}

	s.span.SetAttributes(tags...)
}

func (s *SeveritySpan) Argv(v any) {
	if !s.span.IsRecording() {
		return
	}

	if v == nil {
		return
	}

	kvset := expandObject(string(__ATTR_ARGV), v)
	s.span.SetAttributes(
		kvset...,
	)
}

func (s *SeveritySpan) Reply(code ReplyCode, v any) {
	if !s.span.IsRecording() {
		return
	}

	if v != nil {
		kvset := expandObject(string(__ATTR_REPLY), v)
		s.span.SetAttributes(
			kvset...,
		)
	}

	// output later
	s.replyCode = code
}

func (s *SeveritySpan) Err(err error) {
	if !s.span.IsRecording() {
		return
	}

	s.span.RecordError(err, trace.WithAttributes(
		__ATTR_EVENT_SEVERITY.String(ERR.Name()),
	), trace.WithStackTrace(true))

	// output later
	s.err = err
}

func (s *SeveritySpan) Debug(message string, v ...any) SpanEvent {
	if !s.span.IsRecording() {
		return nopEventInstance
	}

	event := &SeverityEvent{
		timestamp: time.Now(),
		span:      s.span,
		severity:  DEBUG,
		message:   fmt.Sprintf(message, v...),
	}
	s.events = append(s.events, event)
	return event
}

func (s *SeveritySpan) Info(message string, v ...any) SpanEvent {
	if !s.span.IsRecording() {
		return nopEventInstance
	}

	event := &SeverityEvent{
		timestamp: time.Now(),
		span:      s.span,
		severity:  INFO,
		message:   fmt.Sprintf(message, v...),
	}
	s.events = append(s.events, event)
	return event
}

func (s *SeveritySpan) Notice(message string, v ...any) SpanEvent {
	if !s.span.IsRecording() {
		return nopEventInstance
	}

	event := &SeverityEvent{
		timestamp: time.Now(),
		span:      s.span,
		severity:  NOTICE,
		message:   fmt.Sprintf(message, v...),
	}
	s.events = append(s.events, event)
	return event
}

func (s *SeveritySpan) Warning(reason string, v ...any) SpanEvent {
	if !s.span.IsRecording() {
		return nopEventInstance
	}

	event := &SeverityEvent{
		timestamp: time.Now(),
		span:      s.span,
		severity:  WARN,
		message:   fmt.Sprintf(reason, v...),
	}
	s.events = append(s.events, event)
	return event
}

func (s *SeveritySpan) Crit(reason string, v ...any) SpanEvent {
	if !s.span.IsRecording() {
		return nopEventInstance
	}

	event := &SeverityEvent{
		timestamp: time.Now(),
		span:      s.span,
		severity:  CRIT,
		message:   fmt.Sprintf(reason, v...),
	}
	s.events = append(s.events, event)
	return event
}

func (s *SeveritySpan) Alert(reason string, v ...any) SpanEvent {
	if !s.span.IsRecording() {
		return nopEventInstance
	}

	event := &SeverityEvent{
		timestamp: time.Now(),
		span:      s.span,
		severity:  ALERT,
		message:   fmt.Sprintf(reason, v...),
	}
	s.events = append(s.events, event)
	return event
}

func (s *SeveritySpan) Emerg(reason string, v ...any) SpanEvent {
	if !s.span.IsRecording() {
		return nopEventInstance
	}

	event := &SeverityEvent{
		timestamp: time.Now(),
		span:      s.span,
		severity:  EMERG,
		message:   fmt.Sprintf(reason, v...),
	}
	s.events = append(s.events, event)
	return event
}

func (s *SeveritySpan) otelSpan() trace.Span {
	return s.span
}
