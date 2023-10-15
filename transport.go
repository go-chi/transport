package transport

import "net/http"

// RoundTripFunc, similar to http.HandlerFunc, is an adapter
// to allow the use of ordinary functions as http.RoundTrippers.
type RoundTripFunc func(r *http.Request) (*http.Response, error)

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// Chain wraps given base RoundTripper, which is used to make HTTP requests
// (e.g. http.DefaultTransport) with RoundTripper middlewares.
//
// The middlewares can print, debug or modify request/response headers,
// cookies, context timeouts etc.
//
// Note: Per stdlib docs, RoundTrip should not modify the original request,
// except for consuming and closing the request's body. Thus, it's advised
// to clone the original request before modifying it, e.g. golang.org/x/oauth2:
// https://cs.opensource.google/go/x/oauth2/+/refs/tags/v0.13.0:transport.go;l=50.
//
// A typical use case is to set User-Agent, Authorization or TraceID headers:
//
//	authClient := http.Client{
//	    Transport: transport.Chain(
//	        http.DefaultTransport,
//	        transport.UserAgent("my-app/v1.0.0"),
//	        transport.Authorization(fmt.Sprintf("BEARER %v", jwt)),
//	        transport.TraceID,
//	    ),
//	    Timeout: 15 * time.Second,
//	}
//
// Or debug all outgoing requests in a debug mode:
//
//	http.DefaultTransport = transport.Chain(
//		http.DefaultTransport,
//		transport.DebugRequests,
//	)
func Chain(base http.RoundTripper, mw ...func(http.RoundTripper) http.RoundTripper) *chain {
	if base == nil {
		base = http.DefaultTransport
	}

	if c, ok := base.(*chain); ok {
		c.middlewares = append(c.middlewares, mw...)
		return c
	}

	return &chain{
		baseTransport: base,
		middlewares:   mw,
	}
}

type chain struct {
	baseTransport http.RoundTripper
	middlewares   []func(http.RoundTripper) http.RoundTripper
}

func (c *chain) RoundTrip(req *http.Request) (*http.Response, error) {
	rt := c.baseTransport

	// Apply middlewares in reversed order so the first middleware becomes
	// the innermost onion layer and the last becomes the outermost. Example:
	// Given
	//   [Auth, VCTraceID, Debug],
	// the middlewares are applied in this order:
	//   rt = Debug(rt)
	//   rt = VCTraceID(rt)
	//   rt = Auth(rt)
	// The Auth and VCTraceID are called before the Debug middleware,
	// which can then see the final request headers, as seen by http.DefaultTransport.
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		rt = c.middlewares[i](rt)
	}

	return rt.RoundTrip(req)
}
