package model

import (
	"time"
)

type PlayerPosition string

const (
	PositionGoalkeeper PlayerPosition = "GK"
	PositionDefender   PlayerPosition = "DF"
	PositionMidfielder PlayerPosition = "MF"
	PositionForward    PlayerPosition = "FW"
)

type Player struct {
	ID           string         `gorm:"primaryKey;column:id;type:varchar(36);not null"`
	TeamID       string         `gorm:"column:team_id;type:varchar(36);not null;index"`
	Name         string         `gorm:"column:name;type:varchar(255);not null"`
	HeightCM     int            `gorm:"column:height_cm;type:integer"`
	WeightKG     int            `gorm:"column:weight_kg;type:integer"`
	Position     PlayerPosition `gorm:"column:position;type:varchar(10);not null"`
	ShirtNumber  int            `gorm:"column:shirt_number;type:integer;not null"`
	CreatedAt    time.Time      `gorm:"column:created_at;type:timestamptz;not null"`
	CreatedBy    string         `gorm:"column:created_by;type:varchar(36);not null"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;type:timestamptz;not null"`
	UpdatedBy    string         `gorm:"column:updated_by;type:varchar(36)"`
	DeletedAt    *time.Time     `gorm:"column:deleted_at;type:timestamptz;index"`

	// Relationships
	Team *Team `gorm:"foreignKey:TeamID;references:ID"`
}

func (Player) TableName() string {
	return "players"
}

type Players []Player
