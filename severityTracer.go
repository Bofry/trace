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

func (s *SeverityTracer) Open(
	ctx context.Context,
	spanName string,
	opts ...trace.SpanStartOption) *SeveritySpan {

	opts = append(opts, trace.WithNewRoot())
	return s.Start(ctx, spanName, opts...)
}

func (s *SeverityTracer) Start(
	ctx context.Context,
	spanName string,
	opts ...trace.SpanStartOption) *SeveritySpan {

	if ctx == nil {
		ctx = context.Background()
	}
	ctx, span := s.tr.Start(ctx, spanName, opts...)
	return &SeveritySpan{
		span:   span,
		ctx:    ctx,
		events: make([]SpanEvent, 0, 4),
	}
}

func (s *SeverityTracer) Link(
	ctx context.Context,
	link Link,
	spanName string,
	opts ...trace.SpanStartOption) *SeveritySpan {

	if ctx == nil {
		ctx = context.Background()
	}
	ctx = trace.ContextWithSpanContext(ctx, link.SpanContext)
	opts = append(opts, trace.WithLinks(link))
	return s.Start(ctx, spanName, opts...)
}

func (s *SeverityTracer) ExtractWithPropagator(
	ctx context.Context,
	propagator propagation.TextMapPropagator,
	carrier propagation.TextMapCarrier,
	spanName string,
	opts ...trace.SpanStartOption) *SeveritySpan {

	if ctx == nil {
		ctx = context.Background()
	}
	if propagator == nil {
		propagator = otel.GetTextMapPropagator()
	}
	ctx = propagator.Extract(ctx, carrier)
	return s.Start(ctx, spanName, opts...)
}

func (s *SeverityTracer) Extract(
	ctx context.Context,
	carrier propagation.TextMapCarrier,
	spanName string,
	opts ...trace.SpanStartOption) *SeveritySpan {

	if ctx == nil {
		ctx = context.Background()
	}
	propagator := otel.GetTextMapPropagator()
	return s.ExtractWithPropagator(ctx, propagator, carrier, spanName, opts...)
}

func (s *SeverityTracer) InjectWithPropagator(
	ctx context.Context,
	propagator propagation.TextMapPropagator,
	carrier propagation.TextMapCarrier) {

	if ctx == nil {
		return
	}

	sp := SpanFromContext(ctx)
	if IsNoopSeveritySpan(sp) {
		return
	}

	propagator.Inject(sp.ctx, carrier)
}

func (s *SeverityTracer) Inject(
	ctx context.Context,
	carrier propagation.TextMapCarrier) {

	propagator := otel.GetTextMapPropagator()

	s.InjectWithPropagator(ctx, propagator, carrier)
}

func (s *SeverityTracer) otelTracer() trace.Tracer {
	return s.tr
}
