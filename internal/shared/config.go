package config

import (
	"fmt"
	"os"
	"strings"
)

// Config holds the Yahoo OAuth settings needed to start the app.
type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

// Load reads .env (relative to the working directory, if present) and returns
// the configuration. All missing variables are reported in one error.
func Load() (Config, error) {
	if err := LoadDotEnv(".env"); err != nil {
		return Config{}, err
	}

	cfg := Config{
		ClientID:     os.Getenv("YAHOO_CLIENT_ID"),
		ClientSecret: os.Getenv("YAHOO_CLIENT_SECRET"),
		RedirectURI:  os.Getenv("YAHOO_REDIRECT_URI"),
	}

	var missing []string
	if cfg.ClientID == "" {
		missing = append(missing, "YAHOO_CLIENT_ID")
	}
	if cfg.ClientSecret == "" {
		missing = append(missing, "YAHOO_CLIENT_SECRET")
	}
	if cfg.RedirectURI == "" {
		missing = append(missing, "YAHOO_REDIRECT_URI")
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required env: %s", strings.Join(missing, ", "))
	}
	return cfg, nil
}
