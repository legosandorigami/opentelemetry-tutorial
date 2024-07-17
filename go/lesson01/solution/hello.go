package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/legosandorigami/opentelemetry-tutorial/lib/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	if len(os.Args) != 2 {
		panic("ERROR: Expecting one argument")
	}

	// initializing the OpenTelemetry TracerProvider with the service name "hello-world"
	tracerProvider, err := tracing.InitTracerProvider("hello-world")
	if err != nil {
		log.Fatalf("failed to create otel exporter: %v", err)
	}

	// creating a context and defering the shutdown of the TracerProvider to ensure proper cleanup
	ctx := context.Background()
	defer func() {
		if err := tracerProvider.Shutdown(ctx); err != nil {
			log.Fatalf("failed to shutdown TracerProvider: %v", err)
		}
	}()

	// getting a tracer from the tracer provider
	tracer := tracerProvider.Tracer("say-hello-tracer")

	helloTo := os.Args[1]

	// starting a new span named "say-hello"
	ctx, span := tracer.Start(ctx, "say-hello")

	// adding an attribute to the span
	span.SetAttributes(attribute.String("hello-to", helloTo))

	helloStr := fmt.Sprintf("Hello, %s!", helloTo)

	// adding logs to the span
	span.AddEvent("event", trace.WithAttributes(attribute.String("println", fmt.Sprintf("string-format: %s", helloStr))),
		trace.WithTimestamp(time.Now()),
	)

	println(helloStr)

	tracing.PrintSpanContents(span)

	// ending the span
	span.End()
}
