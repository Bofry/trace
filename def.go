package trace

import (
	"context"
	"sync/atomic"
	_ "unsafe"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	__severity_none__ = -1

	NONE   Severity = __severity_none__
	DEBUG  Severity = iota // severity : 0
	INFO                   // severity : 1
	NOTICE                 // severity : 2
	WARN                   // severity : 3
	ERR                    // severity : 4
	CRIT                   // severity : 5
	ALERT                  // severity : 6
	EMERG                  // severity : 7

	__severity_minimum__ = DEBUG
	__severity_maximum__ = EMERG
)

const (
	__STATUS_CODE_ERROR = "error"

	PASS  ReplyCode = ReplyCode("pass")
	FAIL  ReplyCode = ReplyCode("fail")
	UNSET ReplyCode = ReplyCode("")
)

const (
	__ATTR_EVENT                    attribute.Key = "event"
	__ATTR_EVENT_MESSAGE            attribute.Key = "event.message"
	__ATTR_EVENT_SEVERITY           attribute.Key = "event.severity"
	__ATTR_EVENT_STATUS_CODE        attribute.Key = "event.status_code"
	__ATTR_EVENT_STATUS_DESCRIPTION attribute.Key = "event.status_description"
	__ATTR_ERROR                    attribute.Key = "error"
	__ATTR_ARGV                     attribute.Key = "argv"
	__ATTR_REPLY                    attribute.Key = "reply"
	__ATTR_VARS                     attribute.Key = "vars"

	__ATTR_ENVIRONMENT  attribute.Key = "environment"
	__ATTR_OS           attribute.Key = "os"
	__ATTR_PID          attribute.Key = "pid"
	__ATTR_SIGNATURE    attribute.Key = "signature"
	__ATTR_VERSION      attribute.Key = "version"
	__ATTR_FACILITY     attribute.Key = "facility"
	__ATTR_SERVICE_NAME attribute.Key = semconv.ServiceNameKey

	__ATTR_HTTP_METHOD       attribute.Key = "http.method"
	__ATTR_HTTP_REQUEST      attribute.Key = "http.request"
	__ATTR_HTTP_REQUEST_PATH attribute.Key = "http.request_path"
	__ATTR_HTTP_RESPONSE     attribute.Key = "http.response"
	__ATTR_HTTP_USER_AGENT   attribute.Key = "http.user_agent"
)

const (
	// SpanKindUnspecified is an unspecified SpanKind and is not a valid
	// SpanKind. SpanKindUnspecified should be replaced with SpanKindInternal
	// if it is received.
	SpanKindUnspecified = trace.SpanKindUnspecified
	// SpanKindInternal is a SpanKind for a Span that represents an internal
	// operation within an application.
	SpanKindInternal = trace.SpanKindInternal
	// SpanKindServer is a SpanKind for a Span that represents the operation
	// of handling a request from a client.
	SpanKindServer = trace.SpanKindServer
	// SpanKindClient is a SpanKind for a Span that represents the operation
	// of client making a request to a server.
	SpanKindClient = trace.SpanKindClient
	// SpanKindProducer is a SpanKind for a Span that represents the operation
	// of a producer sending a message to a message broker. Unlike
	// SpanKindClient and SpanKindServer, there is often no direct
	// relationship between this kind of Span and a SpanKindConsumer kind. A
	// SpanKindProducer Span will end once the message is accepted by the
	// message broker which might not overlap with the processing of that
	// message.
	SpanKindProducer = trace.SpanKindProducer
	// SpanKindConsumer is a SpanKind for a Span that represents the operation
	// of a consumer receiving a message from a message broker. Like
	// SpanKindProducer Spans, there is often no direct relationship between
	// this Span and the Span that produced the message.
	SpanKindConsumer = trace.SpanKindConsumer
)

const (
	// FlagsSampled is a bitmask with the sampled bit set. A SpanContext
	// with the sampling bit set means the span is sampled.
	FlagsSampled = trace.FlagsSampled
)

type (
	KeyValue = attribute.KeyValue
	Key      = attribute.Key
	SpanKind = trace.SpanKind
	Link     = trace.Link

	ReplyCode string

	TracerTagMarshaler interface {
		MarshalTracerTag(builder *TracerTagBuilder) error
	}

	SpanExtractor interface {
		Extract(ctx context.Context) *SeveritySpan
	}

	tracerProviderHolder struct {
		v *SeverityTracerProvider
	}

	spanExtractorHolder struct {
		v SpanExtractor
	}
)

var (
	__InvalidKeyValue = attribute.Bool("", false)

	globalTracerProvider = defaultTracerProviderValue()
	globalSpanExtractor  = defaultSpanExtractor()

	noopSpan = trace.SpanFromContext(context.Background())
)

//go:linkname WithStackTrace go.opentelemetry.io/otel/trace.WithStackTrace
func WithStackTrace(b bool) trace.SpanEndEventOption

//go:linkname WithLinks go.opentelemetry.io/otel/trace.WithLinks
func WithLinks(links ...Link) trace.SpanStartOption

//go:linkname WithNewRoot go.opentelemetry.io/otel/trace.WithNewRoot
func WithNewRoot() trace.SpanStartOption

//go:linkname WithSpanKind go.opentelemetry.io/otel/trace.WithSpanKind
func WithSpanKind(kind SpanKind) trace.SpanStartOption

//go:linkname WithInstrumentationVersion go.opentelemetry.io/otel/trace.WithInstrumentationVersion
func WithInstrumentationVersion(version string) trace.TracerOption

//go:linkname WithSchemaURL go.opentelemetry.io/otel/trace.WithSchemaURL
func WithSchemaURL(schemaURL string) trace.TracerOption

//go:linkname GetTextMapPropagator go.opentelemetry.io/otel.GetTextMapPropagator
func GetTextMapPropagator() propagation.TextMapPropagator

//go:linkname SetTextMapPropagator go.opentelemetry.io/otel.SetTextMapPropagator
func SetTextMapPropagator(propagator propagation.TextMapPropagator)

func GetTracerProvider() *SeverityTracerProvider {
	return globalTracerProvider.Load().(tracerProviderHolder).v
}

func SetTracerProvider(tp *SeverityTracerProvider) {
	current := GetTracerProvider()
	if current != tp {
		globalTracerProvider.Store(tracerProviderHolder{
			v: tp,
		})
	}
}

func GetSpanExtractor() SpanExtractor {
	return globalSpanExtractor.Load().(spanExtractorHolder).v
}

func SetSpanExtractor(extractor SpanExtractor) {
	current := GetSpanExtractor()
	if current != extractor {
		globalSpanExtractor.Store(spanExtractorHolder{
			v: extractor,
		})
	}
}

func Tracer(name string, opts ...trace.TracerOption) *SeverityTracer {
	return GetTracerProvider().Tracer(name, opts...)
}

func OtelSpanFromSeveritySpan(span *SeveritySpan) trace.Span {
	return span.otelSpan()
}

func OtelTracerFromTracer(tr *SeverityTracer) trace.Tracer {
	return tr.otelTracer()
}

func IsOtelNoopSpan(span trace.Span) bool {
	return span == noopSpan
}

func IsNoopSeveritySpan(span *SeveritySpan) bool {
	return IsOtelNoopSpan(span.otelSpan())
}

func Argv() VarsBuilder {
	return make(VarsBuilder)
}

func Vars() VarsBuilder {
	return make(VarsBuilder)
}

func SpanFromContext(ctx context.Context, extractors ...SpanExtractor) *SeveritySpan {
	var span *SeveritySpan

	for _, v := range extractors {
		span = v.Extract(ctx)
		if span != nil {
			return span
		}
	}

	var globalSpanExtractor = GetSpanExtractor()
	span = globalSpanExtractor.Extract(ctx)
	if span != nil {
		return span
	}

	return &SeveritySpan{
		span: trace.SpanFromContext(ctx),
		ctx:  ctx,
	}
}

func CreateSeverityTracerProvider(provider trace.TracerProvider) *SeverityTracerProvider {
	return &SeverityTracerProvider{
		provider: provider,
	}
}

func CreateSeverityTracer(tr trace.Tracer) *SeverityTracer {
	return &SeverityTracer{
		tr: tr,
	}
}

func CreateSeveritySpan(ctx context.Context) *SeveritySpan {
	if ctx == nil {
		ctx = context.Background()
	}
	span := trace.SpanFromContext(ctx)
	return &SeveritySpan{
		span: span,
		ctx:  ctx,
	}
}

func defaultTracerProviderValue() *atomic.Value {
	v := &atomic.Value{}
	v.Store(tracerProviderHolder{
		v: &SeverityTracerProvider{
			provider: otel.GetTracerProvider(),
		},
	})
	return v
}

func defaultSpanExtractor() *atomic.Value {
	v := &atomic.Value{}
	v.Store(spanExtractorHolder{
		v: NewCompositeSpanExtractor(),
	})
	return v
}
