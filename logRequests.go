package transport

import (
	"net/http"
	"time"
	"log"

	"moul.io/http2curl/v2"
)

type DefaultLogger struct{}

func (*DefaultLogger) Info(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func LogRequests(logger Logger) func(http.RoundTripper) http.RoundTripper {
	if logger == nil {
		logger = &DefaultLogger{}
	}

	return func(next http.RoundTripper) http.RoundTripper {
		return RoundTripFunc(func(req *http.Request) (resp *http.Response, err error) {
			r := cloneRequest(req)

			curlCommand, _ := http2curl.GetCurlCommand(r)
			logger.Info("%v", curlCommand)
			logger.Info("request: %s %s", r.Method, r.URL)

			startTime := time.Now()
			defer func() {
				if resp != nil {
					logger.Info("response (HTTP %v): %v %s", time.Since(startTime), resp.Status, r.URL)
				} else {
					logger.Info("response (<nil>): %v %s", time.Since(startTime), r.URL)
				}
			}()

			return next.RoundTrip(r)
		})
	}
}
