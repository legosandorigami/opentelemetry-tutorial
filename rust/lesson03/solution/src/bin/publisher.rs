use std::collections::HashMap;
use axum::{
    extract::Query,
    http::HeaderMap,
    routing::get,
    Router,
};
use tokio::net::TcpListener;

use opentelemetry::{
    global, trace::{Span, Tracer},
};
use opentelemetry_http::HeaderExtractor;

use exercise::init_tracer;


async fn publish_handler(Query(params): Query<HashMap<String, String>>,  headers: HeaderMap) {
    // creating a named instance of Tracer via the configured GlobalTracerProvider
    let tracer = global::tracer("publisher-tracer");
    
    // extracting the span context from the request headers
    let cx = global::get_text_map_propagator(|propagator| {
        propagator.extract(&HeaderExtractor(&headers))
    });
    
    // starting a new span named "publish" as a child of the extracted span context
    let mut span = tracer.start_with_context("publish", &cx);

    if let Some(hello_str) = params.get("hello_str") {
        println!("{}", hello_str);
    }
    span.end();
}



#[tokio::main]
async fn main() {

    // initializing the OpenTelemetry TracerProvider with the service name "publisher"
    let tp = init_tracer("publisher")
        .expect("Error initializing tracer");

    let app = Router::new().route("/publish", get(publish_handler));
    
    let listener = TcpListener::bind("0.0.0.0:8082").await.unwrap();
    
    let _ = axum::serve(listener, app).await;
    
    // shutting down the tracer provider to ensure all spans are flushed.
    tp.shutdown().expect("TracerProvider should shutdown successfully");

}