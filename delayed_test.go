package transport_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"net/http"
	"net/http/httptest"

	"github.com/go-chi/transport"
)

func TestDelayed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "ok")
	}))
	defer server.Close()

	t.Run("default config", func(t *testing.T) {
		client := &http.Client{
			Transport: transport.Chain(
				nil,
				transport.Delayed(
					transport.DelayedConfig{},
				),
			),
		}

		request, err := http.NewRequest("GET", server.URL, nil)
		if err != nil {
			t.Fatal(err)
		}

		timeStart := time.Now()
		resp, err := client.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		timeElapsed := time.Since(timeStart)

		t.Logf("elapsed time: %v", timeElapsed)

		if resp.StatusCode != 200 {
			t.Fatal("expected some header, but did not receive")
		}

		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("response: %s", string(buf))
	})

	t.Run("delayed response", func(t *testing.T) {
		client := &http.Client{
			Transport: transport.Chain(
				nil,
				transport.Delayed(
					transport.DelayedConfig{
						ResponseDelayMin: 100 * time.Millisecond,
						ResponseDelayMax: 200 * time.Millisecond,
					},
				),
			),
		}

		request, err := http.NewRequest("GET", server.URL, nil)
		if err != nil {
			t.Fatal(err)
		}

		timeStart := time.Now()
		_, err = client.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		timeElapsed := time.Since(timeStart)

		if timeElapsed < 100*time.Millisecond {
			t.Fatalf("expected at least 100ms delay, but got %v", timeElapsed)
		}
	})

	t.Run("delayed connect", func(t *testing.T) {
		client := &http.Client{
			Transport: transport.Chain(
				nil,
				transport.Delayed(
					transport.DelayedConfig{
						RequestDelayMin: 100 * time.Millisecond,
						RequestDelayMax: 200 * time.Millisecond,
					},
				),
			),
		}

		request, err := http.NewRequest("GET", server.URL, nil)
		if err != nil {
			t.Fatal(err)
		}

		timeStart := time.Now()
		_, err = client.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		timeElapsed := time.Since(timeStart)

		if timeElapsed < 100*time.Millisecond {
			t.Fatalf("expected at least 100ms delay, but got %v", timeElapsed)
		}
	})

	t.Run("delayed connect and response", func(t *testing.T) {
		client := &http.Client{
			Transport: transport.Chain(
				nil,
				transport.Delayed(
					transport.DelayedConfig{
						RequestDelayMin:  50 * time.Millisecond,
						RequestDelayMax:  100 * time.Millisecond,
						ResponseDelayMin: 50 * time.Millisecond,
						ResponseDelayMax: 100 * time.Millisecond,
					},
				),
			),
		}

		request, err := http.NewRequest("GET", server.URL, nil)
		if err != nil {
			t.Fatal(err)
		}

		timeStart := time.Now()
		_, err = client.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		timeElapsed := time.Since(timeStart)

		if timeElapsed < 100*time.Millisecond {
			t.Fatalf("expected at least 100ms delay, but got %v", timeElapsed)
		}
	})

	t.Run("chained transport", func(t *testing.T) {
		var customTransportHit bool

		customTransport := transport.RoundTripFunc(func(req *http.Request) (*http.Response, error) {
			customTransportHit = true

			return http.DefaultTransport.RoundTrip(req)
		})

		client := &http.Client{
			Transport: transport.Chain(
				customTransport,

				transport.Delayed(
					transport.DelayedConfig{
						RequestDelayMin: 100 * time.Millisecond,
						RequestDelayMax: 200 * time.Millisecond,
					},
				),
			),
		}

		request, err := http.NewRequest("GET", server.URL, nil)
		if err != nil {
			t.Fatal(err)
		}

		timeStart := time.Now()
		_, err = client.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		timeElapsed := time.Since(timeStart)

		if timeElapsed < 100*time.Millisecond {
			t.Fatalf("expected at least 100ms delay, but got %v", timeElapsed)
		}

		if customTransportHit == false {
			t.Fatal("expected custom transport to be hit, but it was not")
		}
	})

	t.Run("honor request context", func(t *testing.T) {
		client := &http.Client{
			Transport: transport.Chain(
				nil,
				transport.Delayed(
					transport.DelayedConfig{
						RequestDelayMin: 100 * time.Millisecond,
						RequestDelayMax: 200 * time.Millisecond,
					},
				),
			),
		}

		request, err := http.NewRequest("GET", server.URL, nil)
		if err != nil {
			t.Fatal(err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		request = request.WithContext(ctx)

		timeStart := time.Now()
		_, err = client.Do(request)
		timeElapsed := time.Since(timeStart)

		if err == nil {
			t.Fatalf("expected error, but got none")
		}

		if timeElapsed < 50*time.Millisecond {
			t.Fatalf("expected at least 50ms delay, but got %v", timeElapsed)
		}

		if timeElapsed > 100*time.Millisecond {
			t.Fatalf("expected less than 100ms delay, but got %v", timeElapsed)
		}
	})
}
