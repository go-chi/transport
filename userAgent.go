package transport

import (
	"net/http"
)

func UserAgent(userAgent string) Middleware {
	return func(rt http.RoundTripper) http.RoundTripper {
		return RoundTripFunc(func(req *http.Request) (resp *http.Response, err error) {
			req.Header.Set("User-Agent", userAgent)

			return rt.RoundTrip(req)
		})
	}
}
