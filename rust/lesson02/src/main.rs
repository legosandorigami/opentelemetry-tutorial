use opentelemetry::{
    global, trace::{Span, TraceContextExt, TraceError, Tracer}, KeyValue
};
use opentelemetry_sdk::{
    propagation::TraceContextPropagator, trace::SdkTracerProvider, Resource
};
use opentelemetry_otlp::WithExportConfig;
use std::env;
use std::time::SystemTime;



const TRACING_BACKEND: &str = "http://192.168.50.4:4318/v1/traces";

/// initializes the OpenTelemetry Tracer with the specified service name and default backend.
fn init_tracer(service: &str) -> Result<SdkTracerProvider, TraceError>{
    init_tracer_with_backend(service, TRACING_BACKEND)
}

/// initializes the OpenTelemetry Tracer with the specified service name and backend.
fn init_tracer_with_backend(service: &str, backend: &str) -> Result<SdkTracerProvider, TraceError>{

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
    

    // Set up a propagator to handle context propagation across services.
    global::set_text_map_propagator(TraceContextPropagator::new());

    Ok(tp)
}

fn format_string(hello_to: &str)-> String {
	// Retrieve or create a named tracer.
	let tracer = global::tracer("say-hello-tracer");

	// Start a new span named "formatString".
    let mut span =  tracer.start("formatString");

    // doing the work
	let hello_str = format!("Hello, {}!", hello_to);

	// adding an event to the span.
    span.add_event_with_timestamp(
        "event",
        SystemTime::now(),
        vec![KeyValue::new("string-format", hello_str.clone())],
    );

    span.end();

	return hello_str
}

fn print_hello(hello_str: String) {
	// Retrieve or create a named tracer
    let tracer = global::tracer("say-hello-tracer");

	// Start a new span named "printHello"
    let mut span =  tracer.start("printHello");

    // doing the work
	println!("{}", hello_str);

	// adding an event to the span.
    span.add_event_with_timestamp(
        "event",
        SystemTime::now(),
        vec![KeyValue::new("println", hello_str)],
    );

    span.end();
}

#[tokio::main]
async fn main() {
    let args: Vec<_> = env::args().collect();
    if args.len() != 2 {
        panic!("ERROR: Expecting one argument");
    }

    let hello_to = args[1].clone();

    // initializing the OpenTelemetry Tracer with the service name "hello-world"
    let tp = init_tracer("hello-world")
        .expect("Error initializing tracer");

    // creating a named instance of Tracer via the configured GlobalTracerProvider
    let tracer = global::tracer("say-hello-tracer");

    // creating a new span named "say-hello".
    tracer.in_span("say-hello", |cx|{
        let span = cx.span();
        // adding an attribute to the span
        span.set_attribute(KeyValue::new("hello-to", hello_to.to_string()));

        // calling `format_string` function.
        let hello_str = format_string(hello_to.as_str());

        // calling `print_hello` function.
	    print_hello(hello_str);
    });    


    // shutting down the tracer provider to ensure all spans are flushed.
    tp.shutdown().expect("TracerProvider should shutdown successfully");
}


// #[tokio::main]
// async fn main() {
//     let args: Vec<_> = env::args().collect();
//     if args.len() != 2 {
//         panic!("ERROR: Expecting one argument");
//     }

//     let hello_to = args[1].clone();

//     // initializing the OpenTelemetry Tracer with the service name "hello-world"
//     let tp = init_tracer("hello-world")
//         .expect("Error initializing tracer");

//     // creating a named instance of Tracer via the configured GlobalTracerProvider
//     let tracer = global::tracer("say-hello-tracer");

//     // creating a new span named "say-hello".
//     let mut span =  tracer.start("say-hello");

//     // adding an attribute to the span
//     span.set_attribute(KeyValue::new("hello-to", hello_to.to_string()));

//     // calling `format_string` function.
//     let hello_str = format_string(hello_to.as_str());

//     // calling `print_hello` function.
// 	print_hello(hello_str);
    
//     // println!("{:#?}", span);
//     span.end();

//     // shutting down the tracer provider to ensure all spans are flushed.
//     tp.shutdown().expect("TracerProvider should shutdown successfully");
// }