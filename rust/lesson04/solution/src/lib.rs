// use tracing::instrument;
use opentelemetry::{
    global, trace::TraceError, KeyValue
};
use opentelemetry_sdk::{
    propagation::{TraceContextPropagator, BaggagePropagator}, trace::SdkTracerProvider, Resource
};
use opentelemetry::propagation::TextMapCompositePropagator;
use opentelemetry_otlp::WithExportConfig;



const TRACING_BACKEND: &str = "http://192.168.50.4:4318/v1/traces";

/// initializes the OpenTelemetry Tracer with the specified service name and default backend.
pub fn init_tracer(service: &str) -> Result<SdkTracerProvider, TraceError>{
    init_tracer_with_backend(service, TRACING_BACKEND)
}

/// initializes the OpenTelemetry Tracer with the specified service name and backend.
pub fn init_tracer_with_backend(service: &str, backend: &str) -> Result<SdkTracerProvider, TraceError>{

    // creating an OTLP trace exporter to send spans using HTTP to the specified backend
    let exporter = opentelemetry_otlp::SpanExporter::builder().with_http().with_endpoint(backend).build()?;

    // defining resource attributes for the service
    let resource = Resource::builder_empty()
        .with_attributes([
                // service name
                KeyValue::new("service.name", service.to_string()),
                // version number of the environment
                KeyValue::new("service.version", "1.0.0".to_string()),
                // environment
                KeyValue::new("environment", "production".to_string()),
            ])
        .build();

    // creating a TracerProvider with the specified exporter and resource attributes
    let tp = SdkTracerProvider::builder()
        .with_batch_exporter(exporter)
        .with_resource(resource)
        .build();

    // setting up the global tracer provider
    global::set_tracer_provider(tp.clone());

    // create an instance of `TraceContextPropagator`(for propagating the traces) and an instance of `BaggagePropagator`(for propagating the baggage).
    let baggage_propagator = BaggagePropagator::new();
    let trace_context_propagator = TraceContextPropagator::new();

    // create a composite propagator that contains both the `baggage_propagator` and `trace_context_propagator`.
    let composite_propagator = TextMapCompositePropagator::new(vec![
        Box::new(baggage_propagator),
        Box::new(trace_context_propagator),
    ]);
    

    // Set up a propagator to handle the propagation of both baggage and context across services.
    global::set_text_map_propagator(composite_propagator);

    Ok(tp)
}