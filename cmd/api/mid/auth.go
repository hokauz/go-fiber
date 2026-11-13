package mid

import (
	"context"
	"fmt"
	"strings"
	"time"

	jwksadapter "go-fiber/adapter/jwks"
	jwtadapter "go-fiber/adapter/jwt/golangjwt"
	"go-fiber/cmd/api/env"
	"go-fiber/pkg/auth"
	"go-fiber/pkg/httputil"
	"go-fiber/pkg/logx"
	"go-fiber/pkg/reqctx"

	"github.com/gofiber/fiber/v3"
)

const (
	defaultWarmupTimeout = 3 * time.Second
)

// AuthOptions controls startup warmup behavior for JWT auth.
type AuthOptions struct {
	WarmupOnStart bool
	WarmupTimeout time.Duration
	WarmupAsync   bool
	OnWarmupError func(error)
}

// DefaultAuthOptions returns best-effort async warmup defaults.
func DefaultAuthOptions() AuthOptions {
	return AuthOptions{
		WarmupOnStart: true,
		WarmupAsync:   true,
		WarmupTimeout: defaultWarmupTimeout,
		OnWarmupError: func(err error) {
			logx.Default().Error("auth jwks warmup failed", logx.Error(err))
		},
	}
}

// Auth builds verifier with default options and returns the auth middleware.
func Auth(envs *env.AuthEnvs) fiber.Handler {
	return AuthWithOptions(envs, DefaultAuthOptions())
}

// AuthWithOptions builds verifier, schedules warmup, and returns the Fiber auth middleware.
// If verifier cannot be created (e.g. missing config), logs and returns a handler that
// responds with 500 on protected routes.
func AuthWithOptions(envs *env.AuthEnvs, opts AuthOptions) fiber.Handler {
	verifier, err := NewVerifier(envs)
	if err != nil {
		logx.Default().Error("auth verifier not configured", logx.Error(err))
		return RequireAuthFiber(nil)
	}

	startWarmup(verifier, opts)

	return RequireAuthFiber(verifier)
}

// RequireAuthFiber validates bearer JWT and populates request principal.
// Note: c.Matched() is false here because this middleware runs when a Use (prefix) route
// matches, before the actual route (Get/Post/etc.) is matched. Fiber only sets Matched=true
// for non-Use routes. Since this middleware is only applied to the private group (which
// contains only protected routes), we always validate.
func RequireAuthFiber(verifier auth.Verifier) fiber.Handler {
	return func(c fiber.Ctx) error {
		principal, err := auth.ValidateAuth(c.Context(), verifier, c.Get(fiber.HeaderAuthorization))

		if err != nil {
			return authErrorToFiber(c, err)
		}

		reqctx.SetPrincipal(c, principal)
		return c.Next()
	}
}

func authErrorToFiber(c fiber.Ctx, err error) error {
	ae, ok := auth.IsAuthError(err)
	if !ok {
		return httputil.InternalError(c, "auth/internal-server-error")
	}
	switch ae.Code {
	case auth.ErrAuthMissingHeader.Code, auth.ErrAuthInvalidScheme.Code, auth.ErrAuthInvalidToken.Code, auth.ErrAuthRequired.Code:
		return httputil.Unauthorized(c, ae.Code)
	case auth.ErrAuthServiceUnavailable.Code:
		return httputil.ServiceUnavailable(c, ae.Code)
	default:
		return httputil.InternalError(c, ae.Code)
	}
}

// NewVerifier creates a JWT verifier using current auth env settings.
func NewVerifier(envs *env.AuthEnvs) (auth.Verifier, error) {
	if envs == nil {
		return nil, fmt.Errorf("%w: missing auth envs", auth.ErrVerifierConfig)
	}

	engine := jwtadapter.New()
	keys, err := jwksadapter.NewProvider(jwksadapter.Config{
		JWKSURL:     strings.TrimSpace(envs.JWKSURL),
		CacheTTL:    time.Duration(envs.JWKSTTLSeconds) * time.Second,
		HTTPTimeout: time.Duration(envs.HTTPTimeoutSeconds) * time.Second,
	})
	if err != nil {
		return nil, err
	}

	return auth.NewOIDCVerifier(auth.OIDCConfig{
		Engine:      engine,
		Keys:        keys,
		AllowedAlgs: parseCSV(envs.AllowedAlgs),
		Issuer:      strings.TrimSpace(envs.Issuer),
		Audience:    strings.TrimSpace(envs.Audience),
	})
}

func startWarmup(verifier auth.Verifier, opts AuthOptions) {
	if verifier == nil || !opts.WarmupOnStart {
		return
	}

	if opts.WarmupTimeout <= 0 {
		opts.WarmupTimeout = defaultWarmupTimeout
	}
	if opts.OnWarmupError == nil {
		opts.OnWarmupError = func(err error) {
			logx.Default().Error("auth jwks warmup failed", logx.Error(err))
		}
	}

	run := func() {
		ctx, cancel := context.WithTimeout(context.Background(), opts.WarmupTimeout)
		defer cancel()

		if err := verifier.Warmup(ctx); err != nil {
			opts.OnWarmupError(err)
		}
	}

	if opts.WarmupAsync {
		go run()
		return
	}

	run()
}

func parseCSV(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
