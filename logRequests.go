package transport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"moul.io/http2curl/v2"
)

type RequestLogger interface {
	LogRequest(req *http.Request, curl *http2curl.CurlCommand)
	LogResponse(r *http.Request, resp *http.Response, startTime time.Time)
}

func LogRequests(logger RequestLogger) func(next http.RoundTripper) http.RoundTripper {
	return func(next http.RoundTripper) http.RoundTripper {
		return RoundTripFunc(func(req *http.Request) (resp *http.Response, err error) {
			r := CloneRequest(req)

			curlCommand, _ := http2curl.GetCurlCommand(r)

			logger.LogRequest(req, curlCommand)

			startTime := time.Now()
			defer func() {
				logger.LogResponse(r, resp, startTime)
			}()

			return next.RoundTrip(r)
		})
	}
}

type DefaultLogger struct {
	PrintResponsePayload bool
}

func (d *DefaultLogger) LogRequest(r *http.Request, curl *http2curl.CurlCommand) {
	log.Printf(curl.String())
	log.Printf("request: %s %s", r.Method, r.URL)
}

func (d *DefaultLogger) LogResponse(r *http.Request, resp *http.Response, startTime time.Time) {
	if resp == nil {
		log.Printf(fmt.Sprintf("response (<nil>): %v %s", time.Since(startTime), r.URL))
		return
	}

	log.Printf("response: %s %s", resp.Status, resp.Request.URL)

	if d.PrintResponsePayload && resp.Header.Get("Content-Type") == "application/json" {
		var b bytes.Buffer

		tee := io.TeeReader(resp.Body, &b)
		resp.Body = io.NopCloser(&b)

		payload, err := io.ReadAll(tee)
		if err == nil {
			// Pretty print the JSON payload
			var prettyJSON bytes.Buffer
			if err := json.Indent(&prettyJSON, payload, "", "    "); err == nil {
				log.Printf("%s", prettyJSON.String())
			}
		}
	}
}
