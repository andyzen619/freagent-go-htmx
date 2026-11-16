package nba

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// DefaultBaseURL is the root of the stats.nba.com endpoint host.
const DefaultBaseURL = "https://stats.nba.com/stats"

// userAgent mimics a browser. stats.nba.com filters requests that look like
// scripts, so a browser-like UA (plus Referer) is required.
const userAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"

var (
	// ErrForbidden means stats.nba.com rejected the request as not allowed,
	// which often means the source IP is blocked. Retrying immediately will
	// not help.
	ErrForbidden = errors.New("nba: forbidden")
	// ErrRateLimited means the caller exceeded the (undocumented) request
	// limit. Back off and retry later.
	ErrRateLimited = errors.New("nba: rate limited")
)

// Client talks to the stats.nba.com endpoints.
type Client struct {
	// BaseURL defaults to DefaultBaseURL.
	BaseURL string
	// HTTPClient defaults to http.DefaultClient.
	HTTPClient *http.Client
	// StatsToken, when set, is sent as the x-nba-stats-token header. It is
	// not usually required, but some requests are rejected without it.
	StatsToken string
}

// NewClient returns a Client for stats.nba.com.
func NewClient() *Client {
	return &Client{
		BaseURL:    DefaultBaseURL,
		HTTPClient: http.DefaultClient,
	}
}

// LeagueDashPlayerStats returns the per-player season stats result set for the
// given parameters. It is the league-wide box score behind NBA.com's player
// stats pages.
func (c *Client) LeagueDashPlayerStats(ctx context.Context, p Params) (*ResultSet, error) {
	sets, err := c.get(ctx, "leaguedashplayerstats", p.Values())
	if err != nil {
		return nil, err
	}
	for i := range sets {
		if sets[i].Name == "LeagueDashPlayerStats" {
			return &sets[i], nil
		}
	}
	return nil, fmt.Errorf("nba: response has no LeagueDashPlayerStats result set")
}

// get performs one GET request and decodes the JSON response.
func (c *Client) get(ctx context.Context, path string, query url.Values) ([]ResultSet, error) {
	u := c.baseURL() + "/" + path
	if query != nil {
		u += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Referer", "https://www.nba.com/")
	req.Header.Set("Origin", "https://www.nba.com")
	req.Header.Set("x-nba-stats-origin", "stats")
	if c.StatsToken != "" {
		req.Header.Set("x-nba-stats-token", c.StatsToken)
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

	if resp.StatusCode != http.StatusOK {
		return nil, httpStatusError(resp)
	}

	var out Response
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("nba: decode %s: %w", path, err)
	}
	return out.ResultSets, nil
}

func (c *Client) baseURL() string {
	if c.BaseURL == "" {
		return DefaultBaseURL
	}
	return strings.TrimRight(c.BaseURL, "/")
}

// httpStatusError turns a non-200 response into a typed error when the status
// is meaningful (403/429) and a plain status error otherwise.
func httpStatusError(resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	switch resp.StatusCode {
	case http.StatusForbidden:
		return fmt.Errorf("%w: %s", ErrForbidden, snippet(body))
	case http.StatusTooManyRequests:
		return fmt.Errorf("%w: %s", ErrRateLimited, snippet(body))
	default:
		return fmt.Errorf("nba: HTTP %s: %s", resp.Status, snippet(body))
	}
}

// snippet trims body to a short one-line description for error messages.
func snippet(body []byte) string {
	s := strings.TrimSpace(string(body))
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > 200 {
		s = s[:200] + "..."
	}
	if s == "" {
		return "empty body"
	}
	return s
}
