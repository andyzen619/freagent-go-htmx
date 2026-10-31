package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
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

// Client holds the OAuth credentials and the current token.
type Client struct {
	cfg   config.Config
	token *Token
}

func New(cfg config.Config) *Client {
	return &Client{cfg: cfg}
}

func (c *Client) LoggedIn() bool {
	return c.token != nil
}

// Token returns the current token, or nil when not logged in.
func (c *Client) Token() *Token {
	return c.token
}

// Save writes the token as JSON to path, owned and readable by the user only.
func (c *Client) Save(path string) error {
	if c.token == nil {
		return fmt.Errorf("not logged in")
	}
	data, err := json.MarshalIndent(c.token, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func (c *Client) AuthCodeURL(state string) string {
	queryParams := url.Values{
		"client_id":     {c.cfg.ClientID},
		"redirect_uri":  {c.cfg.RedirectURI},
		"response_type": {"code"},
		"scope":         {"openid fspt-r"},
		"state":         {state},
	}

	return "https://api.login.yahoo.com/oauth2/request_auth?" + queryParams.Encode()
}

func (c *Client) Exchange(ctx context.Context, code string) error {
	return c.postToken(ctx, url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"redirect_uri": {c.cfg.RedirectURI},
	})
}

func (c *Client) Refresh(ctx context.Context) error {
	if !c.LoggedIn() {
		return fmt.Errorf("not logged in")
	}

	return c.postToken(ctx, url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {c.token.RefreshToken},
	})
}

func (c *Client) postToken(ctx context.Context, form url.Values) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.login.yahoo.com/oauth2/get_token", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}

	req.SetBasicAuth(c.cfg.ClientID, c.cfg.ClientSecret)
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
	c.token = &body.Token
	return nil
}
