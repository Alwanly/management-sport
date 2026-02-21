package model

import (
	"time"
)

type Team struct {
	ID          string     `gorm:"primaryKey;column:id;type:varchar(36);not null"`
	Name        string     `gorm:"column:name;type:varchar(255);not null"`
	LogoURL     string     `gorm:"column:logo_url;type:text"`
	FoundedYear int        `gorm:"column:founded_year;type:integer"`
	Address     string     `gorm:"column:address;type:text"`
	City        string     `gorm:"column:city;type:varchar(255)"`
	CreatedAt   time.Time  `gorm:"column:created_at;type:timestamptz;not null"`
	CreatedBy   string     `gorm:"column:created_by;type:varchar(36);not null"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;type:timestamptz;not null"`
	UpdatedBy   string     `gorm:"column:updated_by;type:varchar(36)"`
	DeletedAt   *time.Time `gorm:"column:deleted_at;type:timestamptz;index"`
}

func (Team) TableName() string {
	return "teams"
}

type Teams []Team
