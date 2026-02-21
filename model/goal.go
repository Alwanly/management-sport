package model

import (
	"time"
)

type Goal struct {
	ID           string    `gorm:"primaryKey;column:id;type:varchar(36);not null"`
	MatchID      string    `gorm:"column:match_id;type:varchar(36);not null;index"`
	PlayerID     string    `gorm:"column:player_id;type:varchar(36);not null;index"`
	MinuteScored int       `gorm:"column:minute_scored;type:integer;not null"`
	CreatedAt    time.Time `gorm:"column:created_at;type:timestamptz;not null"`

	// Relationships
	Match  *Match  `gorm:"foreignKey:MatchID;references:ID"`
	Player *Player `gorm:"foreignKey:PlayerID;references:ID"`
}

func (Goal) TableName() string {
	return "goals"
}

type Goals []Goal
