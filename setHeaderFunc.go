package transport

import (
	"net/http"
)

func SetHeaderFunc(header string, fn func(req *http.Request) string) func(http.RoundTripper) http.RoundTripper {
	return func(next http.RoundTripper) http.RoundTripper {
		return RoundTripFunc(func(req *http.Request) (resp *http.Response, err error) {
			r := cloneRequest(req)
			r.Header.Set(http.CanonicalHeaderKey(header), fn(r))

			return next.RoundTrip(r)
		})
	}
}
