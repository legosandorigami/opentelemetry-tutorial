# Lesson 4 - Baggage

## Objectives

* Understand distributed context propagation
* Use baggage to pass data through the call graph

### Walkthrough

In Lesson 3 we have seen how span context is propagated over the wire between different applications. It is not hard to see that this process can be generalized to propagating more than just the tracing context. With OpenTelemetry instrumentation in place, we can support general purpose _distributed context propagation_ where we associate some metadata with the transaction and make that metadata available anywhere in the distributed call graph. In OpenTelemetry this metadata is called _baggage_, to highlight the fact that it is carried over in-band with all RPC requests, just like baggage.

To see how it works in OpenTracing, let's take the application we built in Lesson 3. You can copy the source code from [../lesson03/solution](../lesson03/solution) package:

```
cp -r ./lesson03/solution ./lesson04/exercise
```

The `formatter` service takes the `hello_to` parameter and returns a string `Hello, {hello_to}!`. Let's modify it so that we can customize the greeting too, but without modifying the public API of that service.

### Set Baggage in the Client

Let's modify the function `main` in `client.rs` as follows:

```rust
// checking if the number of command-line arguments is exactly 2 (program name and one argument).
let args: Vec<_> = env::args().collect();
if args.len() != 3 {
    panic!("ERROR: Expecting two arguments");
}
let hello_to = args[1].clone();
let greeting = args[2].clone();

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
    
// creating baggage items vector and add "greeting"
let baggage_items = vec![KeyValue::new("greeting", greeting),];

// calling the `format_string` function with the span context `context_main` to maintain proper parent-child relationship
let formatted_str = FutureExt::with_context(format_string(&hello_to, baggage_items),context_main.clone()).await?;
```
Here we read a second command line argument as a "greeting". We also need to modify the signature of the function `format_string` in `client.rs` so that it now accepts an additional argument of type `Vec<KeyValue>`. The `format_string` function in `client.rs` will be responsible for propagating the baggage items to the `formatter` service. Let's modify the function `format_string` in `client.rs` as follows:

```rust
async fn format_string(hello_to: &str, baggage_items: Vec<KeyValue>) -> Result<String, reqwest::Error>{
    // retrieve or create a named tracer
    let tracer = global::tracer("say-hello-tracer");

    // start a new span named "formatString".
    let span =  tracer.start("formatString");
    
    // getting spancontext with the current span included.
    let mut cx = opentelemetry::Context::current_with_span(span);

    // adding baggage to the context
    cx = cx.with_baggage(baggage_items);

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

	// previous code

}
```

### Read Baggage in Formatter

Add the following code to the `formatter`'s HTTP handler:

```rust
// extracting baggage from the context
let baggage = cx.baggage();

let mut greeting = "Hello".to_string();
// retrieving the member from the baggage with the key "greeting"
if let Some(greeting_) = baggage.get("greeting"){
    greeting = greeting_.to_string();
};

let mut resp= format!("{} there!", greeting);

if let Some(hello_to) = params.get("hello_to") {
    resp = format!("{}, {}!", greeting, hello_to);
}

// previous code
```

### Run it

As in Lesson 3, first start the `formatter` and `publisher` in separate terminals, then run the client with two arguments, e.g. `hello.go Brian Bonjour`. The `publisher` should print `Bonjour, Bryan!`.

```bash
# client
$ cargo run --bin Bryan Bonjour
```
```bash
# formatter
$ cargo run --bin formatter
```
```bash
# publisher
$ cargo run --bin publisher
```

We see `Hello, Bryan!` instead of `Bonjour, Bryan!`. That is because the baggage has not been propagated. We set the global propagator to an instance of `TraceContextPropagator` in the function `init_tracer` of our helper library. We need to use a composite propagator that can also propagate the baggage in addition to the trace context. Let's update the `init_tracer` function in our helper library in the file `lib.rs`:

```rust

...

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

...

```
Now let's re-run the application:

```bash
# client
$ cargo run --bin Bryan Bonjour
```
```bash
# formatter
$ cargo run --bin formatter
```
```bash
# publisher
$ cargo run --bin publisher
```

### What's the Big Deal?

We may ask - so what, we could've done the same thing by passing the `greeting` as an HTTP request parameter.
However, that is exactly the point of this exercise - we did not have to change any APIs on the path from
the root span in `client.rs` all the way to the server-side span in `formatter`, three levels down.
If we had a much larger application with much deeper call tree, say the `formatter` was 10 levels down,
the exact code changes we made here would have worked, despite 8 more services being in the path.
If changing the API was the only way to pass the data, we would have needed to modify 8 more services
to get the same effect.

Some of the possible applications of baggage include:

  * passing the tenancy in multi-tenant systems
  * passing identity of the top caller
  * passing fault injection instructions for chaos engineering
  * passing request-scoped dimensions for other monitoring data, like separating metrics for prod vs. test traffic


### Now, a Warning... NOW a Warning?

Of course, while baggage is an extermely powerful mechanism, it is also dangerous. If we store a 1Mb value/string
in baggage, every request in the call graph below that point will have to carry that 1Mb of data. So baggage
must be used with caution. In fact, Jaeger client libraries implement centrally controlled baggage restrictions,
so that only blessed services can put blessed keys in the baggage, with possible restrictions on the value length.

## Conclusion

The complete program can be found in the [solution](./solution) package.
