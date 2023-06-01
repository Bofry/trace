package trace_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Bofry/trace"
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

	sp := tr.Open(ctx, "example()")
	defer sp.End()

	sp.Info("example starting")
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

	sp := tr.Open(ctx, "example()")
	defer sp.End()

	sp.Info("example starting")

	// subroutine
	var bar = func(ctx context.Context, arg string) {
		barsp := tr.Start(ctx, "bar()")
		defer barsp.End()

		barsp.Argv(arg)
		barsp.Reply(trace.PASS, "OK")
	}

	ok := trace.IsNoopSeveritySpan(sp)
	sp.Debug("check IsNoopSeveritySpan()").Tags(
		trace.Key("IsNoopSeveritySpan").Bool(ok),
	)

	sp.Debug("log an error").Error(fmt.Errorf("some error"))

	// call subroutine
	bar(sp.Context(), "foo")
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

	sp := tr.Open(ctx, "example()")
	defer sp.End()

	sp.Info("example starting")

	// subroutine
	var bar = func(link trace.Link, arg string) {
		barctx := context.Background() // can use nil instead of context.Background()
		barsp := tr.Link(barctx, link, "bar()")
		defer barsp.End()

		barsp.Argv(arg)
		barsp.Err(fmt.Errorf("an error occurred"))
	}

	// call subroutine
	bar(sp.Link(), "foo")
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

	sp := tr.Open(ctx, "example()")
	defer sp.End()

	sp.Info("example starting")

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
		barsp := bartr.ExtractWithPropagator(barctx, propagator, carrier, "bar()")
		defer barsp.End()

		barsp.Argv(arg)
		barsp.Reply(trace.PASS, "OK")
	}

	/* NOTE: the following statements also can be written as
	 *
	 *  way 1:
	 *    carrier := make(propagation.MapCarrier)
	 *    propagator := trace.GetTextMapPropagator()
	 *    propagator.Inject(sp.Context(), carrier)
	 *
	 *  way 2:
	 *    carrier := make(propagation.MapCarrier)
	 *    propagator := trace.GetTextMapPropagator()
	 *    tr.InjectWithPropagator(sp.Context(), propagator, carrier)
	 *
	 *  way 3:
	 *    carrier := make(propagation.MapCarrier)
	 *    tr.Inject(sp.Context(), carrier)
	 */
	carrier := make(propagation.MapCarrier)
	propagator := trace.GetTextMapPropagator()
	sp.Inject(propagator, carrier)

	// call subroutine
	bar(carrier, "foo")
	// Output:
}
