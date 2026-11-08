package config

import (
	"fmt"
	"os"
	"strings"
)

// Config holds the Yahoo OAuth settings needed to start the app.
type Config struct {
	ClientID      string
	ClientSecret  string
	RedirectURI   string
	LeagueKey     string // optional, e.g. "470.l.123"; the players page needs it
	PlayersSource string // "yahoo" (default) or "mock"
}

// Load reads .env (relative to the working directory, if present) and returns
// the configuration. All missing variables are reported in one error.
func Load() (Config, error) {
	if err := LoadDotEnv(".env"); err != nil {
		return Config{}, err
	}

	cfg := Config{
		ClientID:      os.Getenv("YAHOO_CLIENT_ID"),
		ClientSecret:  os.Getenv("YAHOO_CLIENT_SECRET"),
		RedirectURI:   os.Getenv("YAHOO_REDIRECT_URI"),
		LeagueKey:     os.Getenv("YAHOO_LEAGUE_KEY"),
		PlayersSource: strings.ToLower(strings.TrimSpace(os.Getenv("PLAYERS_SOURCE"))),
	}
	if cfg.PlayersSource == "" {
		cfg.PlayersSource = "yahoo"
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
	if cfg.PlayersSource != "yahoo" && cfg.PlayersSource != "mock" {
		return Config{}, fmt.Errorf(`PLAYERS_SOURCE must be "yahoo" or "mock", got %q`, cfg.PlayersSource)
	}
	return cfg, nil
}
