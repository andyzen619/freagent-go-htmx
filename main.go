package main

import (
	"crypto/rand"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"fantasy_waiver_assistant/internal/auth"
	config "fantasy_waiver_assistant/internal/shared"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	client := auth.New(cfg)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("internal/home/templates/index.html"))
		w.Header().Set("Content-Type", "text/html")
		_ = tmpl.Execute(w, map[string]any{"LoggedIn": client.LoggedIn()})
	})

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		state := rand.Text()
		http.SetCookie(w, &http.Cookie{Name: "oauth_state", Value: state, Path: "/", HttpOnly: true})
		http.Redirect(w, r, client.AuthCodeURL(state), http.StatusFound)
	})

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
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
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println(err)
	}
}
