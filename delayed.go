package transport

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

type DelayedConfig struct {
	// RequestDelayMin is the delay before the request is sent.
	RequestDelayMin time.Duration

	// RequestDelayMax is the maximum delay before the request is sent.
	RequestDelayMax time.Duration

	// ResponseDelayMin is the delay before the response is returned.
	ResponseDelayMin time.Duration

	// ResponseDelayMax is the maximum delay before the response is returned.
	ResponseDelayMax time.Duration
}

// Delayed is a middleware that delays requests and responses, useful when
// testing timeouts.
func Delayed(conf DelayedConfig) func(http.RoundTripper) http.RoundTripper {
	if conf.RequestDelayMin > conf.RequestDelayMax {
		panic(fmt.Errorf("connect delay min %v is greater than max %v", conf.RequestDelayMin, conf.RequestDelayMax))
	}

	if conf.ResponseDelayMin > conf.ResponseDelayMax {
		panic(fmt.Errorf("transport delay min %v is greater than max %v", conf.ResponseDelayMin, conf.ResponseDelayMax))
	}

	return func(next http.RoundTripper) http.RoundTripper {
		return RoundTripFunc(func(req *http.Request) (*http.Response, error) {
			ctx := req.Context()

			requestDelay := randDelay(conf.RequestDelayMin, conf.RequestDelayMax)

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
			responseDelay := randDelay(conf.ResponseDelayMin, conf.ResponseDelayMax)
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
