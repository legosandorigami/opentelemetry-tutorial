from flask import Flask, request
from opentelemetry import propagate, trace
from opentelemetry.sdk.trace import TracerProvider, Tracer
from lib.tracing import init_tracer_provider

app = Flask(__name__)

@app.route("/format")
def format():
    # getting a tracer
    tracer: Tracer = get_tracer("formatter-tracer")
    
    # uncomment the following line to see the request headers
    # print(f"carrier: {request.headers} \nType: {type(request.headers)}")
    
    # extracting the span context from the request headers
    ctx = propagate.extract(carrier=request.headers)
    print("context:", ctx)
    # starting a new span with the extracted context
    with tracer.start_as_current_span('format', context=ctx) as span:
        hello_to = request.args.get('helloTo')
        hello_str = f'Hello, {hello_to}!'
        # adding an event to the span indicating that request for publishing has been completed
        span.add_event('string-format-event', {'string-formatted': hello_str})
        print(span.get_span_context())
        return hello_str


def get_tracer(tracer_name: str):
    # retrieve the global tracer provider
    tracer_provider: TracerProvider = trace.get_tracer_provider()
    # obtain a tracer instance from the tracer provider
    tracer: Tracer = tracer_provider.get_tracer(tracer_name)
    return tracer

def main():
    # setting and retrieving the global tracer provider
    tracer_provider: TracerProvider = init_tracer_provider('formatter')

    # starting the flask app on port 8081
    app.run(port=8081)
    # shutting down the tracer provider to ensure all the spans are exported
    tracer_provider.shutdown()

if __name__ == "__main__":
    main()