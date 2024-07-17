package main

import (
	"context"
	"log"
	"net/http"

	"github.com/legosandorigami/opentelemetry-tutorial/lib/tracing"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func main() {
	// initialize the OpenTelemetry TracerProvider with the service name "publisher"
	tracerPovider, err := tracing.InitTracerProvider("publisher")
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

	// retrieving or creating a tracer with name "publisher-tracer"
	tracer := tracerPovider.Tracer("publisher-tracer")

	http.HandleFunc("/publish", func(w http.ResponseWriter, r *http.Request) {
		// retrieve the global propagator
		propagator := otel.GetTextMapPropagator()

		// extracting the span context from the request headers
		ctx := propagator.Extract(context.Background(), propagation.HeaderCarrier(r.Header))

		// starting a new span named "publish" as a child of the extracted span context
		_, span := tracer.Start(ctx, "publish")
		defer span.End()

		helloStr := r.FormValue("helloStr")
		println(helloStr)

		// printing the span details
		tracing.PrintSpanContents(span)
	})

	log.Fatal(http.ListenAndServe(":8082", nil))
}
