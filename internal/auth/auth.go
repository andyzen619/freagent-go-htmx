package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	config "fantasy_waiver_assistant/internal/shared"
)

const tokenURL = "https://api.login.yahoo.com/oauth2/get_token"

var (
	// ErrNotLoggedIn means there is no token to use or refresh.
	ErrNotLoggedIn = errors.New("not logged in")
	// ErrRejected means Yahoo refused the authorization code or refresh token.
	// A new interactive login is required; transient failures do not wrap this.
	ErrRejected = errors.New("yahoo: authorization rejected")
)

type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresAt    time.Time
}

// Client holds the OAuth credentials and the current token.
type Client struct {
	cfg      config.Config
	mu       sync.RWMutex
	token    *Token
	tokenURL string
}

func New(cfg config.Config) *Client {
	return &Client{cfg: cfg, tokenURL: tokenURL}
}

func (c *Client) LoggedIn() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.token != nil && c.token.AccessToken != ""
}

// Token returns a copy of the current token, or nil when not logged in.
func (c *Client) Token() *Token {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.token == nil {
		return nil
	}
	token := *c.token
	return &token
}

// Clear forgets the current token.
func (c *Client) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.token = nil
}

// Save writes the token as JSON to path, owned and readable by the user only.
func (c *Client) Save(path string) error {
	token := c.Token()
	if token == nil {
		return ErrNotLoggedIn
	}
	data, err := json.MarshalIndent(token, "", "  ")
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
		return ErrNotLoggedIn
	}

	c.mu.RLock()
	if c.token == nil {
		c.mu.RUnlock()
		return ErrNotLoggedIn
	}
	refreshToken := c.token.RefreshToken
	c.mu.RUnlock()
	if refreshToken == "" {
		return fmt.Errorf("%w: no refresh token", ErrRejected)
	}

	return c.postToken(ctx, url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
	})
}

func (c *Client) postToken(ctx context.Context, form url.Values) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.tokenURL, strings.NewReader(form.Encode()))
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

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		err := fmt.Errorf("yahoo: token endpoint HTTP %s: %s", resp.Status, strings.TrimSpace(string(body)))
		switch resp.StatusCode {
		case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden:
			return fmt.Errorf("%w: %v", ErrRejected, err)
		default:
			return err
		}
	}

	var body struct {
		Token
		ExpiresIn int `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return fmt.Errorf("yahoo: decode token response: %w", err)
	}
	if body.AccessToken == "" {
		return fmt.Errorf("%w: no access token in response", ErrRejected)
	}
	if body.RefreshToken == "" {
		body.RefreshToken = form.Get("refresh_token")
	}
	body.Token.ExpiresAt = time.Now().Add(time.Duration(body.ExpiresIn) * time.Second)

	c.mu.Lock()
	defer c.mu.Unlock()
	c.token = &body.Token
	return nil
}
