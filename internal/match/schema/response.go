package schema

import "time"

type ResponseMatchCreate struct {
	ID string `json:"id"`
}

type ResponseMatchGet struct {
	ID         string    `json:"id"`
	MatchDate  time.Time `json:"match_date"`
	MatchTime  time.Time `json:"match_time"`
	HomeTeamID string    `json:"home_team_id"`
	AwayTeamID string    `json:"away_team_id"`
	HomeScore  int       `json:"home_score"`
	AwayScore  int       `json:"away_score"`
	Status     string    `json:"status"`
}

type ResponseMatchItem struct {
	ID         string    `json:"id"`
	MatchDate  time.Time `json:"match_date"`
	HomeTeamID string    `json:"home_team_id"`
	AwayTeamID string    `json:"away_team_id"`
	Status     string    `json:"status"`
}

type ResponseMatchUpdate struct {
	ID string `json:"id"`
}

type ResponseMatchDelete struct{}
