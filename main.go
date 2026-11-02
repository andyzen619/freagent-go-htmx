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
	config "fantasy_waiver_assistant/internal/shared"
	"fantasy_waiver_assistant/internal/yahoo"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	client := auth.New(cfg)

	mux := http.NewServeMux()

	// app holds pages and fragments that need a Yahoo session; /login and
	// /callback stay public on mux, which matches their patterns first.
	app := http.NewServeMux()
	home.New(client).Register(app)

	// Yahoo calls use the in-memory OAuth token. Every route on app is gated by
	// RequireSession, so a token is present by the time a handler runs.
	yahooClient := yahoo.NewClient(func(context.Context) (string, error) {
		if token := client.Token(); token != nil {
			return token.AccessToken, nil
		}
		return "", yahoo.ErrUnauthorized
	})
	players.New(players.NewYahooSource(yahooClient, cfg.LeagueKey)).Register(app)

	mux.Handle("/", middleware.RequireSession(client, app))

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
		fmt.Println("access token:", client.Token().AccessToken)
		http.Redirect(w, r, "/", http.StatusFound)
	})

	fmt.Println("Server Running on 8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Println(err)
	}
}
