package transport

import "net/http"

// Middleware is our middleware creation functionality.
type Middleware func(http.RoundTripper) http.RoundTripper

type roundTripper func(r *http.Request) (*http.Response, error)

func (rt roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt(req)
}

// Chain is a handy function to wrap a base RoundTripper (optional)
// with the middlewares.
func Chain(rt http.RoundTripper, middlewares ...Middleware) http.RoundTripper {
	if rt == nil {
		rt = http.DefaultTransport
	}

	for _, m := range middlewares {
		rt = m(rt)
	}

	return rt
}
