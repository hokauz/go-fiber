package db

import (
	"context"
	"database/sql"
)

// DB is the database port used by application services and handlers.
type DB interface {
	Ping(ctx context.Context) error
	Readiness(ctx context.Context) bool
	WithTx(ctx context.Context, fn func(tx *sql.Tx) error) error
	Raw() *sql.DB
	Close() error
}
