package env

import "testing"

func TestLoad_DBDefaults(t *testing.T) {
	t.Setenv("DB_URL", "")
	t.Setenv("DB_MAX_OPEN_CONNS", "")
	t.Setenv("DB_MAX_IDLE_CONNS", "")
	t.Setenv("DB_CONN_MAX_LIFETIME_SECONDS", "")
	t.Setenv("DB_PING_TIMEOUT_SECONDS", "")
	t.Setenv("PORT", "")
	t.Setenv("SERVER_BASE_PATH", "")
	t.Setenv("SERVER_SHUTDOWN_TIMEOUT", "")
	t.Setenv("SERVER_READ_TIMEOUT_SECONDS", "")
	t.Setenv("SERVER_WRITE_TIMEOUT_SECONDS", "")
	t.Setenv("SERVER_IDLE_TIMEOUT_SECONDS", "")
	t.Setenv("CORS_ORIGINS", "")
	t.Setenv("METRICS_ENABLED", "")
	t.Setenv("METRICS_PORT", "")
	t.Setenv("METRICS_PATH", "")
	t.Setenv("APP_METRICS_PREFIX", "")
	t.Setenv("HEALTH_READY_CACHE_TTL_MS", "")
	t.Setenv("HEALTH_READY_CB_FAILURE_THRESHOLD", "")
	t.Setenv("HEALTH_READY_CB_TIMEOUT_SECONDS", "")
	t.Setenv("HEALTH_READY_CB_SUCCESS_THRESHOLD", "")
	t.Setenv("HEALTH_READY_CB_HALF_OPEN_MAX_CONCURRENT", "")

	envs := Load()
	if envs.DB == nil {
		t.Fatal("DB envs should not be nil")
	}
	if envs.DB.URL != "" {
		t.Fatalf("DB_URL = %q, want empty", envs.DB.URL)
	}
	if envs.DB.MaxOpenConns != 10 {
		t.Fatalf("DB_MAX_OPEN_CONNS = %d, want 10", envs.DB.MaxOpenConns)
	}
	if envs.DB.MaxIdleConns != 5 {
		t.Fatalf("DB_MAX_IDLE_CONNS = %d, want 5", envs.DB.MaxIdleConns)
	}
	if envs.DB.ConnMaxLifetimeSeconds != 300 {
		t.Fatalf("DB_CONN_MAX_LIFETIME_SECONDS = %d, want 300", envs.DB.ConnMaxLifetimeSeconds)
	}
	if envs.DB.PingTimeoutSeconds != 5 {
		t.Fatalf("DB_PING_TIMEOUT_SECONDS = %d, want 5", envs.DB.PingTimeoutSeconds)
	}
	if envs.SERVER == nil {
		t.Fatal("SERVER envs should not be nil")
	}
	if envs.SERVER.Port != "8080" {
		t.Fatalf("PORT = %q, want 8080", envs.SERVER.Port)
	}
	if envs.SERVER.BasePath != "/api" {
		t.Fatalf("SERVER_BASE_PATH = %q, want /api", envs.SERVER.BasePath)
	}
	if envs.SERVER.ShutdownTimeout != 15 {
		t.Fatalf("SERVER_SHUTDOWN_TIMEOUT = %d, want 15", envs.SERVER.ShutdownTimeout)
	}
	if envs.SERVER.ReadTimeoutSeconds != 5 {
		t.Fatalf("SERVER_READ_TIMEOUT_SECONDS = %d, want 5", envs.SERVER.ReadTimeoutSeconds)
	}
	if envs.SERVER.WriteTimeoutSeconds != 15 {
		t.Fatalf("SERVER_WRITE_TIMEOUT_SECONDS = %d, want 15", envs.SERVER.WriteTimeoutSeconds)
	}
	if envs.SERVER.IdleTimeoutSeconds != 60 {
		t.Fatalf("SERVER_IDLE_TIMEOUT_SECONDS = %d, want 60", envs.SERVER.IdleTimeoutSeconds)
	}
	if envs.CORS == nil {
		t.Fatal("CORS envs should not be nil")
	}
	if envs.CORS.Origins != "*" {
		t.Fatalf("CORS_ORIGINS = %q, want *", envs.CORS.Origins)
	}
	if envs.METRICS == nil {
		t.Fatal("METRICS envs should not be nil")
	}
	if !envs.METRICS.Enabled {
		t.Fatal("METRICS_ENABLED = false, want true")
	}
	if envs.METRICS.Port != "9090" {
		t.Fatalf("METRICS_PORT = %q, want 9090", envs.METRICS.Port)
	}
	if envs.METRICS.Path != "/metrics" {
		t.Fatalf("METRICS_PATH = %q, want /metrics", envs.METRICS.Path)
	}
	if envs.METRICS.Prefix != "go_fiber_api" {
		t.Fatalf("APP_METRICS_PREFIX = %q, want go_fiber_api", envs.METRICS.Prefix)
	}
	if envs.HEALTH == nil {
		t.Fatal("HEALTH envs should not be nil")
	}
	if envs.HEALTH.ReadyCacheTTLMS != 300 {
		t.Fatalf("HEALTH_READY_CACHE_TTL_MS = %d, want 300", envs.HEALTH.ReadyCacheTTLMS)
	}
	if envs.HEALTH.ReadyCBFailureThreshold != 5 {
		t.Fatalf("HEALTH_READY_CB_FAILURE_THRESHOLD = %d, want 5", envs.HEALTH.ReadyCBFailureThreshold)
	}
	if envs.HEALTH.ReadyCBTimeoutSeconds != 5 {
		t.Fatalf("HEALTH_READY_CB_TIMEOUT_SECONDS = %d, want 5", envs.HEALTH.ReadyCBTimeoutSeconds)
	}
	if envs.HEALTH.ReadyCBSuccessThreshold != 1 {
		t.Fatalf("HEALTH_READY_CB_SUCCESS_THRESHOLD = %d, want 1", envs.HEALTH.ReadyCBSuccessThreshold)
	}
	if envs.HEALTH.ReadyCBHalfOpenConcurrent != 1 {
		t.Fatalf("HEALTH_READY_CB_HALF_OPEN_MAX_CONCURRENT = %d, want 1", envs.HEALTH.ReadyCBHalfOpenConcurrent)
	}
}

func TestLoad_DBFromEnv(t *testing.T) {
	t.Setenv("DB_URL", "postgresql://user:pass@localhost:5432/postgres")
	t.Setenv("DB_MAX_OPEN_CONNS", "21")
	t.Setenv("DB_MAX_IDLE_CONNS", "7")
	t.Setenv("DB_CONN_MAX_LIFETIME_SECONDS", "901")
	t.Setenv("DB_PING_TIMEOUT_SECONDS", "8")

	envs := Load()
	if envs.DB == nil {
		t.Fatal("DB envs should not be nil")
	}
	if envs.DB.URL != "postgresql://user:pass@localhost:5432/postgres" {
		t.Fatalf("DB_URL = %q", envs.DB.URL)
	}
	if envs.DB.MaxOpenConns != 21 {
		t.Fatalf("DB_MAX_OPEN_CONNS = %d, want 21", envs.DB.MaxOpenConns)
	}
	if envs.DB.MaxIdleConns != 7 {
		t.Fatalf("DB_MAX_IDLE_CONNS = %d, want 7", envs.DB.MaxIdleConns)
	}
	if envs.DB.ConnMaxLifetimeSeconds != 901 {
		t.Fatalf("DB_CONN_MAX_LIFETIME_SECONDS = %d, want 901", envs.DB.ConnMaxLifetimeSeconds)
	}
	if envs.DB.PingTimeoutSeconds != 8 {
		t.Fatalf("DB_PING_TIMEOUT_SECONDS = %d, want 8", envs.DB.PingTimeoutSeconds)
	}
}

func TestLoad_ServerAndMetricsFromEnv(t *testing.T) {
	t.Setenv("PORT", "8081")
	t.Setenv("SERVER_BASE_PATH", "/v1/api")
	t.Setenv("SERVER_SHUTDOWN_TIMEOUT", "42")
	t.Setenv("SERVER_READ_TIMEOUT_SECONDS", "6")
	t.Setenv("SERVER_WRITE_TIMEOUT_SECONDS", "16")
	t.Setenv("SERVER_IDLE_TIMEOUT_SECONDS", "61")
	t.Setenv("CORS_ORIGINS", "https://app.example.com; https://admin.example.com")
	t.Setenv("METRICS_ENABLED", "false")
	t.Setenv("METRICS_PORT", "9100")
	t.Setenv("METRICS_PATH", "/m")
	t.Setenv("APP_METRICS_PREFIX", "custom_prefix")

	envs := Load()
	if envs.SERVER == nil {
		t.Fatal("SERVER envs should not be nil")
	}
	if envs.SERVER.Port != "8081" {
		t.Fatalf("PORT = %q, want 8081", envs.SERVER.Port)
	}
	if envs.SERVER.BasePath != "/v1/api" {
		t.Fatalf("SERVER_BASE_PATH = %q, want /v1/api", envs.SERVER.BasePath)
	}
	if envs.SERVER.ShutdownTimeout != 42 {
		t.Fatalf("SERVER_SHUTDOWN_TIMEOUT = %d, want 42", envs.SERVER.ShutdownTimeout)
	}
	if envs.SERVER.ReadTimeoutSeconds != 6 {
		t.Fatalf("SERVER_READ_TIMEOUT_SECONDS = %d, want 6", envs.SERVER.ReadTimeoutSeconds)
	}
	if envs.SERVER.WriteTimeoutSeconds != 16 {
		t.Fatalf("SERVER_WRITE_TIMEOUT_SECONDS = %d, want 16", envs.SERVER.WriteTimeoutSeconds)
	}
	if envs.SERVER.IdleTimeoutSeconds != 61 {
		t.Fatalf("SERVER_IDLE_TIMEOUT_SECONDS = %d, want 61", envs.SERVER.IdleTimeoutSeconds)
	}
	if envs.CORS == nil {
		t.Fatal("CORS envs should not be nil")
	}
	if envs.CORS.Origins != "https://app.example.com; https://admin.example.com" {
		t.Fatalf("CORS_ORIGINS = %q", envs.CORS.Origins)
	}
	if envs.METRICS == nil {
		t.Fatal("METRICS envs should not be nil")
	}
	if envs.METRICS.Enabled {
		t.Fatal("METRICS_ENABLED = true, want false")
	}
	if envs.METRICS.Port != "9100" {
		t.Fatalf("METRICS_PORT = %q, want 9100", envs.METRICS.Port)
	}
	if envs.METRICS.Path != "/m" {
		t.Fatalf("METRICS_PATH = %q, want /m", envs.METRICS.Path)
	}
	if envs.METRICS.Prefix != "custom_prefix" {
		t.Fatalf("APP_METRICS_PREFIX = %q, want custom_prefix", envs.METRICS.Prefix)
	}
}

func TestLoad_HealthReadinessFromEnv(t *testing.T) {
	t.Setenv("HEALTH_READY_CACHE_TTL_MS", "450")
	t.Setenv("HEALTH_READY_CB_FAILURE_THRESHOLD", "7")
	t.Setenv("HEALTH_READY_CB_TIMEOUT_SECONDS", "9")
	t.Setenv("HEALTH_READY_CB_SUCCESS_THRESHOLD", "2")
	t.Setenv("HEALTH_READY_CB_HALF_OPEN_MAX_CONCURRENT", "3")

	envs := Load()
	if envs.HEALTH == nil {
		t.Fatal("HEALTH envs should not be nil")
	}
	if envs.HEALTH.ReadyCacheTTLMS != 450 {
		t.Fatalf("HEALTH_READY_CACHE_TTL_MS = %d, want 450", envs.HEALTH.ReadyCacheTTLMS)
	}
	if envs.HEALTH.ReadyCBFailureThreshold != 7 {
		t.Fatalf("HEALTH_READY_CB_FAILURE_THRESHOLD = %d, want 7", envs.HEALTH.ReadyCBFailureThreshold)
	}
	if envs.HEALTH.ReadyCBTimeoutSeconds != 9 {
		t.Fatalf("HEALTH_READY_CB_TIMEOUT_SECONDS = %d, want 9", envs.HEALTH.ReadyCBTimeoutSeconds)
	}
	if envs.HEALTH.ReadyCBSuccessThreshold != 2 {
		t.Fatalf("HEALTH_READY_CB_SUCCESS_THRESHOLD = %d, want 2", envs.HEALTH.ReadyCBSuccessThreshold)
	}
	if envs.HEALTH.ReadyCBHalfOpenConcurrent != 3 {
		t.Fatalf("HEALTH_READY_CB_HALF_OPEN_MAX_CONCURRENT = %d, want 3", envs.HEALTH.ReadyCBHalfOpenConcurrent)
	}
}
