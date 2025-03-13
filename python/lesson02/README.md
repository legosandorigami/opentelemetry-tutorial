# Lesson 2 - Context and Tracing Functions

## Objectives

Learn how to:

* Trace individual functions
* Combine multiple spans into a single trace
* Propagate the in-process context

## Walkthrough

First, copy your work or the official solution from [Lesson 1](../lesson01) to `lesson02/exercise/hello.py`,
and make it a module by creating the `__init__.py` file:

```bash
mkdir lesson02/exercise
touch lesson02/exercise/__init__.py
cp lesson01/solution/*.py lesson02/exercise/
```

### Tracing individual functions

In [Lesson 1](../lesson01) we wrote a program that creates a trace that consists of a single span. That single span combined two operations performed by the program, formatting the output string and printing it. Let's move those operations into standalone functions first:

```python
from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider, Tracer, Span

def say_hello(hello_to: str):
    # obtain a tracer instance
    tracer: Tracer = get_tracer("say_hello_tracer")
    # starting a new span for the 'say_hello' operation
    span: Span = tracer.start_span('say_hello')
    # setting an attribute on the span
    span.set_attribute("hello-to", hello_to)
    # calling the format_string function
    hello_str = format_string(span, hello_to)
    # calling the print_hello function
    print_hello(span, hello_str)
    print(span.get_span_context())
    # ending the span so that it gets exported properly
    span.end()

def format_string(root_span: Span, hello_to: str):
    hello_str = f'Hello, {hello_to}!'
    # adding a log to the span indicating that 'string-format-event' event has occured
    root_span.add_event('string-format-event', {'string-formatted': hello_str})
    return hello_str

def print_hello(root_span: Span, hello_str: str):
    print(hello_str)
    # adding a log to the span indicating that 'print-event' event has occured
    root_span.add_event('print-event', {'printed': True})

def get_tracer(tracer_name: str):
    # retrieve the global tracer provider
    tracer_provider: TracerProvider = trace.get_tracer_provider()
    # obtain a tracer instance from the tracer provider
    tracer: Tracer = tracer_provider.get_tracer(tracer_name)
    return tracer
```

Of course, this does not change the outcome. What we really want to do is to wrap each function into its own span.

### Wrapping Each Function in Its Own Span

To track the execution of individual functions as separate spans, we need start a new span in each of the functions:

```python
from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider, Tracer, Span

def say_hello(hello_to: str):
    # obtain a tracer instance
    tracer: Tracer = get_tracer("say_hello_tracer")
    # starting a new span for the 'say_hello' operation
    span: Span = tracer.start_span('say_hello')
    # setting an attribute on the span
    span.set_attribute("hello-to", hello_to)
    # calling the format_string function
    hello_str = format_string(span, hello_to)
    # calling the print_hello function
    print_hello(span, hello_str)
    print(span.get_span_context())
    # ending the span so that it gets exported properly
    span.end()

def format_string(root_span: Span, hello_to: str):
    # obtain a tracer instance from the tracer provider
    tracer: Tracer = get_tracer("say_hello_tracer")
    # starting a new span
    span: Span = tracer.start_span('format_string', context=None)
    hello_str = f'Hello, {hello_to}!'
    # adding a log to the span indicating that 'string-format-event' event has occured
    span.add_event('string-format-event', {'string-formatted': hello_str})
    print(span.get_span_context())
    # ending the span so that it gets exported properly
    span.end()
    return hello_str

def print_hello(root_span: Span, hello_str: str):
    # obtain a tracer instance from the tracer provider
    tracer: Tracer = get_tracer("say_hello_tracer")
    # starting a new span for the 'print_hello' operation
    span: Span = tracer.start_span('print_hello', context=None)
    print(hello_str)
    # adding a log to the span indicating that 'print-event' event has occured
    span.add_event('print-event', {'printed': True})
    print(span.get_span_context())
    # ending the span so that it gets exported properly
    span.end()
```

Let's run it:

```bash
$ python -m lesson02.exercise.hello Brian
SpanContext(trace_id=0x00dc1afbdfaea8186452a698f9dd89b7, span_id=0xb320525457294a17, trace_flags=0x01, trace_state=[], is_remote=False)
Hello, brian!
SpanContext(trace_id=0xcda81f03e785c65e86aabc2c6daf6cba, span_id=0x5a8fdbba85eaac25, trace_flags=0x01, trace_state=[], is_remote=False)
SpanContext(trace_id=0xde14b9ad90eb7d2dc56c21a4c994a9bd, span_id=0x5ebff4f00cc98a23, trace_flags=0x01, trace_state=[], is_remote=False)
```

We got three spans, but there is a problem here. The `trace_id`s are all different. If we search for those IDs in the UI each one will represent a standalone trace with a single span. That's not what we wanted!

What we really wanted was to establish causal relationship between the two new spans to the root span started in `say_hello` function. We can do that by passing an additional option `context` which can be obtained from the `root_span`, to the `start_span` function:

```python
from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider, Tracer, Span

def say_hello(hello_to):
    # obtain a tracer instance
    tracer: Tracer = get_tracer("say_hello_tracer")
    # starting a new span for the 'say_hello' operation
    span: Span = tracer.start_span('say_hello')
    # setting an attribute on the span
    span.set_attribute("hello-to", hello_to)
    # calling the format_string function with the root span
    hello_str = format_string(span, hello_to)
    # calling the print_hello function with the root span
    print_hello(span, hello_str)
    print(span.get_span_context())
    # ending the span to ensure it gets exported properly
    span.end()

def format_string(root_span: Span, hello_to: str):
    # obtain a tracer instance from the tracer provider
    tracer: Tracer = get_tracer("say_hello_tracer")
    # creating a new context with the root span
    ctx = trace.set_span_in_context(span=root_span)
    # starting a new span within context of the root span
    span: Span = tracer.start_span('format_string', context=ctx)
    hello_str = f'Hello, {hello_to}!'
    # adding a log to the span indicating that 'string-format-event' event has occured
    span.add_event('string-format-event', {'string-formatted': hello_str})
    print(span.get_span_context())
    # ending the span to ensure it gets exported properly
    span.end()
    return hello_str

def print_hello(root_span: Span, hello_str: str):
    # obtain a tracer instance from the tracer provider
    tracer: Tracer = get_tracer("say_hello_tracer")
    # creating a new context with the root span
    ctx = trace.set_span_in_context(span=root_span)
    # starting a new span within context of the root span
    span: Span = tracer.start_span('print_hello', context=ctx)
    print(hello_str)
    # adding a log to the span indicating that 'print-event' event has occured
    span.add_event('print-event', {'printed': True})
    print(span.get_span_context())
    # ending the span to ensure it gets exported properly
    span.end()
```
If we modify the `format_string` and `print_hello` functions accordingly and run the app, we'll see that all reported spans now belong to the same trace(all the spans now have the same `trace_id`):

```bash
$ python -m lesson02.exercise.hello Brian
SpanContext(trace_id=0x50870936626e47f99ea53d1ba3d7f86c, span_id=0x5f1a92ba9b3ca2bc, trace_flags=0x01, trace_state=[], is_remote=False)
Hello, Brian!
SpanContext(trace_id=0x50870936626e47f99ea53d1ba3d7f86c, span_id=0x3665366c75ca7bed, trace_flags=0x01, trace_state=[], is_remote=False)
SpanContext(trace_id=0x50870936626e47f99ea53d1ba3d7f86c, span_id=0x715a6ab493209ca1, trace_flags=0x01, trace_state=[], is_remote=False)
```

If we find this trace in the UI, it will show a proper parent-child relationship between the spans.

### Propagate the in-process context

If we think of a trace as a directed acyclic graph where nodes are the spans and edges are the causal relationships between them. In OpenTelemetry, relationships between spans are implicitly handled through the context, with the parent-child relationship being the default when creating a new span within the context of an existing span. This means that the `root_span` has a logical dependency on the child span before the `root_span` can complete its operation.

You may have noticed that we had to pass the `Span` object as the first argument to each function. This can be cumbersome and error-prone. Unlike Go, languages like Java and Python support thread-local storage, which is convenient for storing request-scoped data like the current span.

In Python, the `contextvars` module allows for managing context in asynchronous environments. OpenTelemetry leverages this to maintain context across asynchronous calls. This means that you don't need to manually pass the span object through every function. Instead, the current span is stored in a context variable, and OpenTelemetry ensures that this context is correctly propagated across function calls, threads, and asynchronous tasks. This makes your code cleaner and reduces the likelihood of errors related to context propagation. Also using `contextvars` module helps maintain backwards compatibility, as you do not need to change function signatures just to instrument your application. This means you can add tracing to your existing codebase without significant modifications, preserving the original structure and behavior of your functions.

Instead of calling the `tracer.start_span` function directly, we can utilize the `start_as_current_span` method of the `tracer`. This method makes activates the span and makes it accessible via the `trace.get_current_span` function. Typically used within a `with` statement, `start_as_current_span` function ensures that the span is automatically ended and the previous span is restored when the block exits. Any function called within this `with` block that creates a new span will have the active root context available, allowing the new span to automatically become a child of the active span.

```python
def say_hello(hello_to):
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
```

Notice that we're no longer passing the span as an argument to the functions because the functions are being called within a context manager. Creating a child span of a currently active span is a common pattern in OpenTelemetry and happens automatically. Therefore, we do not need to explicitly pass the optional `context` parameter of the currently active root span to the `tracer.start_as_current_span` function, the OpenTelemetry SDK automatically fetches the current active context if one exists.

```python

 # creating a new context with the root span
    ctx = trace.set_span_in_context(span=root_span)
    # starting a new span within context of the root span
    span: Span = tracer.start_span('print_hello', context=ctx)
def format_string(hello_to):
    # obtain a tracer instance from the tracer provider
    tracer: Tracer = get_tracer("say_hello_tracer")
    # starting a new span for the 'format_string' operation
    with tracer.start_as_current_span('format_string') as span:
        hello_str = f'Hello, {hello_to}!'
        # adding a log to the span indicating that 'string-format-event' event has occured
        span.add_event('string-format-event', {'string-formatted': hello_str})
        print(span.get_span_context())
        return hello_str

def print_hello(hello_str):
    # obtain a tracer instance from the tracer provider
    tracer: Tracer = get_tracer("say_hello_tracer")
    # starting a new span for the 'print_hello' operation
    with tracer.start_as_current_span('print_hello') as span:
        print(hello_str)
        # adding a log to the span indicating that 'print-event' event has occured
        span.add_event('print-event', {'printed': True})
        print(span.get_span_context())
```

Note that because the `start_as_current_span` method is used within a `with` statement, we access the span directly within the block. To annotate it with attributes or events, we use the span instance available in the with block.

If we run this modified program, we will see that all three reported spans still have the same trace ID.

### What's the Big Deal?

The last change we made may not seem particularly useful. But imagine that your program is
much larger with many functions calling each other. By using the Scope Manager mechanism we can access
the current span from any place in the program without having to pass the span object as the argument to
all the function calls. This is especially useful if we are using instrumented RPC frameworks that perform
tracing functions automatically - they have a stable way of finding the current span.

## Conclusion

The complete program can be found in the [solution](./solution) package.

Next lesson: [Tracing RPC Requests](../lesson03).
