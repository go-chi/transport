//go:build go1.21

package transport

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type LogOptions struct {
	Concise bool
	CURL    bool
}

func LogRequests(opts LogOptions) func(next http.RoundTripper) http.RoundTripper {
	return func(next http.RoundTripper) http.RoundTripper {
		return RoundTripFunc(func(req *http.Request) (resp *http.Response, err error) {
			ctx := req.Context()
			r := CloneRequest(req)

			var buf bytes.Buffer
			if opts.CURL && r.Body != nil {
				r.Body = io.NopCloser(io.TeeReader(r.Body, &buf))
			}

			slog.LogAttrs(ctx, slog.LevelDebug, fmt.Sprintf("Send request: %v %s", r.Method, r.URL.String()))

			startTime := time.Now()
			defer func() {
				level := slog.LevelError
				var statusCode int
				if resp != nil {
					statusCode = resp.StatusCode
					if statusCode >= 200 && statusCode < 400 {
						level = slog.LevelInfo
					}
				}

				attrs := []slog.Attr{}
				if err != nil {
					attrs = append(attrs, slog.Any("error", err))
				}
				if opts.CURL {
					attrs = append(attrs, slog.String("curl", curl(r, &buf)))
				}

				if opts.Concise {
					slog.LogAttrs(ctx, level, fmt.Sprintf("Send request: %v %s => HTTP %v (%v)", r.Method, r.URL.String(), statusCode, time.Since(startTime)), attrs...)
				} else {
					attrs = append(attrs,
						slog.String("url", r.URL.String()),
						slog.Duration("duration", time.Since(startTime)),
						slog.Int("status", statusCode),
					)
					slog.LogAttrs(ctx, level, fmt.Sprintf("Send request"), attrs...)
				}
			}()

			return next.RoundTrip(r)
		})
	}
}
