package model

import (
	"time"
)

type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
)

type User struct {
	ID        string    `gorm:"primaryKey;column:id;type:varchar(36);not null"`
	Username  string    `gorm:"column:username;type:varchar(255);not null;uniqueIndex"`
	Password  string    `gorm:"column:password;type:varchar(255);not null"`
	Role      UserRole  `gorm:"column:role;type:varchar(50);not null;default:'user'"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamptz;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamptz;not null"`
}

func (User) TableName() string {
	return "users"
}

type Users []User
