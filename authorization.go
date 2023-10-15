package transport

import (
	"net/http"
)

func Authorization(authorization string) func(http.RoundTripper) http.RoundTripper {
	return func(next http.RoundTripper) http.RoundTripper {
		return RoundTripFunc(func(req *http.Request) (resp *http.Response, err error) {
			req.Header.Set("Authorization", authorization)

			return next.RoundTrip(req)
		})
	}
}
