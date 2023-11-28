package transport

import (
	"crypto/x509"
	"net/http"
	"net/url"
	"regexp"
	"log"
	"time"
	"strconv"
	"math"
)

var (
	// A regular expression to match the error returned by net/http when the
	// configured number of redirects is exhausted. This error isn't typed
	// specifically so we resort to matching on the error string.
	tooManyRedirectsRe = regexp.MustCompile(`stopped after \d+ redirects\z`)

	// A regular expression to match the error returned by net/http when the
	// scheme specified in the URL is invalid. This error isn't typed
	// specifically so we resort to matching on the error string.
	invalidSchemeRe = regexp.MustCompile(`unsupported protocol scheme`)

	// A regular expression to match the error returned by net/http when the
	// TLS certificate is not trusted. This error isn't typed
	// specifically so we resort to matching on the error string.
	untrustedCertificateRe = regexp.MustCompile(`certificate is not trusted`)
)

func Retry(baseTransport http.RoundTripper, maxRetries int) func(http.RoundTripper) http.RoundTripper {
	return func(next http.RoundTripper) http.RoundTripper {
		return RoundTripFunc(func(req *http.Request) (resp *http.Response, err error) {
			defer func() {
				if isRetryable(err, resp) {
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
						if isRetryable(err, resp) {
							log.Printf("retrying %d request: %s %s", i, req.Method, req.URL)
							log.Printf("response (%v): %v %s", time.Since(startTime), resp.Status, resp.Request.URL)
							continue
						} else {
							break
						}
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
	if serverErr, ok := err.(*url.Error); ok {
		// Too many redirects.
		if tooManyRedirectsRe.MatchString(serverErr.Error()) {
			return false
		}

		// Invalid protocol scheme.
		if invalidSchemeRe.MatchString(serverErr.Error()) {
			return false
		}

		// TLS cert verification failure.
		if untrustedCertificateRe.MatchString(serverErr.Error()) {
			return false
		}

		if _, ok := serverErr.Err.(x509.UnknownAuthorityError); ok {
			return false
		}
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
