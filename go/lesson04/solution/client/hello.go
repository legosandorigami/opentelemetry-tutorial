package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"

	xhttp "github.com/legosandorigami/opentelemetry-tutorial/lib/http"
	"github.com/legosandorigami/opentelemetry-tutorial/lib/tracing"
)

func main() {
	// checking if the number of command-line arguments is exactly 3 (program name and two arguments)
	if len(os.Args) != 3 {
		panic("ERROR: Expecting two arguments")
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
	greeting := os.Args[2]

	// starting a new span named "say-hello" creating a span with the context that contains the baggage just created above
	ctx, span := tracer.Start(ctx, "say-hello")
	span.SetAttributes(attribute.String("hello-to", helloTo))
	defer span.End()

	// creating baggage items map and add "greeting"
	baggageItems := map[string]string{"greeting": greeting}

	// calling `formatString` function with the context ctx.
	helloStr, err := formatString(ctx, helloTo, baggageItems)
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

func formatString(ctx context.Context, helloTo string, baggageItems map[string]string) (string, error) {
	// retreiving a tracer named "say-hello-tracer" from the tracer provider
	tracer := otel.Tracer("say-hello-tracer")

	// preparing to send an http get request to the "formatter" service
	v := url.Values{}
	v.Set("helloTo", helloTo)
	url := "http://localhost:8081/format?" + v.Encode()

	// creating baggage members from the baggage items
	baggageMembers := make([]baggage.Member, 0)
	for k, v := range baggageItems {
		bm, err := baggage.NewMember(k, v)
		if err != nil {
			return "", fmt.Errorf("failed to create a new baggage member: %v", err)
		}
		baggageMembers = append(baggageMembers, bm)
	}

	// creating a baggage containing the baggage members
	b, err := baggage.New(baggageMembers...)
	if err != nil {
		return "", fmt.Errorf("failed to create a new baggage: %v", err)
	}

	// adding baggage to the context ctx
	ctx = baggage.ContextWithBaggage(ctx, b)

	// creating a span with the context ctx that contains the baggage, and custom attributes indicating that it is an RPC
	ctx, span := tracer.Start(ctx, "formatString",
		trace.WithAttributes(
			semconv.NetPeerNameKey.String(url),
			semconv.HTTPMethodKey.String("GET"),
		),
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	// creating a new HTTP request to formatter microservice
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	// retrieving the propagator and injecting the span context into the request headers
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))

	// Uncomment the line below to see the injected baggage and trace ID in the request headers
	// fmt.Println(req.Header)

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

	// preparing to send an http get request to the "publisher" service
	v := url.Values{}
	v.Set("helloStr", helloStr)
	url := "http://localhost:8082/publish?" + v.Encode()

	// creating a new HTTP request to printer microservice
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	// creating a span with custom attributes
	ctx, span := tracer.Start(ctx, "printHello",
		trace.WithAttributes(
			semconv.NetPeerNameKey.String(url),
			semconv.HTTPMethodKey.String("GET"),
		),
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	// retrieving the propagator and injecting the span context into the request headers
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))

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
