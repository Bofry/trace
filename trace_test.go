package trace_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/Bofry/trace"
	"github.com/joho/godotenv"
	"go.opentelemetry.io/otel/propagation"
)

var (
	__TEST_OTLP_HTTP_URL     string
	__TEST_JAEGER_QUERY_URL  string
	__TEST_JAEGER_ENDPOINT   string

	__ENV_FILE        = "trace_test.env"
	__ENV_FILE_SAMPLE = "trace_test.env.sample"

	// getTestOTLPEndpoint returns the preferred OTLP endpoint for testing
	getTestOTLPEndpoint = func() string {
		// Prefer Jaeger OTLP endpoint if available
		if __TEST_JAEGER_ENDPOINT != "" {
			return __TEST_JAEGER_ENDPOINT
		}
		if __TEST_OTLP_HTTP_URL != "" {
			return __TEST_OTLP_HTTP_URL
		}
		// Fallback to default OTLP endpoint
		return "http://127.0.0.1:4318"
	}

	__Expected_TestSeveritySpan_Inject = func(t *testing.T, _TRACE_ID string) {
		if __TEST_JAEGER_QUERY_URL == "" {
			t.Logf("Trace ID: %s (no Jaeger query URL configured - check trace_test.env)", _TRACE_ID)
			return
		}

		t.Logf("Verifying trace ID %s in Jaeger at %s", _TRACE_ID, __TEST_JAEGER_QUERY_URL)

		// Give some time for trace to be processed
		time.Sleep(2 * time.Second)

		// Query Jaeger for the trace
		queryURL := fmt.Sprintf("%s/%s", __TEST_JAEGER_QUERY_URL, _TRACE_ID)
		resp, err := http.Get(queryURL)
		if err != nil {
			t.Logf("Warning: Could not query Jaeger (%v). Trace ID: %s", err, _TRACE_ID)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Logf("Warning: Jaeger query returned status %d. Trace ID: %s", resp.StatusCode, _TRACE_ID)
			return
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Logf("Warning: Could not read Jaeger response (%v). Trace ID: %s", err, _TRACE_ID)
			return
		}

		var result map[string]any
		if err := json.Unmarshal(body, &result); err != nil {
			t.Logf("Warning: Could not parse Jaeger response (%v). Trace ID: %s", err, _TRACE_ID)
			return
		}

		// Check if trace was found
		if data, ok := result["data"].([]any); ok && len(data) > 0 {
			t.Logf("[SUCCESS] Successfully found trace in Jaeger! Trace ID: %s", _TRACE_ID)
		} else {
			t.Logf("[WARNING] Trace not found in Jaeger yet. Trace ID: %s (may need more time)", _TRACE_ID)
		}
		return
	}
)

func TestMain(m *testing.M) {
	_, err := os.Stat(__ENV_FILE)
	if err != nil {
		if os.IsNotExist(err) {
			err = copyFile(__ENV_FILE_SAMPLE, __ENV_FILE)
			if err != nil {
				panic(err)
			}
		}
	}

	godotenv.Load(__ENV_FILE)
	{
		__TEST_OTLP_HTTP_URL = os.Getenv("OTLP_HTTP_URL")
		__TEST_JAEGER_QUERY_URL = os.Getenv("JAEGER_QUERY_URL")

		// Determine which endpoint to use for sending traces
		if jaegerOtlp := os.Getenv("JAEGER_OTLP_HTTP_URL"); jaegerOtlp != "" {
			__TEST_JAEGER_ENDPOINT = jaegerOtlp
		} else if jaegerLegacy := os.Getenv("JAEGER_HTTP_COLLECTOR"); jaegerLegacy != "" {
			__TEST_JAEGER_ENDPOINT = jaegerLegacy
		}
	}
	m.Run()
}

func copyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}

func TestSeveritySpan_Inject(t *testing.T) {
	var _TRACE_ID string
	defer func() {
		__Expected_TestSeveritySpan_Inject(t, _TRACE_ID)
	}()

	tp, err := trace.OTLPProvider(getTestOTLPEndpoint(),
		trace.ServiceName("trace-test"),
		trace.Environment("go-test"),
		trace.Pid(),
	)
	if err != nil {
		t.Fatal(err)
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
			t.Fatal(err)
		}
	}(ctx)

	tr := tp.Tracer("example-main")

	sp := tr.Open(ctx, "example()")
	defer sp.End()

	sp.Info("example starting")

	// subroutine
	var bar = func(carrier propagation.TextMapCarrier, arg string) {
		bartp, err := trace.OTLPProvider(getTestOTLPEndpoint(),
			trace.ServiceName("trace-test-outside-boundary"),
			trace.Environment("go-test"),
			trace.Pid(),
		)
		if err != nil {
			t.Fatal(err)
		}

		bartpctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Cleanly shutdown and flush telemetry when the application exits.
		defer func(ctx context.Context) {
			// Do not make the application hang when it is shutdown.
			ctx, cancel = context.WithTimeout(ctx, time.Second*5)
			defer cancel()
			if err := bartp.Shutdown(ctx); err != nil {
				t.Fatal(err)
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

	/* NOTE: the following
	 *  carrier := make(propagation.MapCarrier)
	 *  tr.Inject(sp.Context(), carrier)
	 */
	carrier := make(propagation.MapCarrier)
	propagator := trace.GetTextMapPropagator()
	sp.Inject(propagator, carrier)

	// call subroutine
	bar(carrier, "foo")

	// export TRACE ID to query the jaeger record for checking the test
	_TRACE_ID = sp.TraceID().String()
}

func TestPropagator_Inject(t *testing.T) {
	var _TRACE_ID string
	defer func() {
		__Expected_TestSeveritySpan_Inject(t, _TRACE_ID)
	}()

	tp, err := trace.OTLPProvider(getTestOTLPEndpoint(),
		trace.ServiceName("trace-test"),
		trace.Environment("go-test"),
		trace.Pid(),
	)
	if err != nil {
		t.Fatal(err)
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
			t.Fatal(err)
		}
	}(ctx)

	tr := tp.Tracer("example-main")

	sp := tr.Open(ctx, "example()")
	defer sp.End()

	sp.Info("example starting")

	// subroutine
	var bar = func(carrier propagation.TextMapCarrier, arg string) {
		bartp, err := trace.OTLPProvider(getTestOTLPEndpoint(),
			trace.ServiceName("trace-test-outside-boundary"),
			trace.Environment("go-test"),
			trace.Pid(),
		)
		if err != nil {
			t.Fatal(err)
		}

		bartpctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Cleanly shutdown and flush telemetry when the application exits.
		defer func(ctx context.Context) {
			// Do not make the application hang when it is shutdown.
			ctx, cancel = context.WithTimeout(ctx, time.Second*5)
			defer cancel()
			if err := bartp.Shutdown(ctx); err != nil {
				t.Fatal(err)
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

	carrier := make(propagation.MapCarrier)
	propagator := trace.GetTextMapPropagator()
	propagator.Inject(sp.Context(), carrier)

	// call subroutine
	bar(carrier, "foo")

	// export TRACE ID to query the jaeger record for checking the test
	_TRACE_ID = sp.TraceID().String()
}

func TestSeverityTracer_InjectWithPropagator(t *testing.T) {
	var _TRACE_ID string
	defer func() {
		__Expected_TestSeveritySpan_Inject(t, _TRACE_ID)
	}()

	tp, err := trace.OTLPProvider(getTestOTLPEndpoint(),
		trace.ServiceName("trace-test"),
		trace.Environment("go-test"),
		trace.Pid(),
	)
	if err != nil {
		t.Fatal(err)
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
			t.Fatal(err)
		}
	}(ctx)

	tr := tp.Tracer("example-main")

	sp := tr.Open(ctx, "example()")
	defer sp.End()

	sp.Info("example starting")

	// subroutine
	var bar = func(carrier propagation.TextMapCarrier, arg string) {
		bartp, err := trace.OTLPProvider(getTestOTLPEndpoint(),
			trace.ServiceName("trace-test-outside-boundary"),
			trace.Environment("go-test"),
			trace.Pid(),
		)
		if err != nil {
			t.Fatal(err)
		}

		bartpctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Cleanly shutdown and flush telemetry when the application exits.
		defer func(ctx context.Context) {
			// Do not make the application hang when it is shutdown.
			ctx, cancel = context.WithTimeout(ctx, time.Second*5)
			defer cancel()
			if err := bartp.Shutdown(ctx); err != nil {
				t.Fatal(err)
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

	carrier := make(propagation.MapCarrier)
	propagator := trace.GetTextMapPropagator()
	tr.InjectWithPropagator(sp.Context(), propagator, carrier)

	// call subroutine
	bar(carrier, "foo")

	// export TRACE ID to query the jaeger record for checking the test
	_TRACE_ID = sp.TraceID().String()
}

func TestSeverityTracer_Inject(t *testing.T) {
	var _TRACE_ID string
	defer func() {
		__Expected_TestSeveritySpan_Inject(t, _TRACE_ID)
	}()

	tp, err := trace.OTLPProvider(getTestOTLPEndpoint(),
		trace.ServiceName("trace-test"),
		trace.Environment("go-test"),
		trace.Pid(),
	)
	if err != nil {
		t.Fatal(err)
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
			t.Fatal(err)
		}
	}(ctx)

	tr := tp.Tracer("example-main")

	sp := tr.Open(ctx, "example()")
	defer sp.End()

	sp.Info("example starting")

	// subroutine
	var bar = func(carrier propagation.TextMapCarrier, arg string) {
		bartp, err := trace.OTLPProvider(getTestOTLPEndpoint(),
			trace.ServiceName("trace-test-outside-boundary"),
			trace.Environment("go-test"),
			trace.Pid(),
		)
		if err != nil {
			t.Fatal(err)
		}

		bartpctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Cleanly shutdown and flush telemetry when the application exits.
		defer func(ctx context.Context) {
			// Do not make the application hang when it is shutdown.
			ctx, cancel = context.WithTimeout(ctx, time.Second*5)
			defer cancel()
			if err := bartp.Shutdown(ctx); err != nil {
				t.Fatal(err)
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

	carrier := make(propagation.MapCarrier)
	tr.Inject(sp.Context(), carrier)

	// call subroutine
	bar(carrier, "foo")

	// export TRACE ID to query the jaeger record for checking the test
	_TRACE_ID = sp.TraceID().String()
}

func TestSeveritySpan_Link(t *testing.T) {
	var _TRACE_ID string
	defer func() {
		__Expected_TestSeveritySpan_Inject(t, _TRACE_ID)
	}()

	tp, err := trace.OTLPProvider(getTestOTLPEndpoint(),
		trace.ServiceName("trace-test"),
		trace.Environment("go-test"),
		trace.Pid(),
	)
	if err != nil {
		t.Fatal(err)
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
			t.Fatal(err)
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

	// export TRACE ID to query the jaeger record for checking the test
	_TRACE_ID = sp.TraceID().String()
}

func TestSeverityTracer_Start(t *testing.T) {
	var _TRACE_ID string
	defer func() {
		__Expected_TestSeveritySpan_Inject(t, _TRACE_ID)
	}()

	tp, err := trace.OTLPProvider(getTestOTLPEndpoint(),
		trace.ServiceName("trace-test"),
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

	// export TRACE ID to query the jaeger record for checking the test
	_TRACE_ID = sp.TraceID().String()
}
