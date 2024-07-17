package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/legosandorigami/opentelemetry-tutorial/lib/tracing"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	// initialize the OpenTelemetry TracerProvider with the service name "formatter"
	tracerPovider, err := tracing.InitTracerProvider("formatter")
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

	// retrieving or creating a tracer with name "formatter-tracer"
	tracer := tracerPovider.Tracer("formatter-tracer")

	http.HandleFunc("/format", func(w http.ResponseWriter, r *http.Request) {

		// retrieving the global propagator
		propagator := otel.GetTextMapPropagator()

		// extracting the span context from the request headers
		ctx := propagator.Extract(context.Background(), propagation.HeaderCarrier(r.Header))

		// starting a new span named "format" as a child of the extracted span context
		_, span := tracer.Start(ctx, "format", trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		helloTo := r.FormValue("helloTo")
		helloStr := fmt.Sprintf("Hello, %s!", helloTo)

		// adding an event to the span indicating that the string was properly formatted
		span.AddEvent("event", trace.WithAttributes(
			attribute.String("string-format", helloStr),
		))

		// printing the span details
		tracing.PrintSpanContents(span)

		w.Write([]byte(helloStr))
	})

	log.Fatal(http.ListenAndServe(":8081", nil))
}
