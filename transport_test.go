package transport_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/transport"
)

func TestChain(t *testing.T) {
	expected := "ok"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "transport-chain/v1.0.0" {
			w.WriteHeader(500)
			return
		}

		if r.Header.Get("Accept-Encoding") != "gzip" {
			w.WriteHeader(500)
			return
		}

		fmt.Fprintf(w, expected)
	}))
	defer server.Close()

	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: transport.Chain(
			nil,
			transport.SetHeader("User-Agent", "transport-chain/v1.0.0"),
			transport.SetHeader("Accept-Encoding", "gzip"),
		),
	}

	request, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Fatal("expected some header, but did not receive")
	}
}

func TestChainWithRetries(t *testing.T) {
	expected := "ok"
	retries := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		retries++

		if retries < 2 {
			w.WriteHeader(502)
			return
		}

		fmt.Fprintf(w, expected)
	}))
	defer server.Close()

	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: transport.Chain(
			http.DefaultTransport,
			transport.Retry(http.DefaultTransport, 5),
			transport.LogRequests(transport.LogOptions{}),
		),
	}

	request, err := http.NewRequest("GET", fmt.Sprintf("%s", server.URL), nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Fatal("expected some header, but did not receive")
	}
}

func TestChainWithRetryAfter(t *testing.T) {
	expected := "ok"
	retries := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		retries++

		w.Header().Add("Retry-After", "1")

		if retries < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		fmt.Fprintf(w, expected)
	}))
	defer server.Close()

	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: transport.Chain(
			http.DefaultTransport,
			transport.Retry(http.DefaultTransport, 5),
			transport.LogRequests(transport.LogOptions{}),
		),
	}

	request, err := http.NewRequest("GET", fmt.Sprintf("%s", server.URL), nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Fatal("expected some header, but did not receive")
	}
}

// echoHeaders is a base transport that returns the request headers it received
// as the response headers.
var echoHeaders = transport.RoundTripFunc(func(req *http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: 200, Header: req.Header.Clone(), Body: http.NoBody, Request: req}, nil
})

func roundTripHeaders(t *testing.T, rt http.RoundTripper) http.Header {
	t.Helper()
	resp, err := rt.RoundTrip(httptest.NewRequest("GET", "http://example.com", nil))
	if err != nil {
		t.Error(err)
		return nil
	}
	return resp.Header
}

func TestChainDoesNotModifyChainBase(t *testing.T) {
	// Three middlewares leave spare capacity in the base slice, so a copy that
	// aliases it would let a and b overwrite each other's middleware.
	shared := transport.Chain(echoHeaders,
		transport.SetHeader("X-Base", "1"),
		transport.SetHeader("X-Env", "test"),
		transport.SetHeader("X-Region", "eu"),
	)

	a := transport.Chain(shared, transport.SetHeader("X-Who", "a"))
	b := transport.Chain(shared, transport.SetHeader("X-Who", "b"))

	base := http.Header{"X-Base": {"1"}, "X-Env": {"test"}, "X-Region": {"eu"}}
	withWho := func(who string) http.Header {
		h := base.Clone()
		h.Set("X-Who", who)
		return h
	}
	for _, tt := range []struct {
		name string
		rt   http.RoundTripper
		want http.Header
	}{
		{"a", a, withWho("a")},
		{"b", b, withWho("b")},
		{"shared", shared, base},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := roundTripHeaders(t, tt.rt); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("headers = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChainReturnsNewChain(t *testing.T) {
	shared := transport.Chain(echoHeaders, transport.SetHeader("X-Base", "1"))

	if transport.Chain(shared, transport.SetHeader("X-Who", "a")) == shared {
		t.Fatal("Chain returned its base instead of a new chain")
	}
}

func TestChainConcurrentWithRoundTrip(t *testing.T) {
	shared := transport.Chain(echoHeaders, transport.SetHeader("X-Base", "1"))

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			who := fmt.Sprint(i)
			got := roundTripHeaders(t, transport.Chain(shared, transport.SetHeader("X-Who", who)))
			if got.Get("X-Who") != who {
				t.Errorf("X-Who = %q, want %q", got.Get("X-Who"), who)
			}
		}(i)
		go func() {
			defer wg.Done()
			if got := roundTripHeaders(t, shared); got.Get("X-Who") != "" {
				t.Errorf("shared chain sent X-Who = %q", got.Get("X-Who"))
			}
		}()
	}
	wg.Wait()
}

func TestChainMiddlewareOrder(t *testing.T) {
	var calls []string
	record := func(name string) func(http.RoundTripper) http.RoundTripper {
		return func(next http.RoundTripper) http.RoundTripper {
			return transport.RoundTripFunc(func(req *http.Request) (*http.Response, error) {
				calls = append(calls, name)
				return next.RoundTrip(req)
			})
		}
	}
	want := []string{"1", "2", "3"}

	for _, tt := range []struct {
		name string
		rt   http.RoundTripper
	}{
		{"plain base", transport.Chain(echoHeaders, record("1"), record("2"), record("3"))},
		{"chain base", transport.Chain(transport.Chain(echoHeaders, record("1")), record("2"), record("3"))},
	} {
		t.Run(tt.name, func(t *testing.T) {
			calls = nil
			roundTripHeaders(t, tt.rt)
			if !reflect.DeepEqual(calls, want) {
				t.Errorf("call order = %v, want %v", calls, want)
			}
		})
	}
}
