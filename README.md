# Trace - Severity-based Distributed Tracing Library

[![Go Version](https://img.shields.io/badge/go-1.24+-blue.svg)](https://golang.org/dl/)
[![OpenTelemetry](https://img.shields.io/badge/OpenTelemetry-v1.38.0-orange.svg)](https://opentelemetry.io/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/Bofry/trace)](https://goreportcard.com/report/github.com/Bofry/trace)
[![Jaeger Compatible](https://img.shields.io/badge/Jaeger-Compatible-yellow.svg)](https://www.jaegertracing.io/)
[![OTLP Support](https://img.shields.io/badge/OTLP-Supported-brightgreen.svg)](https://opentelemetry.io/docs/reference/specification/protocol/otlp/)

> **[中文版文件 (Chinese Documentation)](./README_ZH.md)**
> **推薦閱讀中文版，包含更詳細的使用說明和最佳實踐範例**

**github.com/Bofry/trace** is a structured, severity-based distributed
tracing library built on OpenTelemetry. It enhances standard OpenTelemetry spans
with severity levels (DEBUG through EMERG) and provides fluent APIs for
structured logging and cross-service trace propagation with Jaeger integration.

## Key Features

- **8-Level Severity System**: Following syslog standard (DEBUG to EMERG)
- **Fluent API**: Intuitive method chaining for span operations
- **Cross-Service Tracing**: Complete distributed tracing context propagation
- **Deferred Event System**: Efficient event batching with flush-on-end
- **No-op Detection**: Automatic performance optimization for disabled tracing

## Installation

```bash
go get github.com/Bofry/trace
```

## Quick Start

```go
package main

import (
    "context"
    "github.com/Bofry/trace"
    "log"
    "time"
)

func main() {
    // Create TracerProvider (OTLP - Recommended)
    tp, err := trace.OTLPProvider("http://localhost:4318",
        trace.ServiceName("my-service"),
        trace.Environment("production"),
        trace.Pid(),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Set as global provider
    trace.SetTracerProvider(tp)

    // Create tracer and span
    tracer := trace.Tracer("main")
    span := tracer.Open(context.Background(), "main-operation")
    defer span.End()

    // Use severity-based logging
    span.Info("Operation started")
    span.Debug("Debug info: %s", "some debug data")

    // Record structured data
    span.Argv(map[string]any{"user_id": 123, "action": "create"})
    span.Reply(trace.PASS, "Operation completed")

    // Graceful shutdown
    defer func(ctx context.Context) {
        ctx, cancel := context.WithTimeout(ctx, time.Second*5)
        defer cancel()
        tp.Shutdown(ctx)
    }(context.Background())
}
```

## Architecture Overview

```mermaid
graph TB
    A[SeverityTracerProvider] --> B[SeverityTracer]
    B --> C[SeveritySpan]
    C --> D[SpanEvent]

    A --> E[OpenTelemetry TracerProvider]
    B --> F[OpenTelemetry Tracer]
    C --> G[OpenTelemetry Span]

    style A fill:#e1f5fe
    style C fill:#f3e5f5
```

## Severity Levels

| Level | Value | Name | Usage |
|-------|-------|------|-------|
| DEBUG | 0 | debug | Debug information |
| INFO | 1 | info | General information |
| NOTICE | 2 | notice | Important information |
| WARN | 3 | warn | Warning messages |
| ERR | 4 | err | Error messages |
| CRIT | 5 | crit | Critical errors |
| ALERT | 6 | alert | Immediate attention needed |
| EMERG | 7 | emerg | System unusable |

## Span Management Patterns

### Provider Initialization

```go
// Modern OTLP approach (recommended)
tp, err := trace.OTLPProvider("http://localhost:4318",
    trace.ServiceName("my-service"),
    trace.Environment("production"),
)

// Legacy Jaeger compatibility (auto-converts to OTLP)
tp, err := trace.JaegerCompatibleProvider("http://localhost:14268/api/traces",
    trace.ServiceName("my-service"),
)
```

### Span Creation Methods

```go
// Root span - creates new trace
rootSpan := tracer.Open(ctx, "root-operation")

// Child span - inherits current context
childSpan := tracer.Start(parentSpan.Context(), "child-operation")

// Linked span - related but not parent-child
linkedSpan := tracer.Link(ctx, parentSpan.Link(), "linked-operation")

// Extracted span - from cross-service context
extractedSpan := tracer.Extract(ctx, carrier, "extracted-operation")
```

## Cross-Service Tracing

```mermaid
sequenceDiagram
    participant A as Service A
    participant B as Service B
    participant J as Jaeger

    A->>A: Create Span
    A->>A: span.Inject(carrier)
    A->>B: HTTP Request + Headers
    B->>B: tracer.Extract(carrier)
    B->>B: Create Child Span
    B->>J: Send Traces
    A->>J: Send Traces
    J->>J: Correlate Traces
```

### Context Propagation Example

```go
// Service A - Inject context
carrier := make(propagation.MapCarrier)
span.Inject(nil, carrier)

// Pass through HTTP headers
for key, value := range carrier {
    req.Header.Set(key, value)
}

// Service B - Extract context
carrier := make(propagation.MapCarrier)
for key, values := range req.Header {
    if len(values) > 0 {
        carrier.Set(key, values[0])
    }
}

extractedSpan := tracer.Extract(ctx, carrier, "downstream-operation")
```

## Testing Setup

### Jaeger with Docker

```bash
docker run -d --name jaeger \
  -p 16686:16686 \
  -p 14268:14268 \
  -p 4317:4317 \
  -p 4318:4318 \
  jaegertracing/all-in-one:latest
```

### Test Configuration

```bash
# Copy test environment file
cp trace_test.env.sample trace_test.env

# Edit with your Jaeger endpoints
# JAEGER_TRACE_URL=http://127.0.0.1:14268/api/traces
# JAEGER_QUERY_URL=http://127.0.0.1:16686/api/traces
```

### Development Workflow

```bash
# Standard development sequence
go mod tidy           # Clean up dependencies
go fmt ./...          # Format all code
go vet ./...          # Static analysis
go test ./...         # Run all tests

# Additional testing
go test -cover ./...  # Coverage analysis
go test -race ./...   # Race condition detection
```

## Performance Considerations

- **No-op Detection**: Automatically optimizes performance when tracing is disabled
- **Event Batching**: Deferred event processing with flush-on-end
- **Atomic Operations**: Thread-safe global state management
- **Lazy Evaluation**: Attributes constructed only when needed

## Test Coverage & Performance Analysis

### Test Statistics

- **Test Coverage**: 90.7% (89/89 tests passing)
- **Total Test Files**: 5 comprehensive test suites
- **Benchmark Tests**: 20 performance benchmarks

### Performance Benchmarks

#### Core Operations (ops/sec)

```textplain
BenchmarkSeveritySpan_Debug         42,077,848     28.44 ns/op     72 B/op    2 allocs/op
BenchmarkSeveritySpan_Info          43,334,922     27.89 ns/op     72 B/op    2 allocs/op
BenchmarkSeveritySpan_Warning       44,706,405     27.77 ns/op     72 B/op    2 allocs/op
BenchmarkSeveritySpan_NoopSpan      41,718,452     27.13 ns/op     72 B/op    2 allocs/op
```

#### Additional Operations

```textplain
BenchmarkExpandObject_String        72,578,631     16.97 ns/op     64 B/op    1 allocs/op
BenchmarkExpandObject_Map            3,132,738     384.5 ns/op    640 B/op   15 allocs/op
BenchmarkSpanFromContext           450,409,731      2.676 ns/op     0 B/op    0 allocs/op
BenchmarkSpanEvent_Creation        100,000,000     10.62 ns/op      8 B/op    0 allocs/op
BenchmarkNoopEvent_Operations      665,452,064      1.811 ns/op     0 B/op    0 allocs/op
```

#### Memory Efficiency

- **Recording Span**: ~28ns per operation, 72 bytes allocated
- **Context Operations**: ~3ns per operation, zero allocations
- **Event Creation**: ~11ns per operation, minimal allocations
- **No-op Operations**: ~2ns per operation, zero allocations

## API Reference

### Core Types

- `SeverityTracerProvider`: Enhanced tracer provider
- `SeverityTracer`: Creates severity-enabled spans
- `SeveritySpan`: Span with severity logging methods
- `SpanEvent`: Deferred event system
- `Severity`: 8-level severity enum

### Provider Creation

- `OTLPProvider(endpoint, attrs...)`: **RECOMMENDED** - OTLP HTTP provider
- `OTLPGRPCProvider(endpoint, attrs...)`: OTLP gRPC provider
- `JaegerCompatibleProvider(url, attrs...)`: Legacy Jaeger compatibility layer (auto-converts endpoints to OTLP)

### Span Methods

- **Severity Logging**: `Debug()`, `Info()`, `Notice()`, `Warning()`,
  `Crit()`, `Alert()`, `Emerg()`
- **Data Recording**: `Argv()`, `Reply()`, `Tags()`, `Err()`
- **Context Operations**: `Inject()`, `Link()`, `Context()`

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the [MIT License](LICENSE).

## Related Projects

- [OpenTelemetry Go](https://github.com/open-telemetry/opentelemetry-go)
- [Jaeger Tracing](https://www.jaegertracing.io/)
- [OTLP Protocol](https://opentelemetry.io/docs/reference/specification/protocol/otlp/)

---

**Note**: This library uses modern OTLP protocol as the primary transport.
For Jaeger compatibility, use `JaegerCompatibleProvider` which automatically
converts Jaeger endpoints to OTLP (port 14268→4318, 14250→4317).

> **[完整中文文件請參閱 README_ZH.md](./README_ZH.md)**
