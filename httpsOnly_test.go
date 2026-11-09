package transport_test

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/go-chi/transport"
)

type recorded struct {
	scheme string
	auth   string
	body   string
}

// recorder is a base RoundTripper that never opens a socket.
type recorder struct {
	requests []recorded
	respond  func(*http.Request) *http.Response
}

func (r *recorder) RoundTrip(req *http.Request) (*http.Response, error) {
	var body string
	if req.Body != nil {
		b, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		req.Body.Close()
		body = string(b)
	}
	r.requests = append(r.requests, recorded{
		scheme: req.URL.Scheme,
		auth:   req.Header.Get("Authorization"),
		body:   body,
	})
	return r.respond(req), nil
}

func response(req *http.Request, status int, location string) *http.Response {
	h := http.Header{}
	if location != "" {
		h.Set("Location", location)
	}
	return &http.Response{
		Request:    req,
		StatusCode: status,
		Header:     h,
		Body:       io.NopCloser(strings.NewReader("")),
	}
}

func newClient(rt http.RoundTripper) *http.Client {
	return &http.Client{Transport: transport.Chain(rt, transport.HTTPSOnly)}
}

func TestHTTPSOnly(t *testing.T) {
	t.Run("error omits path and query", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "http://example.com/secret?token=abc", nil)
		if err != nil {
			t.Fatal(err)
		}
		_, err = transport.HTTPSOnly(&recorder{}).RoundTrip(req)
		if !errors.Is(err, transport.ErrNotHTTPS) {
			t.Fatalf("got %v, want ErrNotHTTPS", err)
		}
		if strings.Contains(err.Error(), "secret") || strings.Contains(err.Error(), "token=abc") {
			t.Fatalf("error leaks the path or query: %v", err)
		}
	})

	t.Run("allows https", func(t *testing.T) {
		rec := &recorder{respond: func(req *http.Request) *http.Response {
			return response(req, http.StatusOK, "")
		}}
		resp, err := newClient(rec).Get("https://example.com/")
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if len(rec.requests) != 1 {
			t.Fatalf("got %d requests, want 1", len(rec.requests))
		}
	})

	t.Run("refuses http before dispatch", func(t *testing.T) {
		rec := &recorder{respond: func(req *http.Request) *http.Response {
			return response(req, http.StatusOK, "")
		}}
		_, err := newClient(rec).Get("http://example.com/secret?token=abc")
		if !errors.Is(err, transport.ErrNotHTTPS) {
			t.Fatalf("got %v, want ErrNotHTTPS", err)
		}
		if len(rec.requests) != 0 {
			t.Fatalf("got %d requests, want 0", len(rec.requests))
		}
	})

	t.Run("refuses http redirect hops", func(t *testing.T) {
		for _, status := range []int{
			http.StatusMovedPermanently,
			http.StatusFound,
			http.StatusSeeOther,
			http.StatusTemporaryRedirect,
			http.StatusPermanentRedirect,
		} {
			rec := &recorder{respond: func(req *http.Request) *http.Response {
				if req.URL.Scheme == "https" {
					return response(req, status, "http://example.com/next")
				}
				return response(req, http.StatusOK, "")
			}}
			req, err := http.NewRequest(http.MethodPost, "https://example.com/token", strings.NewReader("secret=1"))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Authorization", "Bearer abc")

			_, err = newClient(rec).Do(req)
			if !errors.Is(err, transport.ErrNotHTTPS) {
				t.Fatalf("%d: got %v, want ErrNotHTTPS", status, err)
			}
			for _, r := range rec.requests {
				if r.scheme != "https" {
					t.Fatalf("%d: http request reached the base transport: %+v", status, r)
				}
			}
			if len(rec.requests) != 1 {
				t.Fatalf("%d: got %d requests, want 1", status, len(rec.requests))
			}
		}
	})

	t.Run("follows https redirects", func(t *testing.T) {
		rec := &recorder{respond: func(req *http.Request) *http.Response {
			if req.URL.Path == "/start" {
				return response(req, http.StatusFound, "https://example.com/next")
			}
			return response(req, http.StatusOK, "")
		}}
		resp, err := newClient(rec).Get("https://example.com/start")
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if len(rec.requests) != 2 {
			t.Fatalf("got %d requests, want 2", len(rec.requests))
		}
	})

	t.Run("closes the body on refusal", func(t *testing.T) {
		body := &closeTracker{Reader: strings.NewReader("x")}
		req, err := http.NewRequest(http.MethodPost, "http://example.com/", body)
		if err != nil {
			t.Fatal(err)
		}
		rt := transport.HTTPSOnly(&recorder{})
		if _, err := rt.RoundTrip(req); !errors.Is(err, transport.ErrNotHTTPS) {
			t.Fatalf("got %v, want ErrNotHTTPS", err)
		}
		if !body.closed {
			t.Fatal("request body was not closed")
		}
	})
}

type closeTracker struct {
	io.Reader
	closed bool
}

func (c *closeTracker) Close() error {
	c.closed = true
	return nil
}
