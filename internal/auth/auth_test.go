package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	config "fantasy_waiver_assistant/internal/shared"
)

func TestClientTokenIsNilBeforeLogin(t *testing.T) {
	c := New(config.Config{})

	if c.LoggedIn() {
		t.Error("fresh client should not be logged in")
	}
	if c.Token() != nil {
		t.Error("Token() should be nil before login")
	}
}

func TestClientClear(t *testing.T) {
	c := New(config.Config{})
	c.token = &Token{AccessToken: "abc"}

	c.Clear()

	if c.LoggedIn() || c.Token() != nil {
		t.Error("Clear should forget the token")
	}
}

func TestRefreshRejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"invalid_grant"}`, http.StatusBadRequest)
	}))
	t.Cleanup(srv.Close)

	c := New(config.Config{})
	c.tokenURL = srv.URL
	c.token = &Token{AccessToken: "expired", RefreshToken: "bad"}

	err := c.Refresh(context.Background())
	if !errors.Is(err, ErrRejected) {
		t.Fatalf("err = %v, want ErrRejected", err)
	}
}

func TestRefreshTransientFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "slow down", http.StatusTooManyRequests)
	}))
	t.Cleanup(srv.Close)

	c := New(config.Config{})
	c.tokenURL = srv.URL
	c.token = &Token{AccessToken: "expired", RefreshToken: "rt"}

	err := c.Refresh(context.Background())
	if err == nil || errors.Is(err, ErrRejected) {
		t.Fatalf("err = %v, want a transient error", err)
	}
}

func TestRefreshSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"new","token_type":"bearer","expires_in":3600}`))
	}))
	t.Cleanup(srv.Close)

	c := New(config.Config{})
	c.tokenURL = srv.URL
	c.token = &Token{AccessToken: "expired", RefreshToken: "rt"}

	if err := c.Refresh(context.Background()); err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	token := c.Token()
	if token == nil || token.AccessToken != "new" || token.RefreshToken != "rt" {
		t.Errorf("token = %+v", token)
	}
}

func TestClientSave(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".tokens.json")

	if err := New(config.Config{}).Save(path); err == nil {
		t.Fatal("Save without a token should fail")
	}

	c := New(config.Config{})
	c.token = &Token{AccessToken: "abc", RefreshToken: "rt", TokenType: "bearer"}
	if err := c.Save(path); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("file mode = %o, want 600", perm)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.Contains(string(data), `"access_token": "abc"`) {
		t.Errorf("saved file missing access token: %s", data)
	}
}
