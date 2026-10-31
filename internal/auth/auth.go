package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	config "fantasy_waiver_assistant/internal/shared"
)

type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresAt    time.Time
}

var (
	settings config.Config
	token    *Token
)

// Init provides the OAuth credentials. Call it once at startup.
func Init(cfg config.Config) {
	settings = cfg
}

func LoggedIn() bool {
	return token != nil
}

func AuthCodeURL(state string) string {
	queryParams := url.Values{
		"client_id":     {settings.ClientID},
		"redirect_uri":  {settings.RedirectURI},
		"response_type": {"code"},
		"scope":         {"openid fspt-r"},
		"state":         {state},
	}

	return "https://api.login.yahoo.com/oauth2/request_auth?" + queryParams.Encode()
}

func Exchange(ctx context.Context, code string) error {
	return postToken(ctx, url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"redirect_uri": {settings.RedirectURI},
	})
}

func Refresh(ctx context.Context) error {
	if !LoggedIn() {
		return fmt.Errorf("not logged in")
	}

	return postToken(ctx, url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {token.RefreshToken},
	})
}

func postToken(ctx context.Context, form url.Values) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.login.yahoo.com/oauth2/get_token", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}

	req.SetBasicAuth(settings.ClientID, settings.ClientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var body struct {
		Token
		ExpiresIn int `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return err
	}
	if body.AccessToken == "" {
		return fmt.Errorf("no access token in response")
	}
	if body.RefreshToken == "" {
		body.RefreshToken = form.Get("refresh_token")
	}
	body.Token.ExpiresAt = time.Now().Add(time.Duration(body.ExpiresIn) * time.Second)
	token = &body.Token
	return nil
}
