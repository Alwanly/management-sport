package schema

import (
	"time"
)

type ResponseGoalsPerPlayerItem struct {
	PlayerID string `json:"player_id"`
	Goals    int64  `json:"goals"`
}

type ResponseTeamGoalsItem struct {
	TeamID string `json:"team_id"`
	Goals  int64  `json:"goals"`
}

type ResponseMatchReportItem struct {
	MatchID    string         `json:"match_id"`
	MatchDate  time.Time      `json:"match_date"`
	MatchTime  time.Time      `json:"match_time"`
	HomeTeam   MatchTeamInfo  `json:"home_team"`
	AwayTeam   MatchTeamInfo  `json:"away_team"`
	FinalScore MatchScoreInfo `json:"final_score"`
	Result     string         `json:"result"`
	Status     string         `json:"status"`
}

type MatchTeamInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	LogoURL string `json:"logo_url"`
}

type MatchScoreInfo struct {
	Home int `json:"home"`
	Away int `json:"away"`
}

type ResponseMatchScorer struct {
	PlayerID     string `json:"player_id"`
	PlayerName   string `json:"player_name"`
	TeamName     string `json:"team_name"`
	Position     string `json:"position"`
	ShirtNumber  int    `json:"shirt_number"`
	MinuteScored int    `json:"minute_scored"`
}

type ResponseTeamStatistics struct {
	TeamID             string `json:"team_id"`
	TeamName           string `json:"team_name"`
	CumulativeHomeWins int64  `json:"cumulative_home_wins"`
	CumulativeAwayWins int64  `json:"cumulative_away_wins"`
}

type ResponseEnhancedMatchReport struct {
	Matches        []ResponseMatchReportItem         `json:"matches"`
	Scorers        map[string][]ResponseMatchScorer  `json:"scorers"`
	TeamStatistics map[string]ResponseTeamStatistics `json:"team_statistics"`
}
