package transport

import (
	"log"
	"net/http"
	"time"
)

func RequestTimer() Middleware {
	return func(rt http.RoundTripper) http.RoundTripper {
		return RoundTripFunc(func(req *http.Request) (*http.Response, error) {
			startTime := time.Now()
			defer func() {
				log.Printf(">>> request duration: %s", time.Since(startTime))
			}()

			return rt.RoundTrip(req)
		})
	}
}
