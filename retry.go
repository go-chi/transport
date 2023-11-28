package transport

import (
	"net/http"
	"net/url"
	"log"
	"time"
	"strconv"
	"math"
)

func Retry(baseTransport http.RoundTripper, maxRetries int) func(http.RoundTripper) http.RoundTripper {
	return func(next http.RoundTripper) http.RoundTripper {
		return RoundTripFunc(func(req *http.Request) (resp *http.Response, err error) {
			defer func() {
				if err == nil || !isRetryable(err, resp) {
					for i := 1; i <= maxRetries; i++ {
						wait := backOff(resp, i)

						timer := time.NewTimer(wait)

						log.Printf("waiting %s", wait.String())

						select {
						case <-req.Context().Done():
							log.Printf("request was cancelled")
							timer.Stop()
							break
						case <-timer.C:
						}

						startTime := time.Now()
						resp, err = baseTransport.RoundTrip(req)
						if err == nil || !isRetryable(err, resp) {
							break
						}

						log.Printf("retrying %d request: %s %s", i, req.Method, req.URL)
						log.Printf("response (%v): %v %s", time.Since(startTime), resp.Status, resp.Request.URL)
					}
				}
			}()

			return next.RoundTrip(req)
		})
	}
}

func backOff(resp *http.Response, attempt int) time.Duration {
	minDuration := 1 * time.Second
	maxDuration := 16 * time.Second

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable {
		if s, ok := resp.Header["Retry-After"]; ok {
			if sleep, err := strconv.ParseInt(s[0], 10, 64); err == nil {
				return time.Second * time.Duration(sleep)
			}
		}
	}

	// simple exp. backoff
	mult := math.Pow(2, float64(attempt)) * float64(minDuration)
	sleep := time.Duration(mult)
	if float64(sleep) != mult || sleep > maxDuration {
		sleep = maxDuration
	}
	return sleep
}

func isRetryable(err error, resp *http.Response) bool {
	if resp == nil {
		return false
	}

	// any error returned from Client.Do will be *url.Error
	if _, ok := err.(*url.Error); ok {
		return false
	}

	// 429 Too Many Requests is recoverable. Sometimes the server puts
	// Retry-After response header to indicate when the server is will be available again
	if resp.StatusCode == http.StatusTooManyRequests {
		return true
	}

	// We retry on 500-range responses to allow the server time to recover, as 500's are typically not permanent
	// errors and may relate to outages on the server side.
	if resp.StatusCode == 0 || (resp.StatusCode >= 500 && resp.StatusCode != http.StatusNotImplemented) {
		return true
	}

	return false
}
