package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	sqlxlib "github.com/jmoiron/sqlx"

	"go-fiber/pkg/db"
	"go-fiber/pkg/logx"
)

var passwordKVPattern = regexp.MustCompile(`(?i)(password=)([^\s]+)`)

var errDBNotInitialized = errors.New("db not initialized")

type adapter struct {
	raw         *sql.DB
	pingTimeout time.Duration
	logger      logx.Logger
}

// New creates a SQLX adapter backed by pgx stdlib.
func New(dsn string, cfg Config, logger logx.Logger) (db.DB, error) {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return nil, errors.New("db dsn is required")
	}

	cfg = cfg.withDefaults()
	logger = safeLogger(logger)
	logger.Info("db connecting",
		logx.String("driver", "pgx"),
		logx.String("dsn", sanitizeDSN(dsn)),
	)

	sx, err := sqlxlib.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	raw := sx.DB
	raw.SetMaxOpenConns(cfg.MaxOpen)
	raw.SetMaxIdleConns(cfg.MaxIdle)
	raw.SetConnMaxLifetime(cfg.MaxLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.PingTimeout)
	defer cancel()

	if err := raw.PingContext(ctx); err != nil {
		_ = raw.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	logger.Info("db connected")
	return &adapter{
		raw:         raw,
		pingTimeout: cfg.PingTimeout,
		logger:      logger,
	}, nil
}

func (a *adapter) Ping(ctx context.Context) error {
	if a == nil || a.raw == nil {
		return errDBNotInitialized
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return a.raw.PingContext(ctx)
}

func (a *adapter) Readiness(ctx context.Context) bool {
	if a == nil || a.raw == nil {
		return false
	}
	if ctx == nil {
		ctx = context.Background()
	}

	timeout := a.pingTimeout
	if timeout <= 0 {
		timeout = DefaultConfig().PingTimeout
	}

	checkCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := a.raw.PingContext(checkCtx); err != nil {
		safeLogger(a.logger).Warn("db readiness check failed", logx.Error(err))
		return false
	}

	return true
}

func (a *adapter) WithTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	if a == nil || a.raw == nil {
		return errDBNotInitialized
	}
	if fn == nil {
		return errors.New("tx function is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	tx, err := a.raw.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return err
	}

	return nil
}

func (a *adapter) Raw() *sql.DB {
	if a == nil {
		return nil
	}
	return a.raw
}

func (a *adapter) Close() error {
	if a == nil || a.raw == nil {
		return nil
	}
	return a.raw.Close()
}

func safeLogger(logger logx.Logger) logx.Logger {
	if logger == nil {
		return logx.Default()
	}
	return logger
}

func sanitizeDSN(dsn string) string {
	u, err := url.Parse(dsn)
	if err == nil && u != nil {
		if u.User != nil {
			if _, hasPassword := u.User.Password(); hasPassword {
				u.User = url.UserPassword(u.User.Username(), "***")
			}
		}
		return u.String()
	}

	return passwordKVPattern.ReplaceAllString(dsn, "$1***")
}
