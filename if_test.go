package transport_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/transport"
)

func TestIfTrue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("debug") != "true" {
			t.Error("expected debug=true")
			w.WriteHeader(500)
			return
		}

		fmt.Fprintf(w, "ok")
	}))
	defer server.Close()

	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: transport.Chain(
			http.DefaultTransport,
			transport.If(true, transport.SetHeader("debug", "true")), // Set header.
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
		t.Fatal("unexpected response")
	}
}

func TestIfFalse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("debug") != "" {
			t.Error("expected no debug header")
			w.WriteHeader(500)
			return
		}

		fmt.Fprintf(w, "ok")
	}))
	defer server.Close()

	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: transport.Chain(
			http.DefaultTransport,
			transport.If(false, transport.SetHeader("debug", "true")), // Do not set header.
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
		t.Fatal("unexpected response")
	}
}
