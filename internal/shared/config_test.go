package config

import (
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	// go test runs in the package directory, so the repo's .env is not loaded here.
	t.Run("missing all", func(t *testing.T) {
		t.Setenv("YAHOO_CLIENT_ID", "")
		t.Setenv("YAHOO_CLIENT_SECRET", "")
		t.Setenv("YAHOO_REDIRECT_URI", "")
		t.Setenv("PLAYERS_SOURCE", "")

		_, err := Load()
		if err == nil {
			t.Fatal("expected an error, got nil")
		}
		for _, name := range []string{"YAHOO_CLIENT_ID", "YAHOO_CLIENT_SECRET", "YAHOO_REDIRECT_URI"} {
			if !strings.Contains(err.Error(), name) {
				t.Errorf("error should name %s, got: %v", name, err)
			}
		}
	})

	t.Run("all set", func(t *testing.T) {
		t.Setenv("YAHOO_CLIENT_ID", "id")
		t.Setenv("YAHOO_CLIENT_SECRET", "secret")
		t.Setenv("YAHOO_REDIRECT_URI", "https://example.test/callback")
		t.Setenv("PLAYERS_SOURCE", "")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.ClientID != "id" || cfg.ClientSecret != "secret" {
			t.Errorf("got %+v", cfg)
		}
		if cfg.RedirectURI != "https://example.test/callback" {
			t.Errorf("RedirectURI = %q", cfg.RedirectURI)
		}
		if cfg.PlayersSource != "yahoo" {
			t.Errorf("PlayersSource = %q, want yahoo", cfg.PlayersSource)
		}
	})

	t.Run("mock players source", func(t *testing.T) {
		t.Setenv("YAHOO_CLIENT_ID", "id")
		t.Setenv("YAHOO_CLIENT_SECRET", "secret")
		t.Setenv("YAHOO_REDIRECT_URI", "https://example.test/callback")
		t.Setenv("PLAYERS_SOURCE", " MOCK ")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.PlayersSource != "mock" {
			t.Errorf("PlayersSource = %q, want mock", cfg.PlayersSource)
		}
	})

	t.Run("invalid players source", func(t *testing.T) {
		t.Setenv("YAHOO_CLIENT_ID", "id")
		t.Setenv("YAHOO_CLIENT_SECRET", "secret")
		t.Setenv("YAHOO_REDIRECT_URI", "https://example.test/callback")
		t.Setenv("PLAYERS_SOURCE", "fixture")

		_, err := Load()
		if err == nil || !strings.Contains(err.Error(), "PLAYERS_SOURCE") {
			t.Fatalf("err = %v, want PLAYERS_SOURCE error", err)
		}
	})
}
