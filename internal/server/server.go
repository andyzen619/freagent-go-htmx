// Package server is the composition root: it builds the app's dependencies and
// registers every route.
package server

import (
	"context"
	"fmt"
	"net/http"

	"fantasy_waiver_assistant/internal/infra/auth"
	"fantasy_waiver_assistant/internal/infra/config"
	"fantasy_waiver_assistant/internal/infra/yahoo"
	"fantasy_waiver_assistant/internal/pages/home"
	"fantasy_waiver_assistant/internal/pages/players"
	"fantasy_waiver_assistant/internal/service/freeagents"
	"fantasy_waiver_assistant/internal/service/session"
	"fantasy_waiver_assistant/internal/web"
)

// Run loads the configuration, wires the dependencies and serves on :8080.
func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	client := auth.New(cfg)
	yahooClient := yahoo.NewClient(func(ctx context.Context) (string, error) {
		if token := client.Token(); token != nil {
			return token.AccessToken, nil
		}
		return "", yahoo.ErrUnauthorized
	})
	validator := session.New(client, yahooClient)

	mux := http.NewServeMux()

	// The home page is public so it can report a logged-out or unverified
	// session. /login and /callback are public as well.
	homePage := home.New(validator)
	homePage.RegisterPublic(mux)

	// app holds pages that need a verified Yahoo session.
	app := http.NewServeMux()
	homePage.Register(app)
	if cfg.PlayersSource == "mock" {
		// Mock players are deterministic local data and need no Yahoo session,
		// so keep them public for UI development.
		players.NewMock(freeagents.NewMockSource()).Register(mux)
	} else {
		players.New(freeagents.NewYahooSource(yahooClient, cfg.LeagueKey)).Register(app)
	}

	// Every route on app first checks the token against Yahoo. Unknown paths
	// fall through to the home page once the session is verified.
	mux.Handle("/", web.RequireSession(validator, app))

	mux.HandleFunc("/login", client.Login)
	mux.HandleFunc("/callback", client.Callback)

	fmt.Println("Server Running on 8080")
	return http.ListenAndServe(":8080", mux)
}
