# OpenTelemetry Tutorials

This repository is a fork of the original [opentracing-tutorial](https://github.com/yurishkuro/opentracing-tutorial) updated to use [OpenTelemetry](https://opentelemetry.io) API. In these tutorials, we focus on sending traces using the OTLP protocol. Many modern observability backends, such as Signoz, Tempo and Jaeger, support the OTLP (OpenTelemetry Protocol) format either natively or through intermediaries. Instrumenting applications with the OTLP exporter is a flexible approach that allows you to switch backends with minimal changes to the application code.

## Tutorials by Language

- [x] [Go tutorial](./go/)
- [x] [Python tutorial](./python)
- [ ] Node.js tutorial
- [ ] Java tutorial
- [ ] C# tutorial

## Prerequisites

These tutorials can be followed along using any of the following tracing backends:
* [Signoz](https://signoz.io/)
* [Tempo with Grafana](https://grafana.com/docs/tempo/latest/)
* [Jaeger](https://jaegertracing.io)

## Backend Setup

### Setting up [Signoz](https://signoz.io/)

Signoz, a distributed tracing backend by [Signoz.io](https://signoz.io/), natively supports OTLP. You can directly send OTLP traces to Signoz without an intermediary.
The easiest way to set up Signoz is using docker-compose. Refer to their [guide](https://signoz.io/docs/install/docker/) for detailed instructions.


### Setting up [Tempo with Grafana](https://grafana.com/docs/tempo/latest/)

Tempo, a distributed tracing backend by [Grafana](https://grafana.com/), natively supports OTLP. You can directly send OTLP traces to Tempo without an intermediary.
For an easy setup, use docker-compose. Refer to their [guide](https://grafana.com/docs/tempo/next/getting-started/docker-example/) for detailed instructions.


### Setting up [Jaeger](https://www.jaegertracing.io/docs/1.59/)

Jaeger is an open-source distributed tracing platform originally developed by [Uber Technologies](http://uber.github.io/). It supports the OpenTelemetry Protocol (OTLP) natively, allowing you to send traces directly to Jaeger without the need for an intermediary.

#### Example: Configure the OTLP exporter to send data directly to Jaeger.
To get started with Jaeger using Docker and the default in-memory storage, follow these steps:

* Start Jaeger via Docker with the necessary ports exposed:
    ```
    docker run -d --name jaeger \
      -e COLLECTOR_OTLP_ENABLED=true \
      -p 16686:16686 \  # Jaeger UI
      -p 4318:4318 \  # OTLP HTTP port
      jaegertracing/all-in-one:latest
    ```
    This command runs Jaeger with the OTLP HTTP port (4318) and Jaeger UI (16686) exposed on localhost.

* Once the backend starts, the Jaeger UI will be accessible at http://localhost:16686.


## Contributions
My main goal was to keep this tutorial as close as possible to the original. Please send me your suggestions and feedback for improving it. I would encourage you to send me pull requests for other languages such as Java, Node.js, and C#. 

---
### original
---

# OpenTracing Tutorials

A collection of tutorials for the OpenTracing API (https://opentracing.io).

**Update (Dec 8 2022)**: Since OpenTracing has been officially retired, I have archived this repository. The tutorials here are still useful when learning about distributed tracing, but using the [OpenTelemetry](https://opentelemetry.io/) API should be preferred over the OpenTracing API for new applications.

The blog post ["Migrating from Jaeger client to OpenTelemetry SDK"](https://medium.com/jaegertracing/migrating-from-jaeger-client-to-opentelemetry-sdk-bd337d796759) can also be used as a reference on how to use the OpenTelemetry SDK as an OpenTracing tracer implementation.

## Tutorials by Language

  * [C# tutorial](./csharp/)
  * [Go tutorial](./go/)
  * [Java tutorial](./java)
  * [Python tutorial](./python)
  * [Node.js tutorial](./nodejs)

Also check out examples from the book [Mastering Distributed Tracing](https://www.shkuro.com/books/2019-mastering-distributed-tracing/):
* [Chapter 4: Instrumentation Basics with OpenTracing](https://github.com/PacktPublishing/Mastering-Distributed-Tracing/tree/master/Chapter04)
* [Chapter 5: Instrumentation of Asynchronous Applications](https://github.com/PacktPublishing/Mastering-Distributed-Tracing/tree/master/Chapter05)
* [Chapter 7: Tracing with Service Mesh](https://github.com/PacktPublishing/Mastering-Distributed-Tracing/tree/master/Chapter07)
* [Chapter 11: Integration with Metrics and Logs](https://github.com/PacktPublishing/Mastering-Distributed-Tracing/tree/master/Chapter11)
* [Chapter 12: Gathering Insights Through Data Mining](https://github.com/PacktPublishing/Mastering-Distributed-Tracing/tree/master/Chapter12)

## Prerequisites

The tutorials are using CNCF Jaeger (https://jaegertracing.io) as the tracing backend.
For this tutorial, we'll start Jaeger via Docker with the default in-memory storage, exposing only the required ports. We'll also enable "debug" level logging:

```
docker run \
  --rm \
  -p 6831:6831/udp \
  -p 6832:6832/udp \
  -p 16686:16686 \
  jaegertracing/all-in-one:1.7 \
  --log-level=debug
```

Alternatively, Jaeger can be downloaded as a binary called `all-in-one` for different platforms from https://jaegertracing.io/download/.

Once the backend starts, the Jaeger UI will be accessible at http://localhost:16686.
