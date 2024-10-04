package transport_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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
