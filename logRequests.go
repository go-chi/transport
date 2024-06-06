package transport

import (
	"log"
	"net/http"
	"time"

	"moul.io/http2curl/v2"
)

func LogRequests(next http.RoundTripper) http.RoundTripper {
	return RoundTripFunc(func(req *http.Request) (resp *http.Response, err error) {
		r := CloneRequest(req)

		curlCommand, _ := http2curl.GetCurlCommand(r)
		log.Printf("%v", curlCommand)
		log.Printf("request: %s %s", r.Method, r.URL)

		startTime := time.Now()
		defer func() {
			if resp != nil {
				log.Printf("response (HTTP %v): %v %s", time.Since(startTime), resp.Status, r.URL)
			} else {
				log.Printf("response (<nil>): %v %s", time.Since(startTime), r.URL)
			}
		}()

		return next.RoundTrip(r)
	})
}
