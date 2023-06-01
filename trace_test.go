package trace_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/Bofry/trace"
	"github.com/joho/godotenv"
	"go.opentelemetry.io/otel/propagation"
)

var (
	__TEST_JAEGER_TRACE_URL string
	__TEST_JAEGER_QUERY_URL string

	__Expected_TestSeveritySpan_Inject = func(t *testing.T, _TRACE_ID string) {
		query_api_url, err := url.JoinPath(__TEST_JAEGER_QUERY_URL, _TRACE_ID)
		if err != nil {
			t.Fatal(err)
		}
		client := &http.Client{}
		req, err := http.NewRequest("GET", query_api_url, nil)
		if err != nil {
			t.Fatal(err)
		}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		/* {
		 *	 "data": [
		 *	   {
		 *       "traceID": "xxxxxx",
		 *       "spans"  : [
		 *	       { ... },
		 *	       ...
		 *       ],
		 *       "processes": {},
		 *       "warnings" : ...
		 *     }
		 *   ],
		 *   "total"  : 0,
		 *   "limit"  : 0,
		 *   "offset" : 0,
		 *   "errors" : ...
		 * }
		 */
		var record map[string]interface{}
		err = json.Unmarshal(body, &record)
		if err != nil {
			t.Fatal(err)
		}

		{
			v, ok := record["data"]
			if !ok {
				t.Errorf("???")
			}
			record_data, ok := v.([]interface{})
			if !ok {
				t.Errorf("???")
			}
			var exceptedRecordDataLength int = 1
			if exceptedRecordDataLength != len(record_data) {
				t.Errorf("data size expect %v, got %v", exceptedRecordDataLength, len(record_data))
			}
			{
				record_data_0, ok := record_data[0].(map[string]interface{})
				if !ok {
					t.Errorf("???")
				}
				record_data_0_traceID, ok := record_data_0["traceID"].(string)
				if !ok {
					t.Errorf("???")
				}
				var expectedRecordData_0_TraceID string = _TRACE_ID
				if expectedRecordData_0_TraceID != record_data_0_traceID {
					t.Errorf("data[0].traceID expect %v, got %v", expectedRecordData_0_TraceID, record_data_0_traceID)
				}
				record_data_0_spans, ok := record_data_0["spans"].([]interface{})
				if !ok {
					t.Errorf("???")
				}
				var expectedRecordData_0_SpansLength int = 2
				if expectedRecordData_0_SpansLength != len(record_data_0_spans) {
					t.Errorf("data[0].spans length expect %v, got %v", expectedRecordData_0_SpansLength, len(record_data_0_spans))
				}
			}
		}
	}
)

func TestMain(m *testing.M) {
	godotenv.Load()
	{
		__TEST_JAEGER_TRACE_URL = os.Getenv("JAEGER_TRACE_URL")
		__TEST_JAEGER_QUERY_URL = os.Getenv("JAEGER_QUERY_URL")
	}
	m.Run()
}

func TestSeveritySpan_Inject(t *testing.T) {
	var _TRACE_ID string
	defer func() {
		__Expected_TestSeveritySpan_Inject(t, _TRACE_ID)
	}()

	tp, err := trace.JaegerProvider(__TEST_JAEGER_TRACE_URL,
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
		bartp, err := trace.JaegerProvider(__TEST_JAEGER_TRACE_URL,
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

	tp, err := trace.JaegerProvider(__TEST_JAEGER_TRACE_URL,
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
		bartp, err := trace.JaegerProvider(__TEST_JAEGER_TRACE_URL,
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

	tp, err := trace.JaegerProvider(__TEST_JAEGER_TRACE_URL,
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
		bartp, err := trace.JaegerProvider(__TEST_JAEGER_TRACE_URL,
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

	tp, err := trace.JaegerProvider(__TEST_JAEGER_TRACE_URL,
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
		bartp, err := trace.JaegerProvider(__TEST_JAEGER_TRACE_URL,
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

	tp, err := trace.JaegerProvider(__TEST_JAEGER_TRACE_URL,
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

	tp, err := trace.JaegerProvider(__TEST_JAEGER_TRACE_URL,
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
