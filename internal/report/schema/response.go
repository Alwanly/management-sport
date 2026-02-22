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
