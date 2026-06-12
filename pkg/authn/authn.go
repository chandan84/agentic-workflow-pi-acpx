// Package authn defines a tiny pluggable authentication interface used by
// every service's RPC layer.
//
// Two impls ship: None (allow all) and StaticToken (single shared secret).
// OIDC lands in a follow-up iteration; the interface is the seam.
package authn

import (
	"context"
	"errors"
	"strings"
)

// Principal is the result of a successful authentication.
type Principal struct {
	Subject string
	Scopes  []string
}

// Authenticator validates an inbound credential (typically the bearer header).
type Authenticator interface {
	Authenticate(ctx context.Context, credential string) (Principal, error)
}

// None allows all callers and identifies them as "anonymous".
type None struct{}

// Authenticate always succeeds for None.
func (None) Authenticate(_ context.Context, _ string) (Principal, error) {
	return Principal{Subject: "anonymous"}, nil
}

// StaticToken accepts a single shared bearer token.
type StaticToken struct{ Token string }

// Authenticate validates credential matches the configured token.
func (s StaticToken) Authenticate(_ context.Context, credential string) (Principal, error) {
	token := strings.TrimPrefix(credential, "Bearer ")
	if s.Token == "" || token != s.Token {
		return Principal{}, errors.New("unauthorized")
	}
	return Principal{Subject: "static", Scopes: []string{"*"}}, nil
}

// From resolves an Authenticator from a config-style mode string.
func From(mode, token string) Authenticator {
	switch mode {
	case "static-token":
		return StaticToken{Token: token}
	default:
		return None{}
	}
}
