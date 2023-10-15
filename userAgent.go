package transport

import (
	"net/http"
)

func UserAgent(userAgent string) func(http.RoundTripper) http.RoundTripper {
	return SetHeader("User-Agent", userAgent)
}
