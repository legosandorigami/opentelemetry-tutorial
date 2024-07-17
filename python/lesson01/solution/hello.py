import sys
import time
from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider, Tracer
from lib.tracing import init_tracer_provider

def say_hello(hello_to):
    # retrieve the global tracer provider
    tracer_provider: TracerProvider = trace.get_tracer_provider()

    # obtain a tracer instance from the tracer provider
    tracer: Tracer = tracer_provider.get_tracer("say_hello_tracer")
    
    # starting a new span for the 'say-hello' operation
    with tracer.start_as_current_span('say-hello') as span:
        # seting an attribute on the span
        span.set_attribute("hello-to", hello_to)

        hello_str = f'Hello, {hello_to}!'
        
        # adding an event to the span indicating that the string has been formatted
        span.add_event('string-format-event', {'string-formatted': hello_str})

        # printing the greeting
        print(hello_str)
        
        # adding an event to the span indicating that the greeting has been printed
        span.add_event('print-event', {'println': True})

        print(span.get_span_context())

def main():
    # ensuring the program is called with one argument
    assert len(sys.argv) == 2

    # setting and retrieving the global tracer provider
    tracer_provider: TracerProvider = init_tracer_provider('hello-world')

    hello_to = sys.argv[1]
    
    say_hello(hello_to)
    
    # Optional: yield to IOLoop to flush the spans
    # time.sleep(2)  # Can be removed if shutdown() is sufficient
    
    # shutting down the tracer provider to ensure all the spans are exported
    tracer_provider.shutdown()

if __name__ == '__main__':
    main()
