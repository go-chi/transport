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

				log.Printf("--> %s %s", req.Method, req.URL)
				log.Printf("%s", c.String())

				if err == nil {
					log.Printf("<-- %d %s", resp.StatusCode, resp.Request.URL)
				}
			}()

			return rt.RoundTrip(req)
		})
	}
}
