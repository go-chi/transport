package transport

import (
	"bytes"
	"io"
	"log"
	"net/http"

	"moul.io/http2curl/v2"
)

func Debug() Middleware {
	return func(rt http.RoundTripper) http.RoundTripper {
		return roundTripper(func(req *http.Request) (resp *http.Response, err error) {
			defer func() {
				c, _ := http2curl.GetCurlCommand(req)

				log.Printf("request: %s %s", req.Method, req.URL)
				if err == nil {
					log.Printf("response: %s %s", resp.Status, resp.Request.URL)
				}

				log.Printf("%s", c.String())
			}()

			return rt.RoundTrip(req)
		})
	}
}

func DebugRequestBody() Middleware {
	return func(rt http.RoundTripper) http.RoundTripper {
		return roundTripper(func(req *http.Request) (resp *http.Response, err error) {
			buf, err := io.ReadAll(req.Body)
			if err != nil {
				log.Printf("error copying req.Body: %v", err)
			}
			req.Body = io.NopCloser(bytes.NewBuffer(buf))

			log.Printf("request body (len %v):\n%s", len(buf), string(buf))

			return rt.RoundTrip(req)
		})
	}
}
