package players

import (
	"context"
	"errors"

	"fantasy_waiver_assistant/internal/yahoo"
)

// ErrNoLeague means YAHOO_LEAGUE_KEY is missing from .env.
var ErrNoLeague = errors.New("YAHOO_LEAGUE_KEY is not set in .env")

// Player is one row on the players page.
type Player struct {
	Name     string
	Position string
	Team     string
}

// YahooSource reads free agents from the Yahoo Fantasy API.
type YahooSource struct {
	client    *yahoo.Client
	leagueKey string
}

// NewYahooSource returns a Source for the given league, e.g. "470.l.123".
func NewYahooSource(client *yahoo.Client, leagueKey string) *YahooSource {
	return &YahooSource{client: client, leagueKey: leagueKey}
}

// FreeAgents returns up to 25 free agents, optionally narrowed to one position.
func (s *YahooSource) FreeAgents(ctx context.Context, pos string) ([]Player, error) {
	if s.leagueKey == "" {
		return nil, ErrNoLeague
	}
	found, err := s.client.LeaguePlayers(ctx, s.leagueKey, yahoo.PlayerFilter{
		Position: pos,
		Status:   yahoo.StatusFreeAgents,
		Count:    25, // Yahoo caps player pages at 25.
	})
	if err != nil {
		return nil, err
	}
	out := make([]Player, len(found.Players))
	for i, p := range found.Players {
		out[i] = Player{Name: p.Name.Full, Position: p.DisplayPosition, Team: p.EditorialTeamAbbr}
	}
	return out, nil
}
