package trace_test

import (
	"context"
	"log"
	"time"
	"trace"

	"go.opentelemetry.io/otel/propagation"
)

func Example() {
	tp, err := trace.JaegerProvider("http://localhost:14268/api/traces",
		trace.ServiceName("trace-demo"),
		trace.Environment("go-test"),
		trace.Pid(),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	trace.SetTracerProvider(tp)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Cleanly shutdown and flush telemetry when the application exits.
	defer func(ctx context.Context) {
		// Do not make the application hang when it is shutdown.
		ctx, cancel = context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}(ctx)

	tr := tp.Tracer("example-main")

	spx := tr.Open(ctx, "example()", trace.WithNewRoot())
	defer spx.End()

	spx.Info("example starting")
	// Output:
}

func ExampleSeverityTracer_Start() {
	tp, err := trace.JaegerProvider("http://localhost:14268/api/traces",
		trace.ServiceName("trace-demo"),
		trace.Environment("go-test"),
		trace.Pid(),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	trace.SetTracerProvider(tp)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Cleanly shutdown and flush telemetry when the application exits.
	defer func(ctx context.Context) {
		// Do not make the application hang when it is shutdown.
		ctx, cancel = context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}(ctx)

	tr := tp.Tracer("example-main")

	spx := tr.Open(ctx, "example()", trace.WithNewRoot())
	defer spx.End()

	spx.Info("example starting")

	// subroutine
	var bar = func(spx trace.SeveritySpanContext, arg string) {
		barspx := tr.Start(spx, "bar()")
		defer barspx.End()

		barspx.Argv(arg)
		barspx.Reply(trace.PASS, "OK")
	}

	// call subroutine
	bar(spx, "foo")
	// Output:
}

func ExampleSeverityTracer_Link() {
	tp, err := trace.JaegerProvider("http://localhost:14268/api/traces",
		trace.ServiceName("trace-demo"),
		trace.Environment("go-test"),
		trace.Pid(),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	trace.SetTracerProvider(tp)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Cleanly shutdown and flush telemetry when the application exits.
	defer func(ctx context.Context) {
		// Do not make the application hang when it is shutdown.
		ctx, cancel = context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}(ctx)

	tr := tp.Tracer("example-main")

	spx := tr.Open(ctx, "example()", trace.WithNewRoot())
	defer spx.End()

	spx.Info("example starting")

	// subroutine
	var bar = func(link trace.Link, arg string) {
		barctx := context.Background() // can use nil instead of context.Background()
		barspx := tr.Link(barctx, link, "bar()")
		defer barspx.End()

		barspx.Argv(arg)
		barspx.Reply(trace.PASS, "OK")
	}

	// call subroutine
	bar(spx.Link(), "foo")
	// Output:
}

func ExampleSeverityTracer_ExtractWithPropagator() {
	tp, err := trace.JaegerProvider("http://localhost:14268/api/traces",
		trace.ServiceName("trace-demo"),
		trace.Environment("go-test"),
		trace.Pid(),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	trace.SetTracerProvider(tp)

	trace.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Cleanly shutdown and flush telemetry when the application exits.
	defer func(ctx context.Context) {
		// Do not make the application hang when it is shutdown.
		ctx, cancel = context.WithTimeout(ctx, time.Second*10)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}(ctx)

	tr := tp.Tracer("example-main")

	spx := tr.Open(ctx, "example()", trace.WithNewRoot())
	defer spx.End()

	spx.Info("example starting")

	// subroutine
	var bar = func(carrier propagation.TextMapCarrier, arg string) {
		bartp, err := trace.JaegerProvider("http://localhost:14268/api/traces",
			trace.ServiceName("trace-demo-outside-boundary"),
			trace.Environment("go-test"),
			trace.Pid(),
		)
		if err != nil {
			log.Fatal(err)
		}

		bartpctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Cleanly shutdown and flush telemetry when the application exits.
		defer func(ctx context.Context) {
			// Do not make the application hang when it is shutdown.
			ctx, cancel = context.WithTimeout(ctx, time.Second*5)
			defer cancel()
			if err := bartp.Shutdown(ctx); err != nil {
				log.Fatal(err)
			}
		}(bartpctx)

		propagator := trace.GetTextMapPropagator()

		bartr := bartp.Tracer("example-bar")
		barctx := context.Background() // can use nil instead of context.Background()
		barspx := bartr.ExtractWithPropagator(barctx, propagator, carrier, "bar()")
		defer barspx.End()

		barspx.Argv(arg)
		barspx.Reply(trace.PASS, "OK")
	}

	carrier := make(propagation.MapCarrier)
	propagator := trace.GetTextMapPropagator()
	spx.Inject(propagator, carrier)

	// call subroutine
	bar(carrier, "foo")
	// Output:
}
