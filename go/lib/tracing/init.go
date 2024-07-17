package tracing

import (
	"context"
	"log"

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
			semconv.ServiceNameKey.String(service),        // service name
			semconv.ServiceVersionKey.String("1.0.0"),     // version number of the application
			attribute.String("environment", "production"), // environment
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

	// setting up a propagator to handle trace context propagation across the services
	// otel.SetTextMapPropagator(propagation.TraceContext{})

	// Uncomment the code below to set up composite propagator
	// setting up a composite propagator to handle context propagation (traces and baggage) across services
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp, nil
}

// prints the span contents
func PrintSpanContents(span trace.Span) {
	spanCtx := span.SpanContext()

	data, err := spanCtx.MarshalJSON()
	if err != nil {
		return
	}

	log.Printf("%v\n", string(data))
}
