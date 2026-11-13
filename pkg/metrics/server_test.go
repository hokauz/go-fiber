package metrics

import (
	"context"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestNewServer_ServesMetricsEndpoint(t *testing.T) {
	col := NewCollector("srv_api", nil)
	col.ObserveRequest("GET", "/api/livez", 200, time.Millisecond)
	srv := NewServer("127.0.0.1:0", "metrics", col.Registry())

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Serve(ln)
	}()

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
		select {
		case serveErr := <-errCh:
			if serveErr != nil && serveErr != http.ErrServerClosed {
				t.Fatalf("serve error: %v", serveErr)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("timeout waiting metrics server shutdown")
		}
	})

	resp, err := http.Get("http://" + ln.Addr().String() + "/metrics")
	if err != nil {
		t.Fatalf("get /metrics: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}

	if !strings.Contains(string(body), "srv_api_http_requests_total") {
		t.Fatal("expected custom prefixed metrics in /metrics output")
	}
}
