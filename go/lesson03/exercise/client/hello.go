package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	xhttp "github.com/legosandorigami/opentelemetry-tutorial/lib/http"
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

	// getting a tracer from the tracer provider named "say-hello-tracer"
	tracer := tracerPovider.Tracer("say-hello-tracer")

	helloTo := os.Args[1]

	// starting a new span named "say-hello"
	ctx, span := tracer.Start(ctx, "say-hello")
	span.SetAttributes(attribute.String("hello-to", helloTo))
	defer span.End()

	// calling `formatString` function with the context ctx.
	helloStr, err := formatString(ctx, helloTo)
	if err != nil {
		log.Fatalf(err.Error())
	}

	// calling `printHello` function with the context ctx.
	err = printHello(ctx, helloStr)
	if err != nil {
		log.Fatalf(err.Error())
	}

	// printing the span details
	tracing.PrintSpanContents(span)
}

func formatString(ctx context.Context, helloTo string) (string, error) {
	// retreiving a tracer from the tracer provider
	tracer := otel.Tracer("say-hello-tracer")

	// preparing to send an http get request to the "formatter" service
	v := url.Values{}
	v.Set("helloTo", helloTo)
	url := "http://localhost:8081/format?" + v.Encode()

	// starting a new span named "formatString"
	_, span := tracer.Start(ctx, "formatString")
	defer span.End()

	// creating a new HTTP request to formatter microservice
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		// recording the error in the span
		span.RecordError(err, trace.WithAttributes(
			attribute.String("format-request-error", fmt.Sprintf("Failed to create a request to  the `formatter` service for the string %s", helloTo))))
		return "", err
	}

	//sending a get request
	resp, err := xhttp.Do(req)
	if err != nil {
		// recording the error in the span
		span.RecordError(err, trace.WithAttributes(
			attribute.String("format-response-error", fmt.Sprintf("Failed to format the string %s", helloTo))))
		return "", err
	}

	helloStr := string(resp)

	// adding an event to the span indicating a successful response was received
	span.AddEvent("format-event-response", trace.WithAttributes(
		attribute.String("format-response", fmt.Sprintf("string-format: %s", helloStr)),
	))

	// printing the span details
	tracing.PrintSpanContents(span)

	return helloStr, nil
}

func printHello(ctx context.Context, helloStr string) error {
	// retreiving a tracer from the tracer provider
	tracer := otel.Tracer("say-hello-tracer")

	// starting a new span named "printHello"
	_, span := tracer.Start(ctx, "printHello")
	defer span.End()

	// preparing to send an http get request to the "publisher" service
	v := url.Values{}
	v.Set("helloStr", helloStr)
	url := "http://localhost:8082/publish?" + v.Encode()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		// recording the error in the span
		span.RecordError(err, trace.WithAttributes(
			attribute.String("publish-request-error", fmt.Sprintf("Failed to create a request to  the `publisher` service for the string %s", helloStr))))
		return err
	}

	//sending a get request
	if _, err := xhttp.Do(req); err != nil {
		// recording the error in the span
		span.RecordError(err, trace.WithAttributes(
			attribute.String("publish-response-error", fmt.Sprintf("Failed to publish the string %s", helloStr))))
		return err
	}

	// printing the span details
	tracing.PrintSpanContents(span)

	return nil
}
