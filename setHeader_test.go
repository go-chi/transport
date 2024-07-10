package transport_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/transport"
)

func TestSetHeader(t *testing.T) {
	userAgent := "my-app/v1.0.0"
	authHeader := fmt.Sprintf("BEARER %v", "AUTH")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != userAgent {
			w.WriteHeader(400)
			fmt.Fprintf(w, "%v", fmt.Errorf("unexpected User-Agent header: %q", r.Header.Get("User-Agent")))
			return
		}

		if r.Header.Get("Authorization") != authHeader {
			w.WriteHeader(401)
			fmt.Fprintf(w, "%v", fmt.Errorf("unexpected Authorization header: %q", r.Header.Get("Authorization")))
			return
		}

		if r.Header.Get("X-EXTRA") != "value" {
			w.WriteHeader(400)
			fmt.Fprintf(w, "%v", fmt.Errorf("unexpected X-EXTRA header: %q", r.Header.Get("X-EXTRA")))
			return
		}

		w.WriteHeader(200)
		fmt.Fprintf(w, "OK!")
	}))
	defer srv.Close()

	authClient := http.Client{
		Transport: transport.Chain(
			http.DefaultTransport,
			transport.SetHeader("User-Agent", userAgent),
			transport.SetHeader("Authorization", authHeader),
			transport.SetHeader("x-extra", "value"),
			transport.LogRequests(&transport.DefaultLogger{PrintResponsePayload: true}),
		),
		Timeout: 15 * time.Second,
	}

	resp, err := authClient.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		t.Fatal(string(b))
	}
}
