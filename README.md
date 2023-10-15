# Go HTTP Transport - middleware for outgoing HTTP requests

## Examples

Set up HTTP client, which sets `User-Agent`, `Authorization` and `TraceID` headers automatically :
```go
httpClient := http.Client{
    Transport: transport.Chain(
        http.DefaultTransport,
        transport.UserAgent("my-app/v1.0.0"),
        transport.Authorization(fmt.Sprintf("BEARER %v", jwt)),
        transport.TraceID,
    ),
    Timeout: 15 * time.Second,
}
```

Or debug all outgoing requests globally within your application:
```go
if debugMode {
    http.DefaultTransport = transport.Chain(
        http.DefaultTransport,
        transport.Debug,
    )
}
```

# Authors
- [Golang.cz](https://golang.cz/)
- See [list of contributors](https://github.com/golang-cz/transport/graphs/contributors).

# License
[MIT license](./LICENSE)
