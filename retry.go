package transport

import (
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"time"
)

func Retry(baseTransport http.RoundTripper, maxRetries int) func(http.RoundTripper) http.RoundTripper {
	return func(next http.RoundTripper) http.RoundTripper {
		return RoundTripFunc(func(req *http.Request) (resp *http.Response, err error) {
			defer func() {
				if isRetryable(resp) {
					ctx := req.Context()

					for i := 1; i <= maxRetries; i++ {
						wait := backOff(resp, i)

						timer := time.NewTimer(wait)

						slog.LogAttrs(ctx, slog.LevelDebug, "waiting for backoff",
							slog.String("wait", wait.String()),
							slog.Int("attempt", i),
						)

						select {
						case <-ctx.Done():
							slog.Log(ctx, slog.LevelDebug, "request was cancelled")
							timer.Stop()
							break
						case <-timer.C:
						}

						startTime := time.Now()
						resp, err = baseTransport.RoundTrip(req)
						if !isRetryable(resp) {
							break
						}

						slog.LogAttrs(ctx, slog.LevelWarn, "retrying request",
							slog.Int("attempt", i),
							slog.String("method", req.Method),
							slog.String("url", req.URL.String()),
							slog.String("response.status", resp.Status),
							slog.Duration("response.duration", time.Since(startTime)),
						)
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

func isRetryable(resp *http.Response) bool {
	if resp == nil {
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
