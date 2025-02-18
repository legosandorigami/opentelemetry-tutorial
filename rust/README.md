# OpenTelemetry Tutorial - Rust

## Installing

These tutorials can be followed using any following tracing backends:
* [Signoz](https://signoz.io/)
* [Tempo with Grafana](https://grafana.com/docs/tempo/latest/)
* [Jaeger](https://jaegertracing.io)

Refer to this [guide](../README.md) to learn how to set up and run any of the above tracing backends.


## Lessons

* [Lesson 01 - Hello World](./lesson01)
  * Instantiate a Tracer
  * Create a simple trace
  * Annotate the trace
* [Lesson 02 - Context and Tracing Functions](./lesson02)
  * Trace individual functions
  * Combine multiple spans into a single trace
  * Propagate the in-process context
* [Lesson 03 - Tracing RPC Requests](./lesson03)
  * Trace a transaction across more than one microservice
  * Pass the context between processes using `inject_context` and `extract`
  * Apply OpenTracing-recommended tags
* [Lesson 04 - Baggage](./lesson04)
  * Understand distributed context propagation
  * Use baggage to pass data through the call graph
