use std::env;
use reqwest::Client;
use std::collections::HashMap;
use exercise::init_tracer;

use opentelemetry::{
    global, trace::{Span, TraceContextExt, Tracer, FutureExt}, KeyValue,
};

use opentelemetry_http::HeaderInjector;

const FORMAT_URL: &str = "http://localhost:8081/format";
const PUBLISH_URL: &str = "http://localhost:8082/publish";

#[tokio::main]
async fn main()-> Result<(), reqwest::Error>{
    // checking if the number of command-line arguments is exactly 2 (program name and one argument).
    let args: Vec<_> = env::args().collect();
    if args.len() != 2 {
        panic!("ERROR: Expecting one argument");
    }
    let hello_to = args[1].clone();

    // initializing the OpenTelemetry TracerProvider with the service name "hello-world"
    let tp = init_tracer("hello-world")
        .expect("Error initializing tracer");

    // creating a named instance of Tracer via the configured GlobalTracerProvider
    let tracer = global::tracer("say-hello-tracer");

    // creating a new span named "say-hello".
    let mut span =  tracer.start("say-hello");

    // adding an attribute to the span
    span.set_attribute(KeyValue::new("hello-to", hello_to.to_string()));

    // getting the context with the current span included
    let context_main = opentelemetry::Context::default().with_span(span);

    // calling the `format_string` function with the span context `context_main` to maintain proper parent-child relationship
    let formatted_str = FutureExt::with_context(format_string(&hello_to), context_main.clone()).await?;

    // calling the `publish_string` function with the span context `context_main` to maintain proper parent-child relationship
    FutureExt::with_context(publish_string(&formatted_str), context_main.clone()).await?;

    // ending the current span
    context_main.span().end();

    // shutting down the tracer provider to ensure all spans are flushed.
    tp.shutdown().expect("TracerProvider should shutdown successfully");
    Ok(())
}


async fn format_string(hello_to: &str) -> Result<String, reqwest::Error>{
    // retrieve or create a named tracer
    let tracer = global::tracer("say-hello-tracer");

    // start a new span named "formatString".
    let span =  tracer.start("formatString");
    
    // getting spancontext with the current span included.
    let cx = opentelemetry::Context::current_with_span(span);

    // preparing to send an http get request to the "formatter" service
    let client = Client::new();
    let mut params = HashMap::new();
    params.insert("hello_to".to_string(), hello_to.to_string());

    // creating a get request to the formatter service
    let mut req = client.get(FORMAT_URL).query(&params).build()?;

    // fetching the headers from the request just created
    let headers = req.headers_mut();

    // injecting the span context into request headers
    global::get_text_map_propagator(|propagator| {
        propagator.inject_context(&cx, &mut HeaderInjector(headers))
    });

    //sending a get request
    match client.execute(req).await{
        Err(err) =>{
            // recording the error in the span
            cx.span().record_error(&err);

            // ending the span
            cx.span().end();

            return Err(err);
        },
        Ok(resp) =>{
            let hello_str = resp.text().await?;

            // adding an event to the span indicating a successful response was received
            cx.span().add_event("format-event-response", 
            vec![
                KeyValue::new("format-response", format!("string-format: {}", hello_str)),
                ]);

            // ending the span
            cx.span().end();
            
            Ok(hello_str)
        }   
    }
}

async fn publish_string(hello_str: &str) -> Result<(), reqwest::Error>{
    // retrieve or create a named tracer
    let tracer = global::tracer("say-hello-tracer");

    // start a new span named "printHello".
    let span =  tracer.start("printHello");

    // getting spancontext with the current span included.
    let cx = opentelemetry::Context::current_with_span(span);

    // preparing to send an http get request to the "publisher" service
    let client = Client::new();
    let mut params = HashMap::new();
    params.insert("hello_str".to_string(), hello_str.to_string());

    // creating a get request to the formatter service
    let mut req = client.get(PUBLISH_URL).query(&params).build()?;

    // fetching the headers from the request just created
    let headers = req.headers_mut();

    // injecting the span context into request headers
    global::get_text_map_propagator(|propagator| {
        propagator.inject_context(&cx, &mut HeaderInjector(headers))
    });

    // sending a get request
    match client.execute(req).await{
        Err(err) =>{
            // recording the error in the span
            cx.span().record_error(&err);
            
            // ending the span
            cx.span().end();
            
            return Err(err);
        },
        Ok(_) =>{
            // ending the span
            cx.span().end();
            
            Ok(())
        }   
    }
}