package transport

import (
	"net/http"
)

func SetHeaderFunc(header string, fn func(req *http.Request) string) func(http.RoundTripper) http.RoundTripper {
	return func(next http.RoundTripper) http.RoundTripper {
		return RoundTripFunc(func(req *http.Request) (resp *http.Response, err error) {
			req.Header.Set(http.CanonicalHeaderKey(header), fn(req))

			return next.RoundTrip(req)
		})
	}
}
