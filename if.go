package transport

import (
	"net/http"
)

// If sets given transport if given condition is true. Otherwise it sets nil transport, which will be ignored.
//
// Example:
//
//	http.DefaultTransport = transport.Chain(
//	  http.DefaultTransport,
//	  transport.If(debugMode, transport.LogRequests),
//	)
func If(condition bool, transport func(http.RoundTripper) http.RoundTripper) func(http.RoundTripper) http.RoundTripper {
	if condition {
		return transport
	}
	return nil
}
