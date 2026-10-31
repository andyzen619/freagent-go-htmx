package auth

import (
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
