package yahoo

import (
	"errors"
	"os"
	"reflect"
	"testing"
)

func decodeTestdata(t *testing.T, name string) *Envelope {
	t.Helper()
	f, err := os.Open("testdata/" + name)
	if err != nil {
		t.Fatalf("open %s: %v", name, err)
	}
	t.Cleanup(func() { f.Close() })

	env, err := decodeResponse(f)
	if err != nil {
		t.Fatalf("decode %s: %v", name, err)
	}
	return env
}

func findStat(cats []StatCategory, id int) *StatCategory {
	for i := range cats {
		if cats[i].StatID == id {
			return &cats[i]
		}
	}
	return nil
}

func findModifier(mods []StatModifier, id int) *StatModifier {
	for i := range mods {
		if mods[i].StatID == id {
			return &mods[i]
		}
	}
	return nil
}

func findRosterPosition(positions []RosterPosition, name string) *RosterPosition {
	for i := range positions {
		if positions[i].Position == name {
			return &positions[i]
		}
	}
	return nil
}

func TestDecodeLeague(t *testing.T) {
	env := decodeTestdata(t, "league.xml")
	lg := env.League
	if lg == nil {
		t.Fatal("no league in response")
	}

	if lg.LeagueKey != "390.l.1000" {
		t.Errorf("LeagueKey = %q", lg.LeagueKey)
	}
	if lg.LeagueID != 1000 {
		t.Errorf("LeagueID = %d", lg.LeagueID)
	}
	if lg.Name != "Yahoo Public 1000" {
		t.Errorf("Name = %q", lg.Name)
	}
	if lg.DraftStatus != "postdraft" {
		t.Errorf("DraftStatus = %q", lg.DraftStatus)
	}
	if lg.NumTeams != 10 {
		t.Errorf("NumTeams = %d", lg.NumTeams)
	}
	if lg.ScoringType != "head" {
		t.Errorf("ScoringType = %q", lg.ScoringType)
	}
	if lg.LeagueType != "public" {
		t.Errorf("LeagueType = %q", lg.LeagueType)
	}
	if lg.CurrentWeek != 16 || lg.StartWeek != 1 || lg.EndWeek != 16 {
		t.Errorf("weeks = %d/%d/%d", lg.CurrentWeek, lg.StartWeek, lg.EndWeek)
	}
	if lg.StartDate != "2019-09-05" || lg.EndDate != "2019-12-23" {
		t.Errorf("dates = %q/%q", lg.StartDate, lg.EndDate)
	}
	if !lg.IsFinished {
		t.Error("IsFinished = false")
	}
	if lg.GameCode != "nfl" || lg.Season != 2019 {
		t.Errorf("game = %q/%d", lg.GameCode, lg.Season)
	}
	if lg.LeagueUpdateTimestamp != 1577869389 {
		t.Errorf("LeagueUpdateTimestamp = %d", lg.LeagueUpdateTimestamp)
	}
	if lg.WeeklyDeadline != "" {
		t.Errorf("WeeklyDeadline = %q, want empty", lg.WeeklyDeadline)
	}
	if lg.Settings != nil {
		t.Error("Settings should be nil for a metadata-only response")
	}
}

func TestDecodeLeagueSettings(t *testing.T) {
	env := decodeTestdata(t, "league_settings.xml")
	if env.League == nil || env.League.Settings == nil {
		t.Fatal("no league settings in response")
	}
	s := env.League.Settings

	if s.DraftType != "live" || s.ScoringType != "head" {
		t.Errorf("draft/scoring = %q/%q", s.DraftType, s.ScoringType)
	}
	if s.IsAuctionDraft {
		t.Error("IsAuctionDraft = true")
	}
	if !s.UsesPlayoff || s.PlayoffStartWeek != 15 || s.NumPlayoffTeams != 4 {
		t.Errorf("playoff = %v/%d/%d", s.UsesPlayoff, s.PlayoffStartWeek, s.NumPlayoffTeams)
	}
	if s.WaiverType != "R" || s.WaiverRule != "gametime" || s.UsesFAAB {
		t.Errorf("waivers = %q/%q/%v", s.WaiverType, s.WaiverRule, s.UsesFAAB)
	}
	if s.MaxTeams != 10 || s.WaiverTime != 2 {
		t.Errorf("teams/waiver time = %d/%d", s.MaxTeams, s.WaiverTime)
	}
	if s.TradeEndDate != "2019-11-09" || s.TradeRatifyType != "vote" {
		t.Errorf("trades = %q/%q", s.TradeEndDate, s.TradeRatifyType)
	}
	if s.PlayerPool != "ALL" || s.CantCutList != "yahoo" {
		t.Errorf("pool/cant cut = %q/%q", s.PlayerPool, s.CantCutList)
	}
	if !s.PickemEnabled || !s.UsesFractionalPoints || !s.UsesNegativePoints {
		t.Errorf("point flags = %v/%v/%v", s.PickemEnabled, s.UsesFractionalPoints, s.UsesNegativePoints)
	}

	if len(s.RosterPositions) != 9 {
		t.Fatalf("RosterPositions = %d, want 9", len(s.RosterPositions))
	}
	if rp := findRosterPosition(s.RosterPositions, "W/R/T"); rp == nil || rp.Count != 1 || rp.PositionType != "O" {
		t.Errorf("W/R/T roster position = %+v", rp)
	}

	cat := findStat(s.StatCategories, 9)
	if cat == nil {
		t.Fatal("stat category 9 missing")
	}
	if cat.Name != "Rushing Yards" || cat.DisplayName != "Rush Yds" || !cat.Enabled {
		t.Errorf("stat 9 = %+v", cat)
	}
	if !reflect.DeepEqual(cat.PositionTypes, []string{"O"}) {
		t.Errorf("stat 9 position types = %v", cat.PositionTypes)
	}

	if mod := findModifier(s.StatModifiers, 4); mod == nil || mod.Value != 0.04 {
		t.Errorf("modifier 4 = %+v", mod)
	}
	if mod := findModifier(s.StatModifiers, 9); mod == nil || mod.Value != 0.1 {
		t.Errorf("modifier 9 = %+v", mod)
	}
}

func TestDecodePlayer(t *testing.T) {
	env := decodeTestdata(t, "player_league.xml")
	if env.League == nil || env.League.Players == nil {
		t.Fatal("no players in response")
	}
	if env.League.Players.Count != 1 || len(env.League.Players.Players) != 1 {
		t.Fatalf("players = %d/%d", env.League.Players.Count, len(env.League.Players.Players))
	}

	p := env.League.Players.Players[0]
	if p.PlayerKey != "390.p.30121" || p.PlayerID != 30121 {
		t.Errorf("key/id = %q/%d", p.PlayerKey, p.PlayerID)
	}
	if p.Name.Full != "Christian McCaffrey" || p.Name.ASCIIFirst != "Christian" {
		t.Errorf("name = %+v", p.Name)
	}
	if p.EditorialTeamAbbr != "Car" || p.EditorialTeamFullName != "Carolina Panthers" {
		t.Errorf("team = %q/%q", p.EditorialTeamAbbr, p.EditorialTeamFullName)
	}
	if !reflect.DeepEqual(p.ByeWeeks, []int{7}) {
		t.Errorf("ByeWeeks = %v", p.ByeWeeks)
	}
	if p.UniformNumber != 22 || p.DisplayPosition != "RB" {
		t.Errorf("number/position = %d/%q", p.UniformNumber, p.DisplayPosition)
	}
	if !reflect.DeepEqual(p.EligiblePositions, []string{"RB", "W/R/T"}) {
		t.Errorf("EligiblePositions = %v", p.EligiblePositions)
	}
	if p.PositionType != "O" || p.PrimaryPosition != "RB" {
		t.Errorf("position types = %q/%q", p.PositionType, p.PrimaryPosition)
	}
	if p.IsUndroppable {
		t.Error("IsUndroppable = true")
	}
	if p.Headshot.Size != "small" || p.Headshot.URL == "" {
		t.Errorf("headshot = %+v", p.Headshot)
	}
	if p.Ownership != nil || p.PercentOwned != nil {
		t.Error("sub-resources should be nil when not requested")
	}
}

func TestDecodePlayerStats(t *testing.T) {
	env := decodeTestdata(t, "player_stats.xml")
	p := env.League.Players.Players[0]

	if p.PlayerStats == nil {
		t.Fatal("PlayerStats missing")
	}
	if p.PlayerStats.CoverageType != "season" || p.PlayerStats.Season != 2019 {
		t.Errorf("coverage = %q/%d", p.PlayerStats.CoverageType, p.PlayerStats.Season)
	}
	var rushing string
	for _, st := range p.PlayerStats.Stats {
		if st.StatID == 9 {
			rushing = st.Value
		}
	}
	if rushing != "1387" {
		t.Errorf("stat 9 = %q, want 1387", rushing)
	}
	if p.PlayerPoints == nil || p.PlayerPoints.Total != 413.2 {
		t.Errorf("points = %+v", p.PlayerPoints)
	}
}

func TestDecodeLeaguePlayers(t *testing.T) {
	env := decodeTestdata(t, "league_players.xml")
	if env.League == nil || env.League.Players == nil {
		t.Fatal("no players in response")
	}
	ps := env.League.Players
	if ps.Count != 2 || len(ps.Players) != 2 {
		t.Fatalf("players = %d/%d", ps.Count, len(ps.Players))
	}

	free := ps.Players[0]
	if free.Status != "Q" || free.StatusFull != "Questionable" || free.InjuryNote != "Hamstring" {
		t.Errorf("injury = %q/%q/%q", free.Status, free.StatusFull, free.InjuryNote)
	}
	if free.Ownership == nil || free.Ownership.OwnershipType != "freeagent" {
		t.Errorf("ownership = %+v", free.Ownership)
	}
	if free.Ownership.OwnerTeamKey != "" {
		t.Errorf("free agent owner = %q", free.Ownership.OwnerTeamKey)
	}
	if free.PercentOwned == nil || free.PercentOwned.Value != 99 {
		t.Errorf("percent owned = %+v", free.PercentOwned)
	}
	if free.DraftAnalysis == nil || free.DraftAnalysis.AveragePick != "1.2" {
		t.Errorf("draft analysis = %+v", free.DraftAnalysis)
	}

	owned := ps.Players[1]
	if owned.Ownership == nil || owned.Ownership.OwnershipType != "team" {
		t.Fatalf("ownership = %+v", owned.Ownership)
	}
	if owned.Ownership.OwnerTeamKey != "470.l.1000.t.3" || owned.Ownership.OwnerTeamName != "Team Curtain" {
		t.Errorf("owner = %q/%q", owned.Ownership.OwnerTeamKey, owned.Ownership.OwnerTeamName)
	}
	if !reflect.DeepEqual(owned.ByeWeeks, []int{5, 6}) {
		t.Errorf("multiple bye weeks = %v", owned.ByeWeeks)
	}
	if owned.PercentOwned.Delta != -2 {
		t.Errorf("delta = %d", owned.PercentOwned.Delta)
	}
}

func TestDecodeErrorDocument(t *testing.T) {
	f, err := os.Open("testdata/error.xml")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	_, err = decodeResponse(f)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v, want *APIError", err)
	}
	if apiErr.Description != "You must be logged in to view this league." {
		t.Errorf("description = %q", apiErr.Description)
	}
}
