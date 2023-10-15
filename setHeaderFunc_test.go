package transport_test

import (
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-cz/transport"
)

func TestSetHeaderFunc(t *testing.T) {
	uniqueAuth := map[string]bool{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")

		// Error out if we see the same Authorization header twice.
		if _, ok := uniqueAuth[auth]; ok {
			w.WriteHeader(401)
			fmt.Fprintf(w, "%v", fmt.Errorf("received same Authorization header twice: %q", r.Header.Get("Authorization")))
			return
		}

		uniqueAuth[auth] = true

		w.WriteHeader(200)
	}))
	defer srv.Close()

	authClient := http.Client{
		Transport: transport.Chain(
			http.DefaultTransport,
			transport.SetHeaderFunc("Authorization", issueRandomAuthToken),
			transport.DebugRequests,
		),
		Timeout: 15 * time.Second,
	}

	// Send several requests. Each should have random Authorization header value.
	for i := 0; i < 5; i++ {
		resp, err := authClient.Get(srv.URL)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != 200 {
			b, _ := io.ReadAll(resp.Body)
			t.Fatal(string(b))
		}
	}
}

func issueRandomAuthToken(req *http.Request) string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("BEARER %x", b)
}
