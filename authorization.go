package transport

import (
	"net/http"
)

func Authorization(authorization string) func(http.RoundTripper) http.RoundTripper {
	return SetHeader("Authorization", authorization)
}
