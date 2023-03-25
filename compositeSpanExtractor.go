package trace

import "context"

var (
	_ SpanExtractor = CompositeSpanExtractor(nil)
)

type CompositeSpanExtractor []SpanExtractor

func NewCompositeSpanExtractor(extractors ...SpanExtractor) CompositeSpanExtractor {
	return CompositeSpanExtractor(extractors)
}

// Extract implements SpanExtractor
func (ext CompositeSpanExtractor) Extract(ctx context.Context) *SeveritySpan {
	for _, h := range ext {
		sp := h.Extract(ctx)
		if sp != nil {
			return sp
		}
	}
	return nil
}
