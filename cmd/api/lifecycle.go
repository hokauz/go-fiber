package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"

	"go-fiber/cmd/api/env"
	"go-fiber/pkg/db"
	"go-fiber/pkg/logx"
	"go-fiber/pkg/netutil"
)

const (
	defaultListenPort      = "8080"
	defaultShutdownTimeout = 15 * time.Second
)

type lifecycleServerError struct {
	name string
	err  error
}

// runLifecycle starts Fiber, waits for OS shutdown signals, then gracefully
// shuts down the API server and DB resources.
func runLifecycle(app *fiber.App, database db.DB, appLog logx.Logger, server *env.ServerEnvs, metricsSrv *http.Server) error {
	if app == nil {
		return errors.New("fiber app is required")
	}

	appLog = safeLifecycleLogger(appLog)
	listenAddr := apiListenAddr(server)
	shutdownTimeout := defaultShutdownTimeout
	if server != nil && server.ShutdownTimeout > 0 {
		shutdownTimeout = time.Duration(server.ShutdownTimeout) * time.Second
	}

	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverCount := 1
	if metricsSrv != nil {
		serverCount++
	}
	errCh := make(chan lifecycleServerError, serverCount)

	go func() {
		errCh <- lifecycleServerError{name: "api", err: app.Listen(listenAddr)}
	}()
	if metricsSrv != nil {
		go func() {
			errCh <- lifecycleServerError{name: "metrics", err: metricsSrv.ListenAndServe()}
		}()
	}

	select {
	case srvErr := <-errCh:
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := shutdownServers(shutdownCtx, app, metricsSrv); err != nil {
			appLog.Warn("server shutdown after early failure had errors", logx.Error(err))
		}
		_ = closeDatabaseWithContext(shutdownCtx, database, appLog)
		if isExpectedServerExit(srvErr.err) {
			return nil
		}
		return fmt.Errorf("%s server failed before signal: %w", srvErr.name, srvErr.err)
	case <-sigCtx.Done():
		appLog.Info("shutdown signal received", logx.Duration("timeout", shutdownTimeout))
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := shutdownServers(shutdownCtx, app, metricsSrv); err != nil {
		appLog.Error("failed to shutdown servers", logx.Error(err))
		_ = closeDatabaseWithContext(shutdownCtx, database, appLog)
		return err
	}
	appLog.Info("servers shutdown completed")

	if err := closeDatabaseWithContext(shutdownCtx, database, appLog); err != nil && !errors.Is(err, context.DeadlineExceeded) {
		appLog.Error("failed to close db", logx.Error(err))
	}

	for i := 0; i < serverCount; i++ {
		select {
		case srvErr := <-errCh:
			if !isExpectedServerExit(srvErr.err) {
				appLog.Debug("server exited with error", logx.String("server", srvErr.name), logx.Error(srvErr.err))
			}
		case <-shutdownCtx.Done():
			appLog.Warn("timeout waiting listeners to stop", logx.Error(shutdownCtx.Err()))
			return nil
		}
	}

	return nil
}

// startServerOrFatal runs the lifecycle and calls appLog.Fatal on error.
func startServerOrFatal(app *fiber.App, database db.DB, appLog logx.Logger, server *env.ServerEnvs, metricsSrv *http.Server, cleanups ...func()) {
	if err := runLifecycle(app, database, appLog, server, metricsSrv); err != nil {
		runLifecycleCleanups(cleanups...)
		appLog.Fatal("api lifecycle failed", err)
	}
	runLifecycleCleanups(cleanups...)
}

func closeDatabaseWithContext(ctx context.Context, database db.DB, appLog logx.Logger) error {
	if database == nil {
		return nil
	}

	done := make(chan error, 1)
	go func() {
		done <- database.Close()
	}()

	select {
	case err := <-done:
		if err != nil {
			return err
		}
		appLog.Info("db closed")
		return nil
	case <-ctx.Done():
		appLog.Warn("database shutdown timed out", logx.Error(ctx.Err()))
		go func() { _ = database.Close() }()
		return ctx.Err()
	}
}

func safeLifecycleLogger(appLog logx.Logger) logx.Logger {
	if appLog == nil {
		return logx.Default()
	}
	return appLog
}

func apiListenAddr(server *env.ServerEnvs) string {
	if server == nil {
		return ":" + defaultListenPort
	}
	return netutil.NormalizeListenAddr(server.Port, defaultListenPort)
}

func shutdownServers(ctx context.Context, app *fiber.App, metricsSrv *http.Server) error {
	if err := app.ShutdownWithContext(ctx); err != nil {
		return err
	}
	if metricsSrv != nil {
		if err := metricsSrv.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}
	return nil
}

func isExpectedServerExit(err error) bool {
	if err == nil || errors.Is(err, http.ErrServerClosed) || errors.Is(err, net.ErrClosed) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "server closed") || strings.Contains(msg, "use of closed network connection")
}

func runLifecycleCleanups(cleanups ...func()) {
	for _, cleanup := range cleanups {
		if cleanup == nil {
			continue
		}
		cleanup()
	}
}
