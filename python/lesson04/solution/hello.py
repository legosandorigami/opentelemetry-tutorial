import requests
import sys
import time
from opentelemetry import trace, propagate, baggage
from opentelemetry.sdk.trace import TracerProvider, Tracer
from lib.tracing import init_tracer_provider
from opentelemetry.semconv.trace import SpanAttributes
from opentelemetry.baggage.propagation import W3CBaggagePropagator

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

# note the change in the function signature, it now also requires a python dictionary inorder to add items to the baggage
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

def print_hello(hello_str):
    # obtain a tracer instance
    tracer: Tracer = get_tracer("say_hello_tracer")
    # starting a new span for the 'print_hello' operation
    with tracer.start_as_current_span('print_hello') as span:
        http_get(8082, 'publish', 'helloStr', hello_str)
        # adding a log to the span indicating that 'print-event' event has occured
        span.add_event('print-event', {'printed': True})
        print(span.get_span_context())

# note the change in the function signature, it now also requires the context that contains the baggage in order to propagate the baggage
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

    # print(f"headers: {headers}")

    # setting baggage to propagate custom data across span boundaries.
    # Baggage items consist of a name, value, and context. The context is obtained from the current span scope
    # and is automatically updated with the new baggage item by the OpenTelemetry SDK.
    ctx = None
    if baggage_to_add is not None and isinstance(baggage_to_add, dict):
        for k, v in baggage_to_add.items():
            ctx = baggage.set_baggage(name=k, value=v, context=ctx)
        
        # uncomment to see the baggage context 
        # print(f"http_get: context: {ctx}")
    
    # injecting the baggage into the request headers form the context ctx for baggage propagation
    W3CBaggagePropagator().inject(headers, context=ctx)
    
    # print(f"headers: {headers}")

    # sending the HTTP GET request with the propagated headers
    resp = requests.get(url, params={param: value}, headers=headers)
    resp.raise_for_status()
    return resp.text
  
def get_tracer(tracer_name: str):
    # retrieve the global tracer provider
    tracer_provider: TracerProvider = trace.get_tracer_provider()
    # obtain a tracer instance from the tracer provider
    tracer: Tracer = tracer_provider.get_tracer(tracer_name)
    return tracer

def main():
    # ensuring the program is called with two argument
    assert len(sys.argv) == 3
    
    # setting and retrieving the global tracer provider
    tracer_provider: TracerProvider = init_tracer_provider('hello-world')

    hello_to = sys.argv[1]
    greeting = sys.argv[2]
    say_hello(hello_to, greeting)
    
    # Optional: yield to IOLoop to flush the spans
    # time.sleep(2)  # Can be removed if shutdown() is sufficient

    # shutting down the tracer provider to ensure all the spans are exported
    tracer_provider.shutdown()

if __name__ == '__main__':
    main()
