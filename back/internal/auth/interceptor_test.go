package auth

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestAuthenticate_DevUser(t *testing.T) {
	// DevUser bypasses all authentication; no headers required.
	headers := http.Header{}
	cfg := &InterceptorConfig{DevUser: "dev@local"}

	user, err := authenticate(headers, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user != "dev@local" {
		t.Errorf("got user %q, want %q", user, "dev@local")
	}
}

func TestAuthenticate_DevUserPriority(t *testing.T) {
	// DevUser takes priority even when Tailscale header is present.
	headers := http.Header{}
	headers.Set(headerTailscaleUserLogin, "ts-user@example.com")
	cfg := &InterceptorConfig{DevUser: "dev@local"}

	user, err := authenticate(headers, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user != "dev@local" {
		t.Errorf("got user %q, want %q", user, "dev@local")
	}
}

func TestAuthenticate_TailscaleHeader(t *testing.T) {
	headers := http.Header{}
	headers.Set(headerTailscaleUserLogin, "user@example.com")

	user, err := authenticate(headers, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user != "user@example.com" {
		t.Errorf("got user %q, want %q", user, "user@example.com")
	}
}

func TestAuthenticate_BearerToken(t *testing.T) {
	headers := http.Header{}
	headers.Set(headerAuthorization, "Bearer valid-token-123")

	cfg := &InterceptorConfig{
		ValidateToken: func(token string) (string, error) {
			if token == "valid-token-123" {
				return "token-user@example.com", nil
			}
			return "", errors.New("invalid token")
		},
	}

	user, err := authenticate(headers, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user != "token-user@example.com" {
		t.Errorf("got user %q, want %q", user, "token-user@example.com")
	}
}

func TestAuthenticate_InvalidBearerToken(t *testing.T) {
	headers := http.Header{}
	headers.Set(headerAuthorization, "Bearer bad-token")

	cfg := &InterceptorConfig{
		ValidateToken: func(token string) (string, error) {
			return "", errors.New("invalid token")
		},
	}

	_, err := authenticate(headers, cfg)
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestAuthenticate_BearerWithoutValidator(t *testing.T) {
	headers := http.Header{}
	headers.Set(headerAuthorization, "Bearer some-token")

	_, err := authenticate(headers, nil)
	if err == nil {
		t.Fatal("expected error when validator is nil")
	}
}

func TestAuthenticate_NoCredentials(t *testing.T) {
	headers := http.Header{}

	_, err := authenticate(headers, nil)
	if err == nil {
		t.Fatal("expected error for missing credentials")
	}
}

func TestAuthenticate_TailscalePriority(t *testing.T) {
	// Tailscale header takes priority over Bearer token.
	headers := http.Header{}
	headers.Set(headerTailscaleUserLogin, "ts-user@example.com")
	headers.Set(headerAuthorization, "Bearer some-token")

	user, err := authenticate(headers, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user != "ts-user@example.com" {
		t.Errorf("got user %q, want %q", user, "ts-user@example.com")
	}
}

func TestUserFromContext(t *testing.T) {
	ctx := context.Background()

	if _, ok := UserFromContext(ctx); ok {
		t.Fatal("expected no user in empty context")
	}

	ctx = context.WithValue(ctx, userKey, "test-user")
	user, ok := UserFromContext(ctx)
	if !ok {
		t.Fatal("expected user in context")
	}
	if user != "test-user" {
		t.Errorf("got user %q, want %q", user, "test-user")
	}
}
