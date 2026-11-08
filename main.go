package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"net/http"

	"fantasy_waiver_assistant/internal/auth"
	"fantasy_waiver_assistant/internal/home"
	"fantasy_waiver_assistant/internal/middleware"
	"fantasy_waiver_assistant/internal/players"
	"fantasy_waiver_assistant/internal/session"
	config "fantasy_waiver_assistant/internal/shared"
	"fantasy_waiver_assistant/internal/yahoo"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	client := auth.New(cfg)

	yahooClient := yahoo.NewClient(func(context.Context) (string, error) {
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

	// app holds pages and fragments that need a verified Yahoo session.
	app := http.NewServeMux()
	homePage.Register(app)
	if cfg.PlayersSource == "mock" {
		// Mock players are deterministic local data and need no Yahoo session,
		// so keep them public for UI development.
		players.NewMock(players.NewMockSource()).Register(mux)
	} else {
		players.New(players.NewYahooSource(yahooClient, cfg.LeagueKey)).Register(app)
	}

	// Every route on app first checks the token against Yahoo. Unknown paths
	// fall through to the home page once the session is verified.
	mux.Handle("/", middleware.RequireSession(validator, app))

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		state := rand.Text()
		http.SetCookie(w, &http.Cookie{Name: "oauth_state", Value: state, Path: "/", HttpOnly: true})
		http.Redirect(w, r, client.AuthCodeURL(state), http.StatusFound)
	})

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("oauth_state")
		if err != nil || c.Value != r.URL.Query().Get("state") {
			http.Error(w, "invalid state", http.StatusBadRequest)
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "authorization denied", http.StatusBadRequest)
			return
		}
		if err := client.Exchange(r.Context(), code); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := client.Save(".tokens.json"); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusFound)
	})

	fmt.Println("Server Running on 8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Println(err)
	}
}
