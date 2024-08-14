# Go HTTP transports & middlewares for outgoing HTTP requests

Chaining transports is a pattern originally inspired by this article https://dev.to/stevenacoffman/tripperwares-http-client-middleware-chaining-roundtrippers-3o00.
This pattern is similar to middleware pattern which is used to enrich a context of http request coming to your application.
There are multiple use-cases where this pattern comes handy such as request logging, caching, authentication and even implementation of retry mechanisms.


## Examples

Set up HTTP client, which sets `User-Agent`, `Authorization` and `TraceID` headers automatically:
```go
authClient := http.Client{
    Transport: transport.Chain(
        http.DefaultTransport,
        transport.SetHeader("User-Agent", userAgent),
        transport.SetHeader("Authorization", authHeader),
        transport.TraceID,
    ),
    Timeout: 15 * time.Second,
}
```

Or debug all outgoing requests globally within your application:
```go
debugMode := os.Getenv("DEBUG") == "true"

http.DefaultTransport = transport.Chain(
    http.DefaultTransport,
    transport.If(debugMode, transport.LogRequests),
)
```

# Authors
- [Golang.cz](https://golang.cz/)
- See [list of contributors](https://github.com/go-chi/transport/graphs/contributors).

# License
[MIT license](./LICENSE)
