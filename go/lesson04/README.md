# Lesson 4 - Baggage

## Objectives

* Understand distributed context propagation
* Use baggage to pass data through the call graph

### Walkthrough

In Lesson 3 we have seen how span context is propagated over the wire between different applications. It is not hard to see that this process can be generalized to propagating more than just the tracing context. With OpenTelemetry instrumentation in place, we can support general purpose _distributed context propagation_ where we associate some metadata with the transaction and make that metadata available anywhere in the distributed call graph. In OpenTelemetry this metadata is called _baggage_, to highlight the fact that it is carried over in-band with all RPC requests, just like baggage.

To see how it works in OpenTracing, let's take the application we built in Lesson 3. You can copy the source code from [../lesson03/solution](../lesson03/solution) package:

```
cp -r ./lesson03/solution ./lesson04/exercise
```

The `formatter` service takes the `helloTo` parameter and returns a string `Hello, {helloTo}!`. Let's modify it so that we can customize the greeting too, but without modifying the public API of that service.

### Set Baggage in the Client

Let's modify the function `main` in `client/hello.go` as follows:

```go
if len(os.Args) != 3 {
    panic("ERROR: Expecting two arguments")
}

greeting := os.Args[2]

// creating baggage items map and add "greeting"
baggageItems := map[string]string{"greeting": greeting}

// calling `formatString` function with the context ctx.
helloStr, err := formatString(ctx, helloTo, baggageItems)
if err != nil {
	log.Fatalf(err.Error())
}
```
Here we read a second command line argument as a "greeting". We also need to modify the signature of the function `formatString` in `client/hello.go` so that it now accepts an additional argument of type `map[string]string`. The `formatString` function in `client/hello.go` will be responsible for propagating the baggage items to the `Formatter` service. Let's modify the function `formatString` in `client/hello.go` as follows:

```go
func formatString(ctx context.Context, helloTo string, baggageItems map[string]string) (string, error) {
	
  // previous code

	// creating baggage members from the baggage items
	baggageMembers := make([]baggage.Member, 0)
	for k, v := range baggageItems {
		bm, err := baggage.NewMember(k, v)
		if err != nil {
			return "", fmt.Errorf("failed to create a new baggage member: %v", err)
		}
		baggageMembers = append(baggageMembers, bm)
	}

	// creating a baggage containing the baggage members
	b, err := baggage.New(baggageMembers...)
	if err != nil {
		return "", fmt.Errorf("failed to create a new baggage: %v", err)
	}

	// adding baggage to the context ctx
	ctx = baggage.ContextWithBaggage(ctx, b)
	// creating a span with the context ctx that contains the baggage, and custom attributes indicating that it is an RPC
	ctx, span := tracer.Start(ctx, "formatString",
		trace.WithAttributes(
			semconv.NetPeerNameKey.String(url),
			semconv.HTTPMethodKey.String("GET"),
		),
		trace.WithSpanKind(trace.SpanKindClient),
	)

	// previous code

}
```

### Read Baggage in Formatter

Add the following code to the `formatter`'s HTTP handler:

```go
// Retrieving baggage items from the context
b := baggage.FromContext(ctx)

// retrieving the member from the baggage with the key "greeting"
greeting := b.Member("greeting").Value()
fmt.Println("from baggage: ", greeting)
if greeting == "" {
	greeting = "Hello"
}

helloTo := r.FormValue("helloTo")
helloStr := fmt.Sprintf("%s, %s!", greeting, helloTo)
```

### Run it

As in Lesson 3, first start the `formatter` and `publisher` in separate terminals, then run the client with two arguments, e.g. `hello.go Brian Bonjour`. The `publisher` should print `Bonjour, Bryan!`.

```
# client
$ go run ./lesson04/exercise/client/hello.go Brian Bonjour
2024/07/10 06:43:09 {"TraceID":"23eedf17aa7bb04365fdc8f547506020","SpanID":"edc33bfde1463ab1","TraceFlags":"01","TraceState":"","Remote":false}
2024/07/10 06:43:09 {"TraceID":"23eedf17aa7bb04365fdc8f547506020","SpanID":"4304752b376af649","TraceFlags":"01","TraceState":"","Remote":false}
2024/07/10 06:43:09 {"TraceID":"23eedf17aa7bb04365fdc8f547506020","SpanID":"f58f0e0e90ad64dd","TraceFlags":"01","TraceState":"","Remote":false}

# formatter
$ go run ./lesson04/exercise/formatter/formatter.go
2024/07/10 06:43:09 {"TraceID":"23eedf17aa7bb04365fdc8f547506020","SpanID":"74fa41b1a45abfb5","TraceFlags":"01","TraceState":"","Remote":false}

# publisher
$ go run ./lesson04/exercise/publisher/publisher.go
Hello, Brian!
2024/07/10 06:43:09 {"TraceID":"23eedf17aa7bb04365fdc8f547506020","SpanID":"6f42754d0ce18d58","TraceFlags":"01","TraceState":"","Remote":false}
```

We see `Hello, Brian!` instead of `Bonjour, Brian!`. That is because the baggage has not been propagated. We set the global propagator to an instance of `propagation.TraceContext` in the function `InitTracerProvider` of our helper library. We need to use a composite propagator that can also propagate the baggage in addition to the trace context. Let's update the `InitTracerProvider` function in our helper library in the file `./lib/tracing/init.go`:

```go
// setting up a composite propagator to handle context propagation (traces and baggage) across services
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
    // for propagating trace context
		propagation.TraceContext{},
    // for propagating the baggage
		propagation.Baggage{},
	))
```
Now let's re-run the application:

```
# client
$ go run ./lesson04/exercise/client/hello.go Brian Bonjour
2024/07/10 06:56:31 {"TraceID":"2248fae14340f8de654ccda4922eb387","SpanID":"e72c1416ab8878d5","TraceFlags":"01","TraceState":"","Remote":false}
2024/07/10 06:56:31 {"TraceID":"2248fae14340f8de654ccda4922eb387","SpanID":"35970ed77a9e1516","TraceFlags":"01","TraceState":"","Remote":false}
2024/07/10 06:56:31 {"TraceID":"2248fae14340f8de654ccda4922eb387","SpanID":"7fc831a8b2352a99","TraceFlags":"01","TraceState":"","Remote":false}

# formatter
$ go run ./lesson04/exercise/formatter/formatter.go
2024/07/10 06:56:31 {"TraceID":"2248fae14340f8de654ccda4922eb387","SpanID":"5f699357c80af5ed","TraceFlags":"01","TraceState":"","Remote":false}

# publisher
$ go run ./lesson04/exercise/publisher/publisher.go
Bonjour, Brian!
2024/07/10 06:56:31 {"TraceID":"2248fae14340f8de654ccda4922eb387","SpanID":"a3d920d82d1e42d9","TraceFlags":"01","TraceState":"","Remote":false}
```

### What's the Big Deal?

We may ask - so what, we could've done the same thing by passing the `greeting` as an HTTP request parameter.
However, that is exactly the point of this exercise - we did not have to change any APIs on the path from
the root span in `hello.go` all the way to the server-side span in `formatter`, three levels down.
If we had a much larger application with much deeper call tree, say the `formatter` was 10 levels down,
the exact code changes we made here would have worked, despite 8 more services being in the path.
If changing the API was the only way to pass the data, we would have needed to modify 8 more services
to get the same effect.

Some of the possible applications of baggage include:

  * passing the tenancy in multi-tenant systems
  * passing identity of the top caller
  * passing fault injection instructions for chaos engineering
  * passing request-scoped dimensions for other monitoring data, like separating metrics for prod vs. test traffic


### Now, a Warning... NOW a Warning?

Of course, while baggage is an extermely powerful mechanism, it is also dangerous. If we store a 1Mb value/string
in baggage, every request in the call graph below that point will have to carry that 1Mb of data. So baggage
must be used with caution. In fact, Jaeger client libraries implement centrally controlled baggage restrictions,
so that only blessed services can put blessed keys in the baggage, with possible restrictions on the value length.

## Conclusion

The complete program can be found in the [solution](./solution) package.
