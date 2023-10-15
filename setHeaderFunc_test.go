package transport_test

import (
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/golang-cz/transport"
	"golang.org/x/sync/errgroup"
)

func TestSetHeaderFunc(t *testing.T) {
	var uniqueAuth sync.Map
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")

		// Error out if we see the same Authorization header twice.
		if _, ok := uniqueAuth.LoadOrStore(auth, true); ok {
			w.WriteHeader(401)
			fmt.Fprintf(w, "%v", fmt.Errorf("received the same Authorization header twice: %q", auth))
			return
		}

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

	// Send request concurrently.
	// Each request should send random Authorization header value.
	//
	// NOTE: On macOS, 128 might be the max number of parallel connections
	//       allowed by OS. To increase the limits, you might need to run:
	//       sudo ulimit -n 6049
	//       sudo sysctl -w kern.ipc.somaxconn=1024
	var g errgroup.Group
	for i := 0; i < 128; i++ {
		g.Go(func() error {
			resp, err := authClient.Get(srv.URL)
			if err != nil {
				return fmt.Errorf("sending auth'd request: %w", err)
			}

			if resp.StatusCode != 200 {
				b, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("HTTP %v:\n%s", resp.StatusCode, string(b))
			}

			return nil
		})
	}
	if err := g.Wait(); err != nil {
		t.Fatal(err)
	}
}

func issueRandomAuthToken(req *http.Request) string {
	b := make([]byte, 32)
	rand.Read(b)
	return fmt.Sprintf("BEARER %x", b)
}
