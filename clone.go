package transport

import "net/http"

// cloneRequest creates a shallow copy of a given request
// to comply with stdlib's http.RoundTripper contract:
//
// RoundTrip should not modify the request, except for
// consuming and closing the Request's Body. RoundTrip may
// read fields of the request in a separate goroutine. Callers
// should not mutate or reuse the request until the Response's
// Body has been closed.
func cloneRequest(orig *http.Request) *http.Request {
	clone := &http.Request{}
	*clone = *orig

	clone.Header = make(http.Header, len(orig.Header))
	for key, value := range orig.Header {
		clone.Header[key] = append([]string{}, value...)
	}

	return clone
}
