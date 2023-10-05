package transport

import (
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
