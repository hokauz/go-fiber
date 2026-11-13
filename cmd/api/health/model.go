package health

import "time"

const (
	defaultReadyCacheTTL           = 300 * time.Millisecond
	defaultCBFailureThreshold      = 5
	defaultCBTimeout               = 5 * time.Second
	defaultCBSuccessThreshold      = 1
	defaultCBHalfOpenMaxConcurrent = 1
	readySingleflightKey           = "readyz"

	ReasonNotInitialized = "not_initialized"
	ReasonUnhealthy      = "unhealthy"
	ReasonCircuitOpen    = "circuit_open"
	ReasonCircuitHalf    = "circuit_half_open"
	ReasonUnknown        = "unknown"
)

// ReadyConfig controls readiness behavior and HTTP circuit breaker settings.
type ReadyConfig struct {
	IsProduction            bool
	CacheTTL                time.Duration
	CBFailureThreshold      int
	CBTimeout               time.Duration
	CBSuccessThreshold      int
	CBHalfOpenMaxConcurrent int
}

type readinessCache struct {
	status    readinessStatus
	expiresAt time.Time
	valid     bool
}

type readinessStatus struct {
	ready  bool
	reason string
}

func normalizeReadyConfig(cfg ReadyConfig) ReadyConfig {
	if cfg.CacheTTL <= 0 {
		cfg.CacheTTL = defaultReadyCacheTTL
	}
	if cfg.CBFailureThreshold <= 0 {
		cfg.CBFailureThreshold = defaultCBFailureThreshold
	}
	if cfg.CBTimeout <= 0 {
		cfg.CBTimeout = defaultCBTimeout
	}
	if cfg.CBSuccessThreshold <= 0 {
		cfg.CBSuccessThreshold = defaultCBSuccessThreshold
	}
	if cfg.CBHalfOpenMaxConcurrent <= 0 {
		cfg.CBHalfOpenMaxConcurrent = defaultCBHalfOpenMaxConcurrent
	}
	return cfg
}

