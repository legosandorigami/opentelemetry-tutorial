import sys
import time
from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider, Tracer, Span
from lib.tracing import init_tracer_provider

# ********************** FOURTH ATTEMPT **************************
def say_hello(hello_to: str):
    # obtain a tracer instance
    tracer: Tracer = get_tracer("say_hello_tracer")
    # starting a new span for the 'say_hello' operation
    with tracer.start_as_current_span('say_hello') as span:
        # setting an attribute on the span
        span.set_attribute("hello-to", hello_to)
        # calling the format_string function
        hello_str = format_string(hello_to)
        # calling the print_hello function
        print_hello(hello_str)
        print(span.get_span_context())

def format_string(hello_to: str):
    # obtain a tracer instance from the tracer provider
    tracer: Tracer = get_tracer("say_hello_tracer")
    # starting a new span for the 'format_string' operation
    with tracer.start_as_current_span('format_string') as span:
        hello_str = f'Hello, {hello_to}!'
        # adding a log to the span indicating that 'string-format-event' event has occured
        span.add_event('string-format-event', {'string-formatted': hello_str})
        print(span.get_span_context())
        return hello_str

def print_hello(hello_str: str):
    # obtain a tracer instance from the tracer provider
    tracer: Tracer = get_tracer("say_hello_tracer")
    # starting a new span for the 'print_hello' operation
    with tracer.start_as_current_span('print_hello') as span:
        print(hello_str)
        # adding a log to the span indicating that 'print-event' event has occured
        span.add_event('print-event', {'printed': True})
        print(span.get_span_context())


# ********************** THIRD ATTEMPT **************************
# def say_hello(hello_to: str):
#     # obtain a tracer instance
#     tracer: Tracer = get_tracer("say_hello_tracer")
#     # starting a new span for the 'say_hello' operation
#     span: Span = tracer.start_span('say_hello')
#     # setting an attribute on the span
#     span.set_attribute("hello-to", hello_to)
#     # calling the format_string function with the root span
#     hello_str = format_string(span, hello_to)
#     # calling the print_hello function with the root span
#     print_hello(span, hello_str)
#     print(span.get_span_context())
#     # ending the span to ensure it gets exported properly
#     span.end()

# def format_string(root_span: Span, hello_to: str):
#     # obtain a tracer instance from the tracer provider
#     tracer: Tracer = get_tracer("say_hello_tracer")
#     # creating a new context with the root span
#     ctx = trace.set_span_in_context(span=root_span)
#     # starting a new span within context of the root span
#     span: Span = tracer.start_span('format_string', context=ctx)
#     hello_str = f'Hello, {hello_to}!'
#     # adding a log to the span indicating that 'string-format-event' event has occured
#     span.add_event('string-format-event', {'string-formatted': hello_str})
#     print(span.get_span_context())
#     # ending the span to ensure it gets exported properly
#     span.end()
#     return hello_str

# def print_hello(root_span: Span, hello_str: str):
#     # obtain a tracer instance from the tracer provider
#     tracer: Tracer = get_tracer("say_hello_tracer")
#     # creating a new context with the root span
#     ctx = trace.set_span_in_context(span=root_span)
#     # starting a new span within context of the root span
#     span: Span = tracer.start_span('print_hello', context=ctx)
#     print(hello_str)
#     # adding a log to the span indicating that 'print-event' event has occured
#     span.add_event('print-event', {'printed': True})
#     print(span.get_span_context())
#     # ending the span to ensure it gets exported properly
#     span.end()


# ********************** SECOND ATTEMPT **************************
# def say_hello(hello_to: str):
#     # obtain a tracer instance
#     tracer: Tracer = get_tracer("say_hello_tracer")
#     # starting a new span for the 'say_hello' operation
#     span: Span = tracer.start_span('say_hello')
#     # setting an attribute on the span
#     span.set_attribute("hello-to", hello_to)
#     # calling the format_string function
#     hello_str = format_string(span, hello_to)
#     # calling the print_hello function
#     print_hello(span, hello_str)
#     print(span.get_span_context())
#     # ending the span so that it gets exported properly
#     span.end()

# def format_string(root_span: Span, hello_to: str):
#     # obtain a tracer instance from the tracer provider
#     tracer: Tracer = get_tracer("say_hello_tracer")
#     # starting a new span
#     span: Span = tracer.start_span('format_string', context=None)
#     hello_str = f'Hello, {hello_to}!'
#     # adding a log to the span indicating that 'string-format-event' event has occured
#     span.add_event('string-format-event', {'string-formatted': hello_str})
#     print(span.get_span_context())
#     # ending the span so that it gets exported properly
#     span.end()
#     return hello_str

# def print_hello(root_span: Span, hello_str: str):
#     # obtain a tracer instance from the tracer provider
#     tracer: Tracer = get_tracer("say_hello_tracer")
#     # starting a new span for the 'print_hello' operation
#     span: Span = tracer.start_span('print_hello', context=None)
#     print(hello_str)
#     # adding a log to the span indicating that 'print-event' event has occured
#     span.add_event('print-event', {'printed': True})
#     print(span.get_span_context())
#     # ending the span so that it gets exported properly
#     span.end()


# ********************** FIRST ATTEMPT **************************
# def say_hello(hello_to: str):
#     # obtain a tracer instance
#     tracer: Tracer = get_tracer("say_hello_tracer")
#     # starting a new span for the 'say_hello' operation
#     span: Span = tracer.start_span('say_hello')
#     # setting an attribute on the span
#     span.set_attribute("hello-to", hello_to)
#     # calling the format_string function
#     hello_str = format_string(span, hello_to)
#     # calling the print_hello function
#     print_hello(span, hello_str)
#     print(span.get_span_context())
#     # ending the span so that it gets exported properly
#     span.end()

# def format_string(root_span: Span, hello_to: str):
#     hello_str = f'Hello, {hello_to}!'
#     # adding a log to the span indicating that 'string-format-event' event has occured
#     root_span.add_event('string-format-event', {'string-formatted': hello_str})
#     return hello_str

# def print_hello(root_span: Span, hello_str: str):
#     print(hello_str)
#     # adding a log to the span indicating that 'print-event' event has occured
#     root_span.add_event('print-event', {'printed': True})

def get_tracer(tracer_name: str):
    # retrieve the global tracer provider
    tracer_provider: TracerProvider = trace.get_tracer_provider()
    # obtain a tracer instance from the tracer provider
    tracer: Tracer = tracer_provider.get_tracer(tracer_name)
    return tracer

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
