# Lesson 1 - Hello World

## Objectives

Learn how to:

* Instantiate a Tracer
* Create a simple trace
* Annotate the trace

## Walkthrough

### A simple Hello-World program

Let's create a simple Python program `lesson01/exercise/hello.py` that takes an argument and prints "Hello, {arg}!".

```
mkdir lesson01/exercise
touch lesson01/exercise/__init__.py
```

```python
# lesson01/exercise/hello.py
import sys

def say_hello(hello_to):
    hello_str = f'Hello, {hello_to}!'
    print(hello_str)

def main():
    assert len(sys.argv) == 2

    hello_to = sys.argv[1]
    say_hello(hello_to)

if __name__ == '__main__':
    main()
```

Run it:
```
$ python -m lesson01.exercise.hello Bryan
Hello, Bryan!
```

### Create a trace

A trace is a directed acyclic graph of spans. A span is a logical representation of some work done in your application. Each span has these minimum attributes: an operation name, a start time, and a finish time.

Let's create a trace that consists of just a single span. To do that we need an instance of the `TracerProvider`. We can use a global instance returned by `opentelemetry.trace.get_tracer_provider` function.

```python
import sys
from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider, Tracer

def say_hello(hello_to):
    # getting an instance of a default tracer provider
    tracer_provider: TracerProvider = trace.get_tracer_provider()

    # obtain a tracer instance named 'say-hello-tracer' from the tracer provider
    tracer: Tracer = tracer_provider.get_tracer("say_hello_tracer")
    
    # starting a new span named 'say-hello'
    span: Span = tracer.start_span('say_hello')
    hello_str = f'Hello, {hello_to}!'
    print(hello_str)

    # ending the span so that it gets exported properly
    span.end()

def main():
    assert len(sys.argv) == 2

    hello_to = sys.argv[1]
    say_hello(hello_to)

if __name__ == '__main__':
    main()
```

We are using the following basic features of the OpenTelemetry API:
  * a `TracerProvider` A tracer provider creates tracers which can then be used to create traces and spans.
  * a `Tracer` instance is used to start new spans via `start_as_current_span` function
  * each `span` is given an _operation name_, `"say-hello"` in this case
  * each `span` must be ended by calling its `end` function.
  * the start and end timestamps of the span will be captured automatically by the tracer implementation.

However, calling `end()` manually is a bit tedious, we can use span as a context manager instead:

```python
def say_hello(hello_to):
    with tracer.start_as_current_span('say-hello') as span:
        hello_str = f'Hello, {hello_to}!'
        print(hello_str)
```

If we run this program, we will see no difference, and no traces in the tracing UI. That's because the global tracer provider is a no-op tracer provider by default.

### Initialize a real tracer

Let's create an instance of a real tracer provider. To do this we need to create a custom tracer provider and set it as the global tracer provider

```python
from opentelemetry import trace, propagate
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.resources import SERVICE_NAME, Resource
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter

# Backend URLs for different tracing systems
TRACING_BACKEND = "http://localhost:4318/v1/traces"

# initialize the tracer provider
def init_tracer_provider(service: str, backend: str = TRACING_BACKEND) -> TracerProvider:    
    # creating an OTLP HTTP exporter
    otlp_exporter = OTLPSpanExporter(endpoint=backend)

    # creating a tracer provider with the OTLP exporter
    tracer_provider = TracerProvider(resource=Resource.create({SERVICE_NAME: service}))

    # adding a span processor to the tracer provider to handle span export
    tracer_provider.add_span_processor(BatchSpanProcessor(otlp_exporter))

    # setting the global tracer provider to the one we just created 
    trace.set_tracer_provider(tracer_provider)

    return tracer_provider
```

To use this instance, let's update the main function:

```python
def main():
    # ensuring the program is called with one argument
    assert len(sys.argv) == 2

    # setting and retrieving the global tracer provider
    tracer_provider: TracerProvider = init_tracer_provider

    hello_to = sys.argv[1]
    
    say_hello(hello_to)
```

Note that we are passing a string `'hello-world'` to the init method. It is used to mark all spans emitted by the tracer as originating from a `hello-world` service.

There's one more thing we need to do. Some exporters, especially batch exporters, may not send the spans immediately. The sleep gives them time to process and send the spans to the backend. Since our program exits immediately, it may not have time to flush the spans to the backends. Let's add the following code to the `main` function in the `hello.py`:

```python
import time

# previous code

# yield to IOLoop to flush the spans
time.sleep(2)

# shutting down the tracer provider to ensure all the spans are exported
tracer_provider.shutdown()
```

If we run the program now, we should see a span logged:

```
$ python -m lesson01.exercise.hello Brian
Hello, Brian!
SpanContext(trace_id=0x3adc967f6daf923b3816df562b93da5d, span_id=0x34a859ed99f6dd5d, trace_flags=0x01, trace_state=[], is_remote=False)
```

If you have a tracing backend(Signoz or Tempo with Grafana or Jaeger) running, you should be able to see the trace in the UI.

### Annotate the Trace with Attributes and Events

Right now the trace we created is very basic. If we call our program with argument `Susan` instead of `Bryan`, the resulting traces will be nearly identical. It would be nice if we could capture the program arguments in the traces to distinguish them.

One naive way is to use the string `"Hello, Bryan!"` as the _operation name_ of the span, instead of `"say-hello"`. However, such practice is highly discouraged in distributed tracing, because the operation name is meant to represent a _class of spans_, rather than a unique instance. For example, in Jaeger UI you can select the operation name from a dropdown when searching for traces. It would be very bad user experience if we ran the program to say hello to a 1000 people and the dropdown then contained 1000 entries. Another reason for choosing more general operation names is to allow the tracing systems to do aggregations. For example, Jaeger tracer has an option of emitting metrics for all the traffic going through the application. Having a unique operation name for each span would make the metrics useless.

The recommended solution is to add attributes(tags) and events(logs) to spans. An _attribute_ is a key-value pair that provides certain metadata about the span. An _event_ is similar to a regular log statement, it contains a timestamp and some data, but it is associated with span from which it was logged.

When should we use attributes vs. events?  The attributes are meant to describe attributes of the span that apply to the whole duration of the span. For example, if a span represents an HTTP request, then the URL of the request should be recorded as an attribute because it does not make sense to think of the URL as something that's only relevant at different points in time on the span. On the other hand, if the server responded with a redirect URL, logging it would make more sense since there is a clear timestamp associated with such an event.

#### Adding Attributes

In the case of `hello Bryan`, the string "Bryan" is a good candidate for a span tag, since it applies to the whole span and not to a particular moment in time. We can record it like this:

```python
with tracer.start_as_current_span('say-hello') as span:
    # adding an attribute
    span.set_attribute("hello-to", hello_to)

    hello_str = f'Hello, {hello_to}!'
    print(hello_str)
```

#### Adding Events

Our hello program is so simple that it's difficult to find a relevant example of a log, but let's try. Right now we're formatting the `hello_str` and then printing it. Both of these operations take certain time, so we can log their completion:

```python
with tracer.start_as_current_span('say-hello') as span:
    # adding an attribute
    span.set_attribute("hello-to", hello_to)

    hello_str = f'Hello, {hello_to}!'
    
    # adding an event to the span indicating that the string has been formatted
    span.add_event('string-format-event', {'string-formatted': hello_str})

    
    # printing the greeting
    print(hello_str)
        
    # adding an event to the span indicating that the greeting has been printed
    span.add_event('print-event', {'println': True})
```

The log statements might look a bit strange if you have not previosuly worked with a structured logging API. Rather than formatting a log message into a single string that is easy for humans to read, structured logging APIs encourage you to separate bits and pieces of that message into key-value pairs that can be automatically processed by log aggregation systems. The idea comes from the realization that today most logs are processed by machines rather than humans.

The OpenTelemetry API for Python exposes structured logging through the use of events and attributes:
* The `add_event` method allows you to add an event to a span, which can include a name and a set of attributes (could be a python dictionary). Events are time-stamped and provide a way to log significant occurrences within the span’s lifetime.
* Attributes can be added to events using the`set_attribute` method, which takes a key-value pair in the form of string and an`AttributeValue`.

If you run the program with these changes, then find the trace in the UI and expand its span (by clicking on it), you will be able to see the attributes and events.

## Conclusion

The complete program can be found in the [solution](./solution) package. We moved the `init_tracer_provider` helper function into its own package so that we can reuse it in the other lessons as `from lib.tracing import init_tracer_provider`.

Next lesson: [Context and Tracing Functions](../lesson02).
