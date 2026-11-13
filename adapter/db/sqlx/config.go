package sqlx

import "time"

// Config controls SQL connection pool and startup ping behavior.
type Config struct {
	MaxOpen     int
	MaxIdle     int
	MaxLifetime time.Duration
	PingTimeout time.Duration
}

// DefaultConfig returns safe defaults for API workloads.
func DefaultConfig() Config {
	return Config{
		MaxOpen:     10,
		MaxIdle:     5,
		MaxLifetime: 5 * time.Minute,
		PingTimeout: 5 * time.Second,
	}
}

func (c Config) withDefaults() Config {
	d := DefaultConfig()
	if c.MaxOpen <= 0 {
		c.MaxOpen = d.MaxOpen
	}
	if c.MaxIdle < 0 {
		c.MaxIdle = d.MaxIdle
	}
	if c.MaxLifetime <= 0 {
		c.MaxLifetime = d.MaxLifetime
	}
	if c.PingTimeout <= 0 {
		c.PingTimeout = d.PingTimeout
	}
	return c
}
