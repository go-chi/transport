package transport

import (
	"log"
	"net/http"
	"time"

	"moul.io/http2curl/v2"
)

func Debug() func(http.RoundTripper) http.RoundTripper {
	return func(next http.RoundTripper) http.RoundTripper {
		return RoundTripFunc(func(req *http.Request) (resp *http.Response, err error) {
			curlCommand, _ := http2curl.GetCurlCommand(req)
			log.Printf("%v", curlCommand)
			log.Printf("request: %s %s", req.Method, req.URL)

			startTime := time.Now()
			defer func() {
				log.Printf("response (%v): %v %s", time.Since(startTime), resp.Status, resp.Request.URL)
			}()

			return next.RoundTrip(req)
		})
	}
}
