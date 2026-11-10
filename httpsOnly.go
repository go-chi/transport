package transport

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrNotHTTPS is returned by HTTPSOnly when a request is not https.
var ErrNotHTTPS = errors.New("transport: non-https request refused")

// HTTPSOnly refuses any request whose URL scheme is not https, before it
// reaches the next RoundTripper.
//
// http.Client sends every redirect hop through its Transport, so with this
// middleware a server can't downgrade a request to plain http by redirecting.
// Without it, a 307 or 308 replays the request body, and a same-host 301, 302
// or 303 keeps the Authorization header, over the cleartext hop.
//
// The returned error wraps ErrNotHTTPS and names only the scheme and host, as
// the path and query may carry secrets. http.Client wraps it in a *url.Error
// that includes the full URL, so strip that before logging if it matters.
//
// Example:
//
//	client := http.Client{
//	    Transport: transport.Chain(
//	        http.DefaultTransport,
//	        transport.HTTPSOnly,
//	    ),
//	}
func HTTPSOnly(next http.RoundTripper) http.RoundTripper {
	return RoundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Scheme != "https" {
			if req.Body != nil {
				// A RoundTripper must close the body even when it returns an error.
				_ = req.Body.Close()
			}
			return nil, fmt.Errorf("%w: %s://%s", ErrNotHTTPS, req.URL.Scheme, req.URL.Host)
		}
		return next.RoundTrip(req)
	})
}
