package trace

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type SeverityTracer struct {
	tr trace.Tracer
}

func NewSeverityTracer(tr trace.Tracer) *SeverityTracer {
	return &SeverityTracer{
		tr: tr,
	}
}

func (s *SeverityTracer) Open(
	ctx context.Context,
	spanName string,
	opts ...trace.SpanStartOption) SeveritySpanContext {

	if ctx == nil {
		ctx = context.Background()
	}
	ctx, span := s.tr.Start(ctx, spanName, opts...)
	return &SeveritySpan{
		span: span,
		ctx:  ctx,
	}
}

func (s *SeverityTracer) Start(
	spx SeveritySpanContext,
	spanName string,
	opts ...trace.SpanStartOption) SeveritySpanContext {

	ctx, span := s.tr.Start(spx.Context(), spanName, opts...)
	return &SeveritySpan{
		span: span,
		ctx:  ctx,
	}
}

func (s *SeverityTracer) Link(
	ctx context.Context,
	link Link,
	spanName string,
	opts ...trace.SpanStartOption) SeveritySpanContext {

	if ctx == nil {
		ctx = context.Background()
	}
	ctx = trace.ContextWithSpanContext(ctx, link.SpanContext)
	opts = append(opts, trace.WithLinks(link))
	return s.Open(ctx, spanName, opts...)
}

func (s *SeverityTracer) ExtractWithPropagator(
	ctx context.Context,
	propagator propagation.TextMapPropagator,
	carrier propagation.TextMapCarrier,
	spanName string,
	opts ...trace.SpanStartOption) SeveritySpanContext {

	if ctx == nil {
		ctx = context.Background()
	}
	if propagator == nil {
		propagator = otel.GetTextMapPropagator()
	}
	ctx = propagator.Extract(ctx, carrier)
	return s.Open(ctx, spanName, opts...)
}

func (s *SeverityTracer) Extract(
	ctx context.Context,
	carrier propagation.TextMapCarrier,
	spanName string,
	opts ...trace.SpanStartOption) SeveritySpanContext {

	if ctx == nil {
		ctx = context.Background()
	}
	propagator := otel.GetTextMapPropagator()
	return s.ExtractWithPropagator(ctx, propagator, carrier, spanName, opts...)
}

func (s *SeverityTracer) otelTracer() trace.Tracer {
	return s.tr
}
