package main

import (
	"strings"
	"time"

	dbadapter "go-fiber/adapter/db/sqlx"
	adapterslog "go-fiber/adapter/log/slog"
	apienv "go-fiber/cmd/api/env"

	"go-fiber/pkg/db"
	"go-fiber/pkg/logx"
)

// newLogger creates the application base logger from env config,
// registers it as logx default and returns the instance.
func newLogger(env *apienv.Envs) logx.Logger {
	baseLogger, err := adapterslog.New(adapterslog.Config{
		Service:     env.LOG.Service,
		Environment: env.MODE,
		Level:       env.LOG.Level,
		Format:      env.LOG.Format,
		AddSource:   env.LOG.AddSource,
	})
	if err != nil {
		logx.Default().Fatal("bootstrap logger failed", err)
	}
	logx.SetDefault(baseLogger)
	return baseLogger
}

// newDatabase creates the DB adapter from env configuration.
// Startup is permissive: on config/connection failure, returns nil.
func newDatabase(env *apienv.Envs, base logx.Logger) db.DB {
	base = safeBootstrapLogger(base)
	if env == nil || env.DB == nil {
		base.Warn("db disabled: missing DB env config")
		return nil
	}

	dsn := strings.TrimSpace(env.DB.URL)
	if dsn == "" {
		base.Info("db disabled: DB_URL is empty")
		return nil
	}

	conn, err := dbadapter.New(dsn, dbadapter.Config{
		MaxOpen:     env.DB.MaxOpenConns,
		MaxIdle:     env.DB.MaxIdleConns,
		MaxLifetime: time.Duration(env.DB.ConnMaxLifetimeSeconds) * time.Second,
		PingTimeout: time.Duration(env.DB.PingTimeoutSeconds) * time.Second,
	}, base)
	if err != nil {
		base.Warn("db initialization failed; continuing without db", logx.Error(err))
		return nil
	}
	return conn
}

func safeBootstrapLogger(base logx.Logger) logx.Logger {
	if base == nil {
		return logx.Default()
	}
	return base
}
