package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/legosandorigami/opentelemetry-tutorial/lib/tracing"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	// checking if the number of command-line arguments is exactly 2 (program name and one argument).
	if len(os.Args) != 2 {
		panic("ERROR: Expecting one argument")
	}

	// initializing the OpenTelemetry TracerProvider with the service name "hello-world"
	tracerPovider, err := tracing.InitTracerProvider("hello-world")
	if err != nil {
		log.Fatalf("failed to create otel exporter: %v", err)
	}

	// creating a context and defering the shutdown of the TracerProvider to ensure proper cleanup
	ctx := context.Background()
	defer func() {
		if err := tracerPovider.Shutdown(ctx); err != nil {
			log.Fatalf("failed to shutdown TracerProvider: %v", err)
		}
	}()

	// creating a tracer from the tracer provider named "say-hello-tracer"
	tracer := tracerPovider.Tracer("say-hello-tracer")

	helloTo := os.Args[1]

	// starting a new span named "say-hello"
	ctx, span := tracer.Start(ctx, "say-hello")
	span.SetAttributes(attribute.String("hello-to", helloTo))
	defer span.End()

	// calling `formatString` function with the context ctx.
	helloStr := formatString(ctx, helloTo)

	// calling `printHello` function with the context ctx.
	printHello(ctx, helloStr)

	tracing.PrintSpanContents(span)
}

func formatString(ctx context.Context, helloTo string) string {
	// Retrieve or create a named tracer.
	tracer := otel.Tracer("say-hello-tracer")

	// Start a new span named "formatString".
	_, span := tracer.Start(ctx, "formatString")
	defer span.End()

	helloStr := fmt.Sprintf("Hello, %s!", helloTo)

	// adding an event to the span.
	span.AddEvent("event",
		trace.WithAttributes(attribute.String("string-format", helloStr)), trace.WithTimestamp(time.Now()))

	// printing the span contents
	tracing.PrintSpanContents(span)

	return helloStr
}

func printHello(ctx context.Context, helloStr string) {
	// Retrieve or create a named tracer
	tracer := otel.Tracer("say-hello-tracer")

	// Start a new span named "printHello"
	_, span := tracer.Start(ctx, "printHello")
	defer span.End()

	println(helloStr)

	// adding an event to the span.
	span.AddEvent("event",
		trace.WithAttributes(attribute.String("println", helloStr)), trace.WithTimestamp(time.Now()))

	// printing the span contents
	tracing.PrintSpanContents(span)
}
