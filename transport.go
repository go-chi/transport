package transport

import "net/http"

type RoundTripFunc func(r *http.Request) (*http.Response, error)

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type chain struct {
	rt          http.RoundTripper
	middlewares []func(http.RoundTripper) http.RoundTripper
}

func (c *chain) RoundTrip(req *http.Request) (*http.Response, error) {
	rt := c.rt

	// Apply middlewares in reversed order, so if the following come in:
	// [Auth, VctraceId, Debug]
	// then they are applied in this order:
	// rt = Debug(rt)
	// rt = VctraceId(rt)
	// rt = Auth(rt)
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		rt = c.middlewares[i](rt)
	}

	return rt.RoundTrip(req)
}

// Chain wraps http.DefaultTransport with extra RoundTripper middlewares.
func Chain(rt http.RoundTripper, middlewares ...func(http.RoundTripper) http.RoundTripper) *chain {
	if rt == nil {
		rt = http.DefaultTransport
	}

	if c, ok := rt.(*chain); ok {
		c.middlewares = append(c.middlewares, middlewares...)
		return c
	}

	return &chain{
		rt:          rt,
		middlewares: middlewares,
	}
}
