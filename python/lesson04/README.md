# Lesson 4 - Baggage

## Objectives

* Understand distributed context propagation
* Use baggage to pass data through the call graph

### Walkthrough

In Lesson 3 we have seen how span context is propagated over the wire between different applications. It is not hard to see that this process can be generalized to passing more than just the tracing context. With OpenTelemetry instrumentation in place, we can support general purpose _distributed context propagation_ where we associate some metadata with the transaction and make that metadata available anywhere in the distributed call graph. In OpenTelemetry this metadata is called _baggage_, to highlight the fact that it is carried over in-band with all RPC requests, just like baggage.

To see how it works in OpenTelemetry, let's take the application we built in Lesson 3. You can copy the source code from [../lesson03/solution](../lesson03/solution) package:

```
mkdir lesson04/exercise
cp -r lesson03/solution/*py lesson04/exercise/
```

The `formatter` service takes the `helloTo` parameter and returns a string `Hello, {helloTo}!`. Let's modify it so that we can customize the greeting too, but without modifying the public API of that service.

### Set Baggage in the Client

Let's modify the function `main` in `hello.py` as follows:

```python
assert len(sys.argv) == 3

hello_to = sys.argv[1]
greeting = sys.argv[2]
say_hello(hello_to, greeting)
```
And update `say_hello` function:

```python
def say_hello(hello_to, greeting):
    # obtain a tracer instance
    tracer: Tracer = get_tracer("say_hello_tracer")
    # starting a new span for the 'say_hello' operation
    with tracer.start_as_current_span('say_hello') as span:
        # setting an attribute on the span
        span.set_attribute("hello-to", hello_to)
        # calling the format_string function
        hello_str = format_string(hello_to, {"greeting": greeting})
        # calling the print_hello function
        print_hello(hello_str)        
        print(span.get_span_context())
```
Here we read a second command line argument as a "greeting". We also need to modify the signatures of the function `format_string` and `http_get` functions in `hello.py` so that they now also accept an additional argument of type python dictionary. The `http_get` function will be responsible for propagating the baggage items to the `Formatter` service. Let's modify the `format_string` and `http_get` functions in `hello.py` as follows:

```python
def format_string(hello_to, baggage_to_add: dict=None):
    # obtain a tracer instance
    tracer: Tracer = get_tracer("say_hello_tracer")
    # starting a new span for the 'format_string' operation
    with tracer.start_as_current_span('format_string') as span:
        hello_str = http_get(8081, 'format', 'helloTo', hello_to, baggage_to_add)
        # adding a log to the span indicating that 'string-format-event' event has occured
        span.add_event('string-format-event', {'string-formatted': hello_str})
        print(span.get_span_context())
        return hello_str  
```

```python
def http_get(port, path, param, value, baggage_to_add: dict = None):
    url = 'http://localhost:%s/%s' % (port, path)
    # retrieving the current active span
    span = trace.get_current_span()
    # setting HTTP method and URL attributes on the span
    span.set_attribute(SpanAttributes.HTTP_METHOD, 'GET')
    span.set_attribute(SpanAttributes.HTTP_URL, url)

    # iniecting the span context into the HTTP headers for trace propagation
    headers = {}
    propagate.inject(headers)

    # setting baggage to propagate custom data across span boundaries.
    # Baggage items consist of a name, value, and context. The context is obtained from the current span scope
    # and is automatically updated with the new baggage item by the OpenTelemetry SDK.
    ctx = None
    if baggage_to_add is not None and isinstance(baggage_to_add, dict):
        for k, v in baggage_to_add.items():
            ctx = baggage.set_baggage(name=k, value=v, context=ctx)
    
    # injecting the baggage into the request headers form the context ctx for baggage propagation
    W3CBaggagePropagator().inject(headers, context=ctx)

    # sending the HTTP GET request with the propagated headers
    resp = requests.get(url, params={param: value}, headers=headers)
    resp.raise_for_status()
    return resp.text
```

`propagate.inject` only injects the trace context. To properly propagate baggage, we need to use the `W3CBaggagePropagator` in addition to the trace context propagator. 
In the modified code for `http_get`, we first inject the trace context. Then, we check if there are any items to add to the baggage. If there are, we create a new baggage context containing these items and inject this new context into the request headers using `W3CBaggagePropagator`. 


### Read Baggage in the Formatter Service

Add the following code to the `formatter`'s HTTP handler in `formatter.py` file:

```python
# previous code

# extracting the trace context from the request headers for distributed tracing
trace_ctx = propagate.extract(carrier=request.headers)
# extracting the baggage context from the request headers for distributed baggage propagation
baggage_ctx = W3CBaggagePropagator().extract(carrier=request.headers)

# starting a new span with the extracted trace context
with tracer.start_as_current_span('format', context=trace_ctx) as span:

    # previous code

    # # retrieving the 'greeting' value from the baggage context
    greeting = baggage.get_baggage("greeting",context=baggage_ctx)
    if greeting is None:
        greeting = "Hello"
    hello_str = f"{greeting}, {hello_to}!"

    # previous code

```

### Run it

As in Lesson 3, first start the `formatter` and `publisher` in separate terminals, then run the client with two arguments, e.g. `Bryan Bonjour`. The `publisher` should print `Bonjour, Bryan!`.

```
# client
$ python -m lesson04.exercise.hello Brian Bonjour
SpanContext(trace_id=0xee3ac40657648e0731bc66d31df8136f, span_id=0x04c15f877782773f, trace_flags=0x01, trace_state=[], is_remote=False)
SpanContext(trace_id=0xee3ac40657648e0731bc66d31df8136f, span_id=0x2165ba180ff281f8, trace_flags=0x01, trace_state=[], is_remote=False)
SpanContext(trace_id=0xee3ac40657648e0731bc66d31df8136f, span_id=0xa2019245569e2b88, trace_flags=0x01, trace_state=[], is_remote=False)

# formatter
$ python -m lesson04.exercise.formatter
SpanContext(trace_id=0xee3ac40657648e0731bc66d31df8136f, span_id=0xee49d176d752a844, trace_flags=0x01, trace_state=[], is_remote=False)
127.0.0.1 - - [13/Jul/2024 14:02:17] "GET /format?helloTo=Brian HTTP/1.1" 200 -

# publisher
$ python -m lesson04.exercise.publisher
Bonjour, Brian!
SpanContext(trace_id=0xee3ac40657648e0731bc66d31df8136f, span_id=0xd8f41fc0eb6c5835, trace_flags=0x01, trace_state=[], is_remote=False)
127.0.0.1 - - [13/Jul/2024 14:02:17] "GET /publish?helloStr=Bonjour,+Brian! HTTP/1.1" 200 -
```

### What's the Big Deal?

We may ask - so what, we could've done the same thing by passing the `greeting` as an HTTP request parameter. However, that is exactly the point of this exercise - we did not have to change any APIs on the path from the root span in `hello.py` all the way to the server-side span in `formatter`, three levels down. If we had a much larger application with much deeper call tree, say the `formatter` was 10 levels down, the exact code changes we made here would have worked, despite 8 more services being in the path. If changing the API was the only way to pass the data, we would have needed to modify 8 more services to get the same effect.

Some of the possible applications of baggage include:

  * passing the tenancy in multi-tenant systems
  * passing identity of the top caller
  * passing fault injection instructions for chaos engineering
  * passing request-scoped dimensions for other monitoring data, like separating metrics for prod vs. test traffic


### Now, a Warning... NOW a Warning?

Of course, while baggage is an extremely powerful mechanism, it is also dangerous. If we store a 1Mb value/string in baggage, every request in the call graph below that point will have to carry that 1Mb of data. So baggage must be used with caution. In fact, Jaeger client libraries implement centrally controlled baggage restrictions, so that only blessed services can put blessed keys in the baggage, with possible restrictions on the value length.

## Conclusion

The complete program can be found in the [solution](./solution) package.
