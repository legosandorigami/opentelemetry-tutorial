# Lesson 1 - Hello World

## Objectives

Learn how to:

* Instantiate a Tracer
* Create a simple trace
* Add attributes to a span

## Walkthrough

### A simple Hello-World program

Let's create a simple Go program `lesson01/exercise/hello.go` that takes an argument and prints "Hello, {arg}!".

```go
package main

import (
    "fmt"
    "os"
)

func main() {
    if len(os.Args) != 2 {
        panic("ERROR: Expecting one argument")
    }
    helloTo := os.Args[1]
    helloStr := fmt.Sprintf("Hello, %s!", helloTo)
    println(helloStr)
}
```

Run it:
``` bash
$ go run ./lesson01/exercise/hello.go Bryan
Hello, Bryan!
```

### Create a trace

A trace is a directed acyclic graph of spans. A span is a logical representation of some work done in your application. Each span has these minimum attributes: an operation name, a start time, and a finish time.

Let's create a trace that consists of just a single span. To do that we need an instance of the `trace.TracerProvider`. We can use a global instance returned by `otel.GetTracerProvider()`.

```go
tracerProvider := otel.GetTracerProvider()

// creating a tracer from the tracer provider just created
tracer := tracerProvider.Tracer("say-hello-tracer")

// starting a new span named "say-hello"
_, span := tracer.Start(context.Background(), "say-hello")
println(helloStr)

// ending the span
span.End()
```

We are using the following basic features of the OpenTelemetry API:
  * a `tracerProvider` A tracer provider creates tracers which can then be used to create traces and spans.
  * a `tracer` instance is used to start new spans via `Start` function
  * each `span` is given an _operation name_, `"say-hello"` in this case
  * each `span` must be ended by calling its `End` function
  * the start and end timestamps of the span will be captured automatically by the tracer implementation.

However, if we run this program, we will see no difference, and no traces in the tracing UI. That's because the function `otel.GetTracerProvider` returns a no-op tracer provider by default.

### Initialize a real tracer

Let's create an instance of a real tracer provider.

```go
import (
	"context"
	"errors"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	traceSdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	TRACING_BACKEND = "localhost:4318"
)

// InitTracerProvider initializes the OpenTelemetry TracerProvider with the specified service name and default backend.
func InitTracerProvider(servicename string) (*traceSdk.TracerProvider, error) {
	return InitTracerProviderWithBackend(servicename, TRACING_BACKEND)
}

// InitTracerProviderWithBackend initializes the OpenTelemetry TracerProvider with the specified service name and backend.
func InitTracerProviderWithBackend(service, backend string) (*traceSdk.TracerProvider, error) {
	ctx := context.Background()

	// creating an OTLP trace exporter to send spans to the specified backend
	exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpoint(backend), otlptracehttp.WithInsecure())
	if err != nil {
		return nil, err
	}

	// defining resource attributes for the service
	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			// adding service name
			semconv.ServiceNameKey.String(service),
			// adding version number of the application
			semconv.ServiceVersionKey.String("1.0.0"),
			// adding environment
			attribute.String("environment", "production"),
		),
	)
	if err != nil {
		return nil, err
	}

	// creating a TracerProvider with the specified exporter and resource attributes
	tp := traceSdk.NewTracerProvider(
		traceSdk.WithBatcher(exporter),
		traceSdk.WithResource(res),
	)

	// setting up the global tracer provider
	otel.SetTracerProvider(tp)

	return tp, nil
}
```

To use this instance, let's change the `main` function:

```go
// set and then retrieve the global tracer provider instance  
tracerProvider := InitTracerProvider("hello-world")

ctx := context.Background()

// shutting down the tracer to ensure proper cleanup 
defer func() {
    if err := tracerProvider.Shutdown(ctx); err != nil {
        log.Fatalf("failed to shutdown TracerProvider: %v", err)
    }
}()

// getting a tracer from the tracer provider
tracer := tracerProvider.Tracer("say-hello-tracer")

// starting a new span named "say-hello"
_, span := tracer.Start(ctx, "say-hello")
println(helloStr)

// printing span details
printSpanContents(span)

// ending the span
span.End()
```

Note that we are passing a string `hello-world` to the init method. It is used to mark all spans emitted by the tracer as originating from a `hello-world` service.

If we run the program now, we should see a span logged:

```bash
$ go run ./lesson01/exercise/hello.go Bryan
2017/09/22 20:26:49 Initializing logging reporter
Hello, Bryan!
2025/03/13 19:22:09 {"TraceID":"99eae68a2e21e3ba24fe85aea81a92f7","SpanID":"0788b6b2c56e3f34","TraceFlags":"01","TraceState":"","Remote":false}
```

If you have a backend(Jaeger, Tempo and Grafana or Signoz) running, you should be able to see the trace in the UI.

### Adding Attributes and Events to a span

Right now the trace we created is very basic. If we call our program with `hello.go Susan` instead of `hello.go Bryan`, the resulting traces will be nearly identical. It would be nice if we could capture the program arguments in the traces to distinguish them.

One naive way is to use the string `"Hello, Bryan!"` as the _operation name_ of the span, instead of `"say-hello"`. However, such practice is highly discouraged in distributed tracing, because the operation name is meant to represent a _class of spans_, rather than a unique instance. For example, in Jaeger UI you can select the operation name from a dropdown when searching for traces. It would be very bad user experience if we ran the program to say hello to a 1000 people and the dropdown then contained 1000 entries. Another reason for choosing more general operation names is to allow the tracing systems to do aggregations. For example, Jaeger tracer has an option of emitting metrics for all the traffic going through the application. Having a unique operation name for each span would make the metrics useless.

The recommended solution is to add attributes and events(logs) to spans. An _attribute_ is a key-value pair that provides certain metadata about the span. An _event_ is similar to a regular log statement, it contains a timestamp and some data, but it is associated with span from which it was logged.

When should we use attributes vs. events?  The attributes are meant to describe attributes of the span that apply to the whole duration of the span. For example, if a span represents an HTTP request, then the URL of the request should be recorded as an attribute because it does not make sense to think of the URL as something that's only relevant at different points in time on the span. On the other hand, if the server responded with a redirect URL, logging it would make more sense since there is a clear timestamp associated with such an event.

#### Adding Attributes

In the case of `hello.go Bryan`, the string "Bryan" is a good candidate for a span attribute, since it applies
to the whole span and not to a particular moment in time. We can record it like this:

```go
import "go.opentelemetry.io/otel/attribute"

span := tracer.Start("say-hello")
span.SetAttributes(attribute.String("hello-to", helloTo))
```

#### Adding Events

Our hello program is so simple that it's difficult to find a relevant example of a log, but let's try.
Right now we're formatting the `helloStr` and then printing it. Both of these operations take certain
time, so we can log their completion:

```go
import "github.com/opentracing/opentracing-go/log"

helloStr := fmt.Sprintf("Hello, %s!", helloTo)

// adding an event to the span.
span.AddEvent("event", trace.WithAttributes(attribute.String("println", fmt.Sprintf("string-format: %s", helloStr))), trace.WithTimestamp(time.Now()),)
```

The log statements might look a bit strange if you have not previously worked with a structured logging API. Rather than formatting a log message into a single string that is easy for humans to read, structured logging APIs encourage you to separate bits and pieces of that message into key-value pairs that can be automatically processed by log aggregation systems. The idea comes from the realization that today most logs are processed by machines rather than humans.

The OpenTelemetry API for Go provides structured logging through the use of events and attributes:
  * The `AddEvent` method allows you to add an event to a span, which can include a name and a set of attributes. Events are time-stamped and provide a way to log significant occurrences within the span’s lifetime.
  * Attributes can be added to events using the `WithAttributes` function, which takes a list of key-value pairs in the form of `attribute.KeyValue`.

The OpenTelemetry Specification also recommends that all events contain an `event` attribute that describes the overall event being logged, with other attributes of the event provided as additional fields.

If you run the program with these changes, then find the trace in the UI and expand its span (by clicking on it), you will be able to see the tags and logs.

## Conclusion

The complete program can be found in the [solution](./solution) package. We moved the `InitTracerProvider` helper function into its own package so that we can reuse it in the other lessons as `tracing.InitTracerProvider`.

Next lesson: [Context and Tracing Functions](../lesson02).