// Package session verifies that the app's Yahoo token is still accepted
// rather than merely present.
package session

import (
	"context"
	"errors"

	"fantasy_waiver_assistant/internal/auth"
	"fantasy_waiver_assistant/internal/yahoo"
)

// Status is the result of checking the current Yahoo session.
type Status uint8

const (
	// Unavailable means Yahoo could not be reached, so the token was preserved.
	Unavailable Status = iota
	// LoggedOut means there is no usable token.
	LoggedOut
	// AccessDenied means the token is valid but the Yahoo app lacks the
	// permission needed for the request.
	AccessDenied
	// Verified means Yahoo accepted the token for the current user.
	Verified
)

func (s Status) String() string {
	switch s {
	case LoggedOut:
		return "logged_out"
	case AccessDenied:
		return "access_denied"
	case Verified:
		return "verified"
	default:
		return "unavailable"
	}
}

// Tokens is the OAuth token lifecycle needed to validate a session.
type Tokens interface {
	LoggedIn() bool
	Refresh(ctx context.Context) error
	Clear()
}

// UserSource proves that Yahoo accepts the current token.
type UserSource interface {
	CurrentUser(ctx context.Context) (*yahoo.User, error)
}

// Validator checks a token against Yahoo and refreshes it once when rejected.
type Validator struct {
	tokens Tokens
	users  UserSource
}

// New returns a Validator backed by tokens and users.
func New(tokens Tokens, users UserSource) *Validator {
	return &Validator{tokens: tokens, users: users}
}

// Validate reports whether the current token is usable. Only a rejected token
// is cleared; network and Yahoo server failures leave it in place.
func (v *Validator) Validate(ctx context.Context) Status {
	if !v.tokens.LoggedIn() {
		return LoggedOut
	}

	_, err := v.users.CurrentUser(ctx)
	if err == nil {
		return Verified
	}
	if errors.Is(err, yahoo.ErrForbidden) {
		return AccessDenied
	}
	if !errors.Is(err, yahoo.ErrUnauthorized) {
		return Unavailable
	}

	if err := v.tokens.Refresh(ctx); err != nil {
		switch {
		case errors.Is(err, auth.ErrRejected), errors.Is(err, auth.ErrNotLoggedIn):
			v.tokens.Clear()
			return LoggedOut
		default:
			return Unavailable
		}
	}

	if _, err := v.users.CurrentUser(ctx); err != nil {
		if errors.Is(err, yahoo.ErrForbidden) {
			return AccessDenied
		}
		if errors.Is(err, yahoo.ErrUnauthorized) {
			v.tokens.Clear()
			return LoggedOut
		}
		return Unavailable
	}
	return Verified
}

type statusKey struct{}

// WithStatus records a validation result for the rest of a request.
func WithStatus(ctx context.Context, status Status) context.Context {
	return context.WithValue(ctx, statusKey{}, status)
}

// FromContext returns a validation result recorded earlier in the request.
func FromContext(ctx context.Context) (Status, bool) {
	status, ok := ctx.Value(statusKey{}).(Status)
	return status, ok
}
