package transport

import (
	"github.com/vcilabs/pkg/utils"
	"log"
	"net/http"
)

func Debug() Middleware {
	return func(rt http.RoundTripper) http.RoundTripper {
		return roundTripper(func(req *http.Request) (resp *http.Response, err error) {
			defer func() {
				c, _ := utils.GetCurlCommand(req)

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
