use std::collections::HashMap;
use axum::{
    extract::Query,
    http::HeaderMap,
    routing::get,
    Router,
};
use tokio::net::TcpListener;

use opentelemetry::{
    baggage::BaggageExt, global, trace::{Span, Tracer}, KeyValue
};
use opentelemetry_http::HeaderExtractor;

use exercise::init_tracer;

async fn format_handler(Query(params): Query<HashMap<String, String>>,  headers: HeaderMap) -> String {
    
    // creating a named instance of Tracer via the configured GlobalTracerProvider
    let tracer = global::tracer("formatter-tracer");
    
    // extracting the span context from the request headers
    let cx = global::get_text_map_propagator(|propagator| {
        propagator.extract(&HeaderExtractor(&headers))
    });

    // extracting baggage from the context
    let baggage = cx.baggage();

    let mut greeting = "Hello".to_string();

    // retrieving the member from the baggage with the key "greeting"
	if let Some(greeting_) = baggage.get("greeting"){
        greeting = greeting_.to_string();
    };
    
    // starting a new span named "format" as a child of the extracted span context
    let mut span = tracer.start_with_context("format", &cx);

    let mut resp= format!("{} there!", greeting);

    if let Some(hello_to) = params.get("hello_to") {
        resp = format!("{}, {}!", greeting, hello_to);
    }

    // adding an event to the span indicating that the string was properly formatted
    span.add_event("format-event-response", 
    vec![
        KeyValue::new("format-response", format!("string-formated: {}", resp)),
        ]
    );
    
    Span::end(&mut span);

    resp
}



#[tokio::main]
async fn main() {
    // initializing the OpenTelemetry TracerProvider with the service name "formatter"
    let tp = init_tracer("formatter")
        .expect("Error initializing tracer");

    let app = Router::new().route("/format", get(format_handler));
    
    let listener = TcpListener::bind("0.0.0.0:8081").await.unwrap();
    
    axum::serve(listener, app).await.unwrap();

    // shutting down the tracer provider to ensure all spans are flushed.
    tp.shutdown().expect("TracerProvider should shutdown successfully");
}