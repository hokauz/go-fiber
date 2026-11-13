package main

import (
	"github.com/gofiber/fiber/v3"

	"go-fiber/pkg/router"

	apiacl "go-fiber/cmd/api/acl"
	"go-fiber/cmd/api/docs"
	"go-fiber/cmd/api/env"
	"go-fiber/cmd/api/mid"

	"go-fiber/cmd/api/health"
	"go-fiber/internal/hello"
)

// @title        Go Fiber API
// @version      1.0
// @description  API with health probes and Scalar docs.
func main() {
	env := env.Load()
	appLog := newLogger(env)
	database := newDatabase(env, appLog)

	r := fiber.New(apiRouterConfig(env.SERVER))

	metrics := mid.NewMetrics(env.METRICS, database)
	basePath := apiBasePath(env.SERVER)

	r.Use(mid.RecovererWith(appLog))
	r.Use(mid.RequestID())
	r.Use(mid.LogContext(appLog))
	r.Use(metrics.Handler())
	r.Use(mid.LoggerWith(appLog))
	r.Use(mid.Limiter())
	r.Use(mid.Idempotency())
	r.Use(mid.Helmet())
	r.Use(mid.CORS(env.CORS))

	public := router.NewGroup(r.Group(basePath))
	healthRegister, healthStop := health.PublicRegister(database, env)

	public.Register(healthRegister)
	public.Register(hello.PublicRegister(database))
	public.Register(apiacl.PublicRegister(database, metrics.Collector()))
	public.Register(docs.PublicRegister(env.IsProduction))

	private := router.NewGroup(r.Group(basePath))
	private.Use(mid.Auth(env.AUTH))

	private.Register(hello.PrivateRegister(database))
	private.Register(apiacl.PrivateRegister(database, metrics.Collector()))

	startServerOrFatal(r, database, appLog, env.SERVER, metrics.Server(), healthStop)
}
