// Package yahoo is a read-only client for the Yahoo Fantasy Sports API.
//
// Yahoo serves XML by default and that is what this client decodes, so callers
// never see the layout quirks of Yahoo's JSON transliteration (see
// docs/yahoo-api-shapes.md). Requests are made with a bearer token supplied by
// a TokenSource, or without one for public leagues.
package yahoo

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// DefaultBaseURL is the root of the Fantasy Sports API.
const DefaultBaseURL = "https://fantasysports.yahooapis.com/fantasy/v2"

// ErrUnauthorized means the access token is missing, expired or invalid. The
// caller can refresh the token and retry the request.
var ErrUnauthorized = errors.New("yahoo: unauthorized")

// TokenSource supplies the bearer token for a request. It may refresh an
// expired token before returning one.
type TokenSource func(ctx context.Context) (string, error)

// StaticToken returns a TokenSource that always returns token.
func StaticToken(token string) TokenSource {
	return func(context.Context) (string, error) { return token, nil }
}

// Client talks to the Yahoo Fantasy Sports API.
type Client struct {
	// BaseURL defaults to DefaultBaseURL.
	BaseURL string
	// HTTPClient defaults to http.DefaultClient.
	HTTPClient *http.Client
	// Token may be nil for public, unauthenticated reads.
	Token TokenSource
}

// NewClient returns a client for the Yahoo Fantasy Sports API.
func NewClient(token TokenSource) *Client {
	return &Client{
		BaseURL:    DefaultBaseURL,
		HTTPClient: http.DefaultClient,
		Token:      token,
	}
}

// League fetches a league resource. Pass sub-resources in out to include them
// in the same request, e.g. "settings" or "teams".
func (c *Client) League(ctx context.Context, leagueKey string, out ...string) (*League, error) {
	env, err := c.get(ctx, leaguePath(leagueKey, out...))
	if err != nil {
		return nil, err
	}
	if env.League == nil {
		return nil, fmt.Errorf("yahoo: response has no league")
	}
	return env.League, nil
}

// LeagueSettings fetches only the league's settings.
func (c *Client) LeagueSettings(ctx context.Context, leagueKey string) (*LeagueSettings, error) {
	env, err := c.get(ctx, leagueSettingsPath(leagueKey))
	if err != nil {
		return nil, err
	}
	if env.League == nil || env.League.Settings == nil {
		return nil, fmt.Errorf("yahoo: response has no league settings")
	}
	return env.League.Settings, nil
}

// LeaguePlayers fetches a league's players collection, filtered and paginated
// by f. Use StatusFreeAgents or StatusWaivers for waiver-wire queries.
func (c *Client) LeaguePlayers(ctx context.Context, leagueKey string, f PlayerFilter) (*Players, error) {
	env, err := c.get(ctx, leaguePlayersPath(leagueKey, f))
	if err != nil {
		return nil, err
	}
	if env.League == nil || env.League.Players == nil {
		return nil, fmt.Errorf("yahoo: response has no players")
	}
	return env.League.Players, nil
}

// Players fetches specific players by key.
func (c *Client) Players(ctx context.Context, playerKeys ...string) (*Players, error) {
	if len(playerKeys) == 0 {
		return nil, fmt.Errorf("yahoo: no player keys given")
	}
	env, err := c.get(ctx, playersPath(playerKeys))
	if err != nil {
		return nil, err
	}
	if env.Players == nil {
		return nil, fmt.Errorf("yahoo: response has no players")
	}
	return env.Players, nil
}

// Player fetches one player. Pass sub-resources in out to include them, e.g.
// "stats" or "percent_owned".
func (c *Client) Player(ctx context.Context, playerKey string, out ...string) (*Player, error) {
	env, err := c.get(ctx, playerPath(playerKey, out...))
	if err != nil {
		return nil, err
	}
	if env.Player == nil {
		return nil, fmt.Errorf("yahoo: response has no player")
	}
	return env.Player, nil
}

// get performs one GET request and decodes the XML response.
func (c *Client) get(ctx context.Context, path string) (*Envelope, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL()+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/xml")

	if c.Token != nil {
		token, err := c.Token(ctx)
		if err != nil {
			return nil, err
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	}

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}
	if resp.StatusCode != http.StatusOK {
		return nil, httpStatusError(resp)
	}

	env, err := decodeResponse(resp.Body)
	if err != nil {
		return nil, err
	}
	if env.Error != nil {
		return nil, env.Error
	}
	return env, nil
}

func (c *Client) baseURL() string {
	if c.BaseURL == "" {
		return DefaultBaseURL
	}
	return strings.TrimRight(c.BaseURL, "/")
}

// httpStatusError turns a non-200 response into an *APIError when the body is a
// Yahoo error document, and a plain status error otherwise.
func httpStatusError(resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	env, decodeErr := decodeResponse(bytes.NewReader(body))
	if decodeErr == nil && env.Error != nil {
		return env.Error
	}
	var apiErr *APIError
	if errors.As(decodeErr, &apiErr) {
		return apiErr
	}

	snippet := strings.TrimSpace(string(body))
	if len(snippet) > 200 {
		snippet = snippet[:200] + "..."
	}
	if snippet == "" {
		return fmt.Errorf("yahoo: HTTP %s", resp.Status)
	}
	return fmt.Errorf("yahoo: HTTP %s: %s", resp.Status, snippet)
}
