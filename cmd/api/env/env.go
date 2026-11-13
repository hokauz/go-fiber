package env

import (
	"log"
	"os"
	"strings"

	configenv "go-fiber/pkg/envloader"
)

// AuthEnvs holds structured env configuration for auth.
type AuthEnvs struct {
	Issuer             string
	Audience           string
	AllowedAlgs        string
	JWKSURL            string
	JWKSTTLSeconds     int
	HTTPTimeoutSeconds int
}

// LogEnvs holds structured env configuration for logging.
type LogEnvs struct {
	Service   string
	Level     string
	Format    string
	AddSource bool
}

// DBEnvs holds structured env configuration for DB connectivity.
type DBEnvs struct {
	URL                    string
	MaxOpenConns           int
	MaxIdleConns           int
	ConnMaxLifetimeSeconds int
	PingTimeoutSeconds     int
}

// ServerEnvs holds structured env configuration for server lifecycle.
type ServerEnvs struct {
	Port                string
	BasePath            string
	ShutdownTimeout     int
	ReadTimeoutSeconds  int
	WriteTimeoutSeconds int
	IdleTimeoutSeconds  int
}

// CORSEnvs holds structured env configuration for CORS middleware.
type CORSEnvs struct {
	Origins string
}

// MetricsEnvs holds structured env configuration for Prometheus metrics.
type MetricsEnvs struct {
	Enabled bool
	Port    string
	Path    string
	Prefix  string
}

// HealthEnvs holds structured env configuration for readiness behavior.
type HealthEnvs struct {
	ReadyCacheTTLMS           int
	ReadyCBFailureThreshold   int
	ReadyCBTimeoutSeconds     int
	ReadyCBSuccessThreshold   int
	ReadyCBHalfOpenConcurrent int
}

// Envs holds structured env configuration for the API.
type Envs struct {
	MODE         string
	IsProduction bool
	APP          string
	LOG          *LogEnvs
	AUTH         *AuthEnvs
	DB           *DBEnvs
	SERVER       *ServerEnvs
	CORS         *CORSEnvs
	METRICS      *MetricsEnvs
	HEALTH       *HealthEnvs
}

var appEnvDefs = []configenv.EnvDef{
	{Name: "APP_ENV", Type: configenv.TypeString, Required: false, Default: "development"},
	{Name: "APP_NAME", Type: configenv.TypeString, Required: false, Default: "go-fiber-api"},
}

var logEnvDefs = []configenv.EnvDef{
	{Name: "APP_LOG_LEVEL", Type: configenv.TypeString, Required: false, Default: "info"},
	{Name: "APP_LOG_FORMAT", Type: configenv.TypeString, Required: false, Default: "json"},
	{Name: "APP_LOG_ADD_SOURCE", Type: configenv.TypeBool, Required: false, Default: false},
}

var authEnvDefs = []configenv.EnvDef{
	{Name: "AUTH_ISSUER", Type: configenv.TypeString, Required: false, Default: ""},
	{Name: "AUTH_AUDIENCE", Type: configenv.TypeString, Required: false, Default: ""},
	{Name: "AUTH_ALLOWED_ALGS", Type: configenv.TypeString, Required: false, Default: "RS256"},
	{Name: "AUTH_JWKS_URL", Type: configenv.TypeString, Required: false, Default: ""},
	{Name: "AUTH_JWKS_TTL_SECONDS", Type: configenv.TypeInt, Required: false, Default: 600},
	{Name: "AUTH_HTTP_TIMEOUT_SECONDS", Type: configenv.TypeInt, Required: false, Default: 3},
}

var dbEnvDefs = []configenv.EnvDef{
	{Name: "DB_URL", Type: configenv.TypeString, Required: false, Default: ""},
	{Name: "DB_MAX_OPEN_CONNS", Type: configenv.TypeInt, Required: false, Default: 10},
	{Name: "DB_MAX_IDLE_CONNS", Type: configenv.TypeInt, Required: false, Default: 5},
	{Name: "DB_CONN_MAX_LIFETIME_SECONDS", Type: configenv.TypeInt, Required: false, Default: 300},
	{Name: "DB_PING_TIMEOUT_SECONDS", Type: configenv.TypeInt, Required: false, Default: 5},
}

var serverEnvDefs = []configenv.EnvDef{
	{Name: "PORT", Type: configenv.TypeString, Required: false, Default: "8080"},
	{Name: "SERVER_BASE_PATH", Type: configenv.TypeString, Required: false, Default: "/api"},
	{Name: "SERVER_SHUTDOWN_TIMEOUT", Type: configenv.TypeInt, Required: false, Default: 15},
	{Name: "SERVER_READ_TIMEOUT_SECONDS", Type: configenv.TypeInt, Required: false, Default: 5},
	{Name: "SERVER_WRITE_TIMEOUT_SECONDS", Type: configenv.TypeInt, Required: false, Default: 15},
	{Name: "SERVER_IDLE_TIMEOUT_SECONDS", Type: configenv.TypeInt, Required: false, Default: 60},
}

var corsEnvDefs = []configenv.EnvDef{
	{Name: "CORS_ORIGINS", Type: configenv.TypeString, Required: false, Default: "*"},
}

var metricsEnvDefs = []configenv.EnvDef{
	{Name: "METRICS_ENABLED", Type: configenv.TypeBool, Required: false, Default: true},
	{Name: "METRICS_PORT", Type: configenv.TypeString, Required: false, Default: "9090"},
	{Name: "METRICS_PATH", Type: configenv.TypeString, Required: false, Default: "/metrics"},
	{Name: "APP_METRICS_PREFIX", Type: configenv.TypeString, Required: false, Default: "go_fiber_api"},
}

var healthEnvDefs = []configenv.EnvDef{
	{Name: "HEALTH_READY_CACHE_TTL_MS", Type: configenv.TypeInt, Required: false, Default: 300},
	{Name: "HEALTH_READY_CB_FAILURE_THRESHOLD", Type: configenv.TypeInt, Required: false, Default: 5},
	{Name: "HEALTH_READY_CB_TIMEOUT_SECONDS", Type: configenv.TypeInt, Required: false, Default: 5},
	{Name: "HEALTH_READY_CB_SUCCESS_THRESHOLD", Type: configenv.TypeInt, Required: false, Default: 1},
	{Name: "HEALTH_READY_CB_HALF_OPEN_MAX_CONCURRENT", Type: configenv.TypeInt, Required: false, Default: 1},
}

// Load loads envs from environment. On error, logs and exits.
func Load() *Envs {
	allDefs := append([]configenv.EnvDef{}, appEnvDefs...)
	allDefs = append(allDefs, logEnvDefs...)
	allDefs = append(allDefs, authEnvDefs...)
	allDefs = append(allDefs, dbEnvDefs...)
	allDefs = append(allDefs, serverEnvDefs...)
	allDefs = append(allDefs, corsEnvDefs...)
	allDefs = append(allDefs, metricsEnvDefs...)
	allDefs = append(allDefs, healthEnvDefs...)
	cfg, msgs := configenv.LoadWith(allDefs)
	if cfg == nil {
		for _, m := range msgs {
			if m.Level == configenv.LevelError {
				log.Printf("env: %s", m.String())
			}
		}
		os.Exit(1)
	}
	for _, m := range msgs {
		if m.Level == configenv.LevelWarning {
			log.Printf("env [warning]: %s", m.String())
		}
	}
	mode := cfg.GetString("APP_ENV")
	return &Envs{
		MODE:         mode,
		IsProduction: strings.EqualFold(strings.TrimSpace(mode), "production"),
		APP:          cfg.GetString("APP_NAME"),
		LOG: &LogEnvs{
			Service:   cfg.GetString("APP_NAME"),
			Level:     cfg.GetString("APP_LOG_LEVEL"),
			Format:    cfg.GetString("APP_LOG_FORMAT"),
			AddSource: cfg.GetBool("APP_LOG_ADD_SOURCE"),
		},
		AUTH: &AuthEnvs{
			Issuer:             cfg.GetString("AUTH_ISSUER"),
			Audience:           cfg.GetString("AUTH_AUDIENCE"),
			AllowedAlgs:        cfg.GetString("AUTH_ALLOWED_ALGS"),
			JWKSURL:            cfg.GetString("AUTH_JWKS_URL"),
			JWKSTTLSeconds:     cfg.GetInt("AUTH_JWKS_TTL_SECONDS"),
			HTTPTimeoutSeconds: cfg.GetInt("AUTH_HTTP_TIMEOUT_SECONDS"),
		},
		DB: &DBEnvs{
			URL:                    cfg.GetString("DB_URL"),
			MaxOpenConns:           cfg.GetInt("DB_MAX_OPEN_CONNS"),
			MaxIdleConns:           cfg.GetInt("DB_MAX_IDLE_CONNS"),
			ConnMaxLifetimeSeconds: cfg.GetInt("DB_CONN_MAX_LIFETIME_SECONDS"),
			PingTimeoutSeconds:     cfg.GetInt("DB_PING_TIMEOUT_SECONDS"),
		},
		SERVER: &ServerEnvs{
			Port:                cfg.GetString("PORT"),
			BasePath:            cfg.GetString("SERVER_BASE_PATH"),
			ShutdownTimeout:     cfg.GetInt("SERVER_SHUTDOWN_TIMEOUT"),
			ReadTimeoutSeconds:  cfg.GetInt("SERVER_READ_TIMEOUT_SECONDS"),
			WriteTimeoutSeconds: cfg.GetInt("SERVER_WRITE_TIMEOUT_SECONDS"),
			IdleTimeoutSeconds:  cfg.GetInt("SERVER_IDLE_TIMEOUT_SECONDS"),
		},
		CORS: &CORSEnvs{
			Origins: cfg.GetString("CORS_ORIGINS"),
		},
		METRICS: &MetricsEnvs{
			Enabled: cfg.GetBool("METRICS_ENABLED"),
			Port:    cfg.GetString("METRICS_PORT"),
			Path:    cfg.GetString("METRICS_PATH"),
			Prefix:  cfg.GetString("APP_METRICS_PREFIX"),
		},
		HEALTH: &HealthEnvs{
			ReadyCacheTTLMS:           cfg.GetInt("HEALTH_READY_CACHE_TTL_MS"),
			ReadyCBFailureThreshold:   cfg.GetInt("HEALTH_READY_CB_FAILURE_THRESHOLD"),
			ReadyCBTimeoutSeconds:     cfg.GetInt("HEALTH_READY_CB_TIMEOUT_SECONDS"),
			ReadyCBSuccessThreshold:   cfg.GetInt("HEALTH_READY_CB_SUCCESS_THRESHOLD"),
			ReadyCBHalfOpenConcurrent: cfg.GetInt("HEALTH_READY_CB_HALF_OPEN_MAX_CONCURRENT"),
		},
	}
}
