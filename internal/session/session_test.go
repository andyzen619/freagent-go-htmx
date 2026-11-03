package session

import (
	"context"
	"errors"
	"testing"

	"fantasy_waiver_assistant/internal/auth"
	"fantasy_waiver_assistant/internal/yahoo"
)

type stubTokens struct {
	loggedIn   bool
	refreshErr error
	refreshes  int
	cleared    bool
}

func (s *stubTokens) LoggedIn() bool { return s.loggedIn }

func (s *stubTokens) Refresh(context.Context) error {
	s.refreshes++
	return s.refreshErr
}

func (s *stubTokens) Clear() {
	s.cleared = true
	s.loggedIn = false
}

type stubUsers struct {
	results []error
	calls   int
}

func (s *stubUsers) CurrentUser(context.Context) (*yahoo.User, error) {
	s.calls++
	if len(s.results) == 0 {
		return &yahoo.User{GUID: "user"}, nil
	}
	err := s.results[0]
	s.results = s.results[1:]
	if err != nil {
		return nil, err
	}
	return &yahoo.User{GUID: "user"}, nil
}

func TestValidateLoggedOutWithoutToken(t *testing.T) {
	users := &stubUsers{}
	v := New(&stubTokens{}, users)

	if got := v.Validate(context.Background()); got != LoggedOut {
		t.Errorf("status = %v, want LoggedOut", got)
	}
	if users.calls != 0 {
		t.Errorf("CurrentUser calls = %d, want 0", users.calls)
	}
}

func TestValidateVerified(t *testing.T) {
	users := &stubUsers{}
	v := New(&stubTokens{loggedIn: true}, users)

	if got := v.Validate(context.Background()); got != Verified {
		t.Errorf("status = %v, want Verified", got)
	}
	if users.calls != 1 {
		t.Errorf("CurrentUser calls = %d, want 1", users.calls)
	}
}

func TestValidateAccessDeniedPreservesToken(t *testing.T) {
	tokens := &stubTokens{loggedIn: true}
	users := &stubUsers{results: []error{yahoo.ErrForbidden}}
	v := New(tokens, users)

	if got := v.Validate(context.Background()); got != AccessDenied {
		t.Errorf("status = %v, want AccessDenied", got)
	}
	if tokens.cleared {
		t.Error("permission failure should not clear the token")
	}
	if tokens.refreshes != 0 {
		t.Errorf("refreshes = %d, want 0", tokens.refreshes)
	}
}

func TestValidateRefreshesOnce(t *testing.T) {
	tokens := &stubTokens{loggedIn: true}
	users := &stubUsers{results: []error{yahoo.ErrUnauthorized}}
	v := New(tokens, users)

	if got := v.Validate(context.Background()); got != Verified {
		t.Errorf("status = %v, want Verified", got)
	}
	if tokens.refreshes != 1 {
		t.Errorf("refreshes = %d, want 1", tokens.refreshes)
	}
	if tokens.cleared {
		t.Error("verified session should not be cleared")
	}
	if users.calls != 2 {
		t.Errorf("CurrentUser calls = %d, want 2", users.calls)
	}
}

func TestValidateRejectedRefreshLogsOut(t *testing.T) {
	tokens := &stubTokens{loggedIn: true, refreshErr: auth.ErrRejected}
	users := &stubUsers{results: []error{yahoo.ErrUnauthorized}}
	v := New(tokens, users)

	if got := v.Validate(context.Background()); got != LoggedOut {
		t.Errorf("status = %v, want LoggedOut", got)
	}
	if !tokens.cleared {
		t.Error("rejected token should be cleared")
	}
}

func TestValidateRejectedRetryLogsOut(t *testing.T) {
	tokens := &stubTokens{loggedIn: true}
	users := &stubUsers{results: []error{yahoo.ErrUnauthorized, yahoo.ErrUnauthorized}}
	v := New(tokens, users)

	if got := v.Validate(context.Background()); got != LoggedOut {
		t.Errorf("status = %v, want LoggedOut", got)
	}
	if !tokens.cleared {
		t.Error("twice-rejected token should be cleared")
	}
}

func TestValidateTransientRefreshPreservesToken(t *testing.T) {
	tokens := &stubTokens{loggedIn: true, refreshErr: errors.New("network down")}
	users := &stubUsers{results: []error{yahoo.ErrUnauthorized}}
	v := New(tokens, users)

	if got := v.Validate(context.Background()); got != Unavailable {
		t.Errorf("status = %v, want Unavailable", got)
	}
	if tokens.cleared {
		t.Error("transient failure should not clear the token")
	}
}

func TestValidateTransientAPIFailurePreservesToken(t *testing.T) {
	tokens := &stubTokens{loggedIn: true}
	users := &stubUsers{results: []error{errors.New("HTTP 503")}}
	v := New(tokens, users)

	if got := v.Validate(context.Background()); got != Unavailable {
		t.Errorf("status = %v, want Unavailable", got)
	}
	if tokens.cleared {
		t.Error("server failure should not clear the token")
	}
}

func TestStatusContext(t *testing.T) {
	ctx := WithStatus(context.Background(), Verified)

	got, ok := FromContext(ctx)
	if !ok || got != Verified {
		t.Errorf("FromContext = %v, %v, want Verified, true", got, ok)
	}
	if _, ok := FromContext(context.Background()); ok {
		t.Error("FromContext without a value should report false")
	}
}
