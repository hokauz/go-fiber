package slog

import "io"

// Config controls slog adapter setup.
type Config struct {
	Level       string
	Format      string
	Service     string
	Environment string
	AddSource   bool
	Writer      io.Writer
}
