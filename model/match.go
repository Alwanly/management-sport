package model

import (
	"time"
)

type MatchStatus string

const (
	MatchStatusScheduled MatchStatus = "scheduled"
	MatchStatusOngoing   MatchStatus = "ongoing"
	MatchStatusFinished  MatchStatus = "finished"
)

type Match struct {
	ID         string      `gorm:"primaryKey;column:id;type:varchar(36);not null"`
	MatchDate  time.Time   `gorm:"column:match_date;type:date;not null;index"`
	MatchTime  time.Time   `gorm:"column:match_time;type:timestamptz;not null"`
	HomeTeamID string      `gorm:"column:home_team_id;type:varchar(36);not null;index"`
	AwayTeamID string      `gorm:"column:away_team_id;type:varchar(36);not null;index"`
	HomeScore  int         `gorm:"column:home_score;type:integer;default:0"`
	AwayScore  int         `gorm:"column:away_score;type:integer;default:0"`
	Status     MatchStatus `gorm:"column:status;type:varchar(20);not null;default:'scheduled'"`
	CreatedAt  time.Time   `gorm:"column:created_at;type:timestamptz;not null"`
	CreatedBy  string      `gorm:"column:created_by;type:varchar(36);not null"`
	UpdatedAt  time.Time   `gorm:"column:updated_at;type:timestamptz;not null"`
	UpdatedBy  string      `gorm:"column:updated_by;type:varchar(36)"`

	// Relationships
	HomeTeam *Team `gorm:"foreignKey:HomeTeamID;references:ID"`
	AwayTeam *Team `gorm:"foreignKey:AwayTeamID;references:ID"`
}

func (Match) TableName() string {
	return "matches"
}

type Matches []Match
