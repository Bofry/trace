package trace

import (
	"context"
)

var (
	_ SpanInjector = NoopSpanInjector(0)

	noopSpanInjectorInstance = NoopSpanInjector(0)
)

type NoopSpanInjector int

// Inject implements SpanInjector.
func (NoopSpanInjector) Inject(ctx context.Context, span *SeveritySpan) context.Context {
	return ctx
}
