package health

import (
	"time"

	"github.com/gofiber/fiber/v3"

	apienv "go-fiber/cmd/api/env"

	"go-fiber/pkg/db"
)

// PublicRegister builds and returns route registration and cleanup functions.
func PublicRegister(database db.DB, envs *apienv.Envs) (func(fiber.Router), func()) {
	public := NewPublicFromEnvs(database, envs)
	return public.Register, public.Stop
}

// NewPublicFromEnvs builds Public health runtime from API envs.
func NewPublicFromEnvs(database db.DB, envs *apienv.Envs) *Public {
	return NewPublic(database, readyConfigFromEnvs(envs))
}

func readyConfigFromEnvs(envs *apienv.Envs) ReadyConfig {
	if envs == nil {
		return ReadyConfig{}
	}

	cfg := ReadyConfig{
		IsProduction: envs.IsProduction,
	}

	if envs.HEALTH == nil {
		return cfg
	}

	cfg.CacheTTL = time.Duration(envs.HEALTH.ReadyCacheTTLMS) * time.Millisecond
	cfg.CBFailureThreshold = envs.HEALTH.ReadyCBFailureThreshold
	cfg.CBTimeout = time.Duration(envs.HEALTH.ReadyCBTimeoutSeconds) * time.Second
	cfg.CBSuccessThreshold = envs.HEALTH.ReadyCBSuccessThreshold
	cfg.CBHalfOpenMaxConcurrent = envs.HEALTH.ReadyCBHalfOpenConcurrent

	return cfg
}
