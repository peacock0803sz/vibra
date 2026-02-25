package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"connectrpc.com/connect"
)

const (
	// Header injected by Tailscale Serve for tailnet-internal requests.
	headerTailscaleUserLogin = "Tailscale-User-Login"

	headerAuthorization = "Authorization"
	bearerPrefix        = "Bearer "
)

type contextKey struct{}

var userKey = contextKey{}

// UserFromContext extracts the authenticated user identity from the context.
func UserFromContext(ctx context.Context) (string, bool) {
	user, ok := ctx.Value(userKey).(string)
	return user, ok
}

// TokenValidator validates a Bearer token and returns the user identity if valid.
type TokenValidator func(token string) (string, error)

// InterceptorConfig holds configuration for the auth interceptor.
type InterceptorConfig struct {
	// ValidateToken validates Bearer tokens for external (Funnel) access.
	// If nil, Bearer token authentication is disabled.
	ValidateToken TokenValidator
	// DevUser, when non-empty, skips all authentication and uses this value
	// as the user identity. Intended for local development only.
	DevUser string
}

// authenticate extracts and validates the user identity from request headers.
// Priority: 1. DevUser (if configured)  2. Tailscale header  3. Bearer token
func authenticate(headers http.Header, cfg *InterceptorConfig) (string, error) {
	if cfg != nil && cfg.DevUser != "" {
		return cfg.DevUser, nil
	}

	if user := headers.Get(headerTailscaleUserLogin); user != "" {
		return user, nil
	}

	if auth := headers.Get(headerAuthorization); strings.HasPrefix(auth, bearerPrefix) {
		token := strings.TrimPrefix(auth, bearerPrefix)
		if cfg != nil && cfg.ValidateToken != nil {
			return cfg.ValidateToken(token)
		}
		return "", errors.New("bearer token authentication not configured")
	}

	return "", errors.New("missing authentication: expected Tailscale-User-Login header or Authorization: Bearer token")
}

// Interceptor is a ConnectRPC interceptor that enforces authentication
// on both unary and streaming RPCs.
type Interceptor struct {
	cfg *InterceptorConfig
}

// NewInterceptor creates a new auth interceptor with the given config.
func NewInterceptor(cfg *InterceptorConfig) *Interceptor {
	return &Interceptor{cfg: cfg}
}

func (a *Interceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		user, err := authenticate(req.Header(), a.cfg)
		if err != nil {
			return nil, connect.NewError(connect.CodeUnauthenticated, err)
		}
		ctx = context.WithValue(ctx, userKey, user)
		return next(ctx, req)
	}
}

func (a *Interceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	// Not used on server side.
	return next
}

func (a *Interceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		user, err := authenticate(conn.RequestHeader(), a.cfg)
		if err != nil {
			return connect.NewError(connect.CodeUnauthenticated, err)
		}
		ctx = context.WithValue(ctx, userKey, user)
		return next(ctx, conn)
	}
}
