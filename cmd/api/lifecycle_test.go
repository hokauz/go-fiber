package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"

	apienv "go-fiber/cmd/api/env"
	"go-fiber/pkg/metrics"
)

type lifecycleDBStub struct {
	mu         sync.Mutex
	closeCalls int
}

func (d *lifecycleDBStub) Ping(context.Context) error { return nil }

func (d *lifecycleDBStub) Readiness(context.Context) bool { return true }

func (d *lifecycleDBStub) WithTx(context.Context, func(*sql.Tx) error) error { return nil }

func (d *lifecycleDBStub) Raw() *sql.DB { return nil }

func (d *lifecycleDBStub) Close() error {
	d.mu.Lock()
	d.closeCalls++
	d.mu.Unlock()
	return nil
}

func (d *lifecycleDBStub) CloseCalls() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.closeCalls
}

func TestRunLifecycle_GracefulShutdownAPIAndMetrics(t *testing.T) {
	apiPort := reserveLocalPort(t)
	metricsPort := reserveLocalPort(t)
	apiURL := fmt.Sprintf("http://127.0.0.1:%d/ping", apiPort)
	metricsURL := fmt.Sprintf("http://127.0.0.1:%d/metrics", metricsPort)

	app := fiber.New()
	app.Get("/ping", func(c fiber.Ctx) error {
		return c.SendStatus(http.StatusNoContent)
	})

	collector := metrics.NewCollector("lifecycle_test", nil)
	metricsSrv := metrics.NewServer(fmt.Sprintf("127.0.0.1:%d", metricsPort), "/metrics", collector.Registry())

	dbStub := &lifecycleDBStub{}
	errCh := make(chan error, 1)
	go func() {
		errCh <- runLifecycle(
			app,
			dbStub,
			nil,
			&apienv.ServerEnvs{
				Port:            fmt.Sprintf("127.0.0.1:%d", apiPort),
				ShutdownTimeout: 5,
			},
			metricsSrv,
		)
	}()

	waitForHTTPStatus(t, apiURL, http.StatusNoContent, 5*time.Second)
	waitForHTTPStatus(t, metricsURL, http.StatusOK, 5*time.Second)

	if err := syscall.Kill(os.Getpid(), syscall.SIGTERM); err != nil {
		t.Fatalf("send SIGTERM: %v", err)
	}

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("runLifecycle returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting runLifecycle to return")
	}

	if got := dbStub.CloseCalls(); got != 1 {
		t.Fatalf("db close calls = %d, want 1", got)
	}

	waitForHTTPUnavailable(t, apiURL, 2*time.Second)
	waitForHTTPUnavailable(t, metricsURL, 2*time.Second)
}

func reserveLocalPort(t *testing.T) int {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve local port: %v", err)
	}
	defer ln.Close()

	addr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatalf("unexpected addr type: %T", ln.Addr())
	}
	return addr.Port
}

func waitForHTTPStatus(t *testing.T, url string, wantStatus int, timeout time.Duration) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	var lastErr error

	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			if resp.StatusCode == wantStatus {
				return
			}
			lastErr = fmt.Errorf("status=%d want=%d", resp.StatusCode, wantStatus)
		} else {
			lastErr = err
		}
		time.Sleep(30 * time.Millisecond)
	}

	t.Fatalf("endpoint %s did not reach status %d in %s (last err: %v)", url, wantStatus, timeout, lastErr)
}

func waitForHTTPUnavailable(t *testing.T, url string, timeout time.Duration) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err != nil {
			return
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		time.Sleep(30 * time.Millisecond)
	}

	t.Fatalf("endpoint %s still reachable after %s", url, timeout)
}
