package transport

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

// DelayedRequest is a middleware that delays requests, useful when testing
// timeouts while waiting on a request to be sent upstream.
func DelayedRequest(requestDelayMin, requestDelayMax time.Duration) func(http.RoundTripper) http.RoundTripper {
	if requestDelayMin > requestDelayMax {
		panic(fmt.Sprintf("requestDelayMin %v is greater than requestDelayMax %v", requestDelayMin, requestDelayMax))
	}
	return delayedRoundTripper(randDelay(requestDelayMin, requestDelayMax), 0)
}

// DelayedResponse is a middleware that delays responses, useful when testing
// timeouts after upstream has processed the request, the response is hold back
// until the delay is over.
func DelayedResponse(responseDelayMin, responseDelayMax time.Duration) func(http.RoundTripper) http.RoundTripper {
	if responseDelayMin > responseDelayMax {
		panic(fmt.Sprintf("responseDelayMin %v is greater than responseDelayMax %v", responseDelayMin, responseDelayMax))
	}
	return delayedRoundTripper(0, randDelay(responseDelayMin, responseDelayMax))
}

func delayedRoundTripper(requestDelay, responseDelay time.Duration) func(http.RoundTripper) http.RoundTripper {
	return func(next http.RoundTripper) http.RoundTripper {
		return RoundTripFunc(func(req *http.Request) (*http.Response, error) {
			ctx := req.Context()

			// wait before sending request
			if requestDelay > 0 {
				ticker := time.NewTicker(requestDelay)
				defer ticker.Stop()

				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-ticker.C:
				}
			}

			res, err := next.RoundTrip(req)

			// wait before sending response body
			if responseDelay > 0 {
				ticker := time.NewTicker(responseDelay)
				defer ticker.Stop()

				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-ticker.C:
				}
			}

			return res, err
		})
	}
}

func randDelay(min, max time.Duration) time.Duration {
	if min >= max {
		return min
	}
	return min + time.Duration(rand.Int63n(int64(max-min)))
}
