# Lesson 2 - Context and Tracing Functions

## Objectives

Learn how to:

* Trace individual functions
* Combine multiple spans into a single trace
* Propagate the in-process context

## Walkthrough

First, copy your work or the official solution from [Lesson 1](../lesson01) to `lesson02/exercise/hello.go`.

### Tracing individual functions

In [Lesson 1](../lesson01) we wrote a program that creates a trace that consists of a single span. That single span combined two operations performed by the program, formatting the output string and printing it. Let's move those operations into standalone functions first:

```go
_, span := tracer.Start(ctx, "say-hello")
span.SetAttributes(attribute.String("hello-to", helloTo))
defer span.End()

// calling `formatString` function with the span
helloStr := formatString(span, helloTo)

// calling `printHello` function with the span
printHello(span, helloStr)
```

and the functions:

```go
// Function to format the greeting string
func formatString(span trace.Span, helloTo string) string {
    helloStr := fmt.Sprintf("Hello, %s!", helloTo)
    
    // Add an event to the span with a timestamp
    span.AddEvent("event",
		trace.WithAttributes(attribute.String("string-format", helloStr)), trace.WithTimestamp(time.Now()))

    return helloStr
}


// Function to print the greeting string
func printHello(span trace.Span, helloStr string) {
    println(helloStr)

    // Add an event to the span with a timestamp
	span.AddEvent("event",
		trace.WithAttributes(attribute.String("println", helloStr)), trace.WithTimestamp(time.Now()))
}
```

Of course, this does not change the outcome. What we really want to do is to wrap each function into its own span.

### Wrapping Each Function in Its Own Span

To track the execution of individual functions as separate spans, we need start a new span in each of the functions. Let's modigy the functions as follows:

```go
// Function to format the greeting string
func formatString(ctx context.Context, helloTo string) string {
	// Retrieve or create a named tracer.
	tracer := otel.Tracer("say-hello-tracer")

	// Start a new span named "formatString".
	_, span := tracer.Start(ctx, "formatString")
	defer span.End()

	helloStr := fmt.Sprintf("Hello, %s!", helloTo)

	// adding an event to the span.
	span.AddEvent("event",
		trace.WithAttributes(attribute.String("string-format", helloStr)), trace.WithTimestamp(time.Now()))

	// printing the span contents
	tracing.PrintSpanContents(span)

	return helloStr
}

// Function to print the greeting string
func printHello(ctx context.Context, helloStr string) {
	// Retrieve or create a named tracer
	tracer := otel.Tracer("say-hello-tracer")

	// Start a new span named "printHello"
	_, span := tracer.Start(ctx, "printHello")
	defer span.End()

	println(helloStr)

	// adding an event to the span.
	span.AddEvent("event",
		trace.WithAttributes(attribute.String("println", helloStr)), trace.WithTimestamp(time.Now()))

	// printing the span contents
	tracing.PrintSpanContents(span)
}
```
and the main function:

```go
// starting a new span named "say-hello"
_, span := tracer.Start(ctx, "say-hello")
span.SetAttributes(attribute.String("hello-to", helloTo))
defer span.End()

// calling `formatString` function with a background context
helloStr := formatString(context.Background(), helloTo)

// calling `printHello` function with a background context 
printHello(context.Background(), helloStr)
```


Let's run it:

```
$ go run ./lesson02/exercise/hello.go Brian
2024/07/09 06:24:45 {"TraceID":"6c69832347b64cc1332d7aa6de1dcafb","SpanID":"b5a6927b2d576cdc","TraceFlags":"01","TraceState":"","Remote":false}
Hello, Brian!
2024/07/09 06:24:45 {"TraceID":"e983181b014bf5a143a9380a9ed1679a","SpanID":"a1b7d7bd75914c74","TraceFlags":"01","TraceState":"","Remote":false}
2024/07/09 06:24:45 {"TraceID":"4d408919a64d6cbfa075ef73230c944e","SpanID":"a0d2b78892f808c8","TraceFlags":"01","TraceState":"","Remote":false}
```

We got three spans, but there is a problem here. The TraceIDs are all different. If we search for those IDs in the UI each one will represent a standalone trace with a single span. That's not what we wanted!

### Establishing Parent-Child Relationship

What we really wanted was to establish causal relationship between the two new spans to the root span started in `main`. We can do that by using the context that was returned when we called `tracer.Start` in the main function. This context includes the SpanContext of the parent span. Any new span created using this context will be a child of the parent span, inheriting its TraceID and establishing a parent-child relationship. So the main function needs to be updated to:

```go
// startin a new span named "say-hello" and getting the context
ctx, span := tracer.Start(ctx, "say-hello")
span.SetAttributes(attribute.String("hello-to", helloTo))
defer span.End()

// calling `formatString` function with the context "ctx"
helloStr := formatString(ctx, helloTo)

// calling `printHello` function with the context "ctx"
printHello(ctx, helloStr)
```

On running the app, we'll see that all reported spans now belong to the same trace:

```
$ go run ./lesson02/exercise/hello.go Brian
2024/07/09 06:29:02 {"TraceID":"d4631a54c578fd7644512d66e904e0c5","SpanID":"2597d8c8c39dc7f8","TraceFlags":"01","TraceState":"","Remote":false}
Hello, Brian!
2024/07/09 06:29:02 {"TraceID":"d4631a54c578fd7644512d66e904e0c5","SpanID":"8c37823820606d06","TraceFlags":"01","TraceState":"","Remote":false}
2024/07/09 06:29:02 {"TraceID":"d4631a54c578fd7644512d66e904e0c5","SpanID":"2ed001eeb94e4653","TraceFlags":"01","TraceState":"","Remote":false}
```

If we find this trace in the UI, it will show a proper parent-child relationship between the spans.

## Conclusion

The complete program can be found in the [solution](./solution) package.

Next lesson: [Tracing RPC Requests](../lesson03).