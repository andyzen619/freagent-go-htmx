package players

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"fantasy_waiver_assistant/internal/yahoo"
)

// ErrNoLeague means YAHOO_LEAGUE_KEY is missing from .env.
var ErrNoLeague = errors.New("YAHOO_LEAGUE_KEY is not set in .env")

const missingStat = "—"

// Stats is the compact, per-game comparison line shown for each player.
type Stats struct {
	Depth     string
	MPG       string
	FGPercent string
	FTPercent string
	ThreePM   string
	Points    string
	Rebounds  string
	Assists   string
	Steals    string
	Blocks    string
	Turnovers string
}

// Value returns the displayed value for a canonical stat key.
func (s Stats) Value(key string) string {
	switch key {
	case "DPT":
		return s.Depth
	case "MPG":
		return s.MPG
	case "FG%":
		return s.FGPercent
	case "FT%":
		return s.FTPercent
	case "3PM":
		return s.ThreePM
	case "PTS":
		return s.Points
	case "REB":
		return s.Rebounds
	case "AST":
		return s.Assists
	case "STL":
		return s.Steals
	case "BLK":
		return s.Blocks
	case "TO":
		return s.Turnovers
	default:
		return ""
	}
}

// Player is one row on the players page.
type Player struct {
	Name     string
	Position string
	Team     string
	Stats    Stats
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

// FreeAgents returns up to 25 free agents with season stats.
func (s *YahooSource) FreeAgents(ctx context.Context) ([]Player, error) {
	if s.leagueKey == "" {
		return nil, ErrNoLeague
	}
	found, err := s.client.LeaguePlayers(ctx, s.leagueKey, yahoo.PlayerFilter{
		Status: yahoo.StatusFreeAgents,
		Count:  25, // Yahoo caps player pages at 25.
		Out:    []string{"stats"},
	})
	if err != nil {
		return nil, err
	}
	out := make([]Player, len(found.Players))
	for i, p := range found.Players {
		out[i] = playerFromYahoo(p)
	}
	return out, nil
}

// Yahoo basketball stat ids. Current-user depth is not exposed by this API, so
// live rows leave Depth as missing rather than inventing a ranking.
const (
	statGamesPlayed = 0
	statMinutes     = 2
	statFGPercent   = 5
	statFTPercent   = 8
	statThreePM     = 10
	statPoints      = 12
	statRebounds    = 15
	statAssists     = 16
	statSteals      = 17
	statBlocks      = 18
	statTurnovers   = 19
)

func playerFromYahoo(p yahoo.Player) Player {
	stats := make(map[int]string)
	if p.PlayerStats != nil {
		for _, stat := range p.PlayerStats.Stats {
			stats[stat.StatID] = stat.Value
		}
	}

	games, hasGames := statNumber(stats, statGamesPlayed)
	perGame := func(id int) string {
		if !hasGames || games <= 0 {
			return missingStat
		}
		return divideStat(stats, id, games)
	}

	return Player{
		Name:     p.Name.Full,
		Position: p.DisplayPosition,
		Team:     p.EditorialTeamAbbr,
		Stats: Stats{
			Depth:     missingStat,
			MPG:       perGame(statMinutes),
			FGPercent: percentStat(stats, statFGPercent),
			FTPercent: percentStat(stats, statFTPercent),
			ThreePM:   perGame(statThreePM),
			Points:    perGame(statPoints),
			Rebounds:  perGame(statRebounds),
			Assists:   perGame(statAssists),
			Steals:    perGame(statSteals),
			Blocks:    perGame(statBlocks),
			Turnovers: perGame(statTurnovers),
		},
	}
}

func statNumber(stats map[int]string, id int) (float64, bool) {
	raw, ok := stats[id]
	if !ok {
		return 0, false
	}
	value, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil {
		return 0, false
	}
	return value, true
}

// Yahoo season stats are totals, so counting stats are divided by games played.
func divideStat(stats map[int]string, id int, games float64) string {
	value, ok := statNumber(stats, id)
	if !ok {
		return missingStat
	}
	return fmt.Sprintf("%.1f", value/games)
}

func percentStat(stats map[int]string, id int) string {
	value, ok := statNumber(stats, id)
	if !ok {
		return missingStat
	}
	if value <= 1 {
		value *= 100
	}
	return fmt.Sprintf("%.1f", value)
}
