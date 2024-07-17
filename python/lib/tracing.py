from opentelemetry import trace, propagate
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.resources import SERVICE_NAME, Resource
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
from opentelemetry.trace.propagation.tracecontext import TraceContextTextMapPropagator

# Tracing Backend URL
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

    # setting up a propagator to propagate traces and baggage accross microservices
    propagate.set_global_textmap(TraceContextTextMapPropagator())

    return tracer_provider

