package schema

import (
	"time"

	"github.com/Alwanly/management-sport/pkg/middleware"
)

type RequestMatchCreate struct {
	MatchDate    time.Time `json:"match_date" validate:"required"`
	MatchTime    time.Time `json:"match_time" validate:"required"`
	HomeTeamID   string    `json:"home_team_id" validate:"required,uuid4"`
	AwayTeamID   string    `json:"away_team_id" validate:"required,uuid4,nefield=HomeTeamID"`
	AuthUserData *middleware.AuthUserData
}

type RequestMatchGet struct {
	ID           string `uri:"id" validate:"required"`
	AuthUserData *middleware.AuthUserData
}

type RequestMatchList struct {
	Page         int    `form:"page" validate:"required,min=1"`
	PageSize     int    `form:"page_size" validate:"required,min=1,max=100"`
	SortBy       string `form:"sort_by" validate:"omitempty,oneof=match_date status"`
	SortOrder    string `form:"sort_order" validate:"omitempty,oneof=asc desc"`
	AuthUserData *middleware.AuthUserData
}

type RequestMatchUpdate struct {
	ID           string    `uri:"id" validate:"required"`
	MatchDate    time.Time `json:"match_date" validate:"required"`
	MatchTime    time.Time `json:"match_time" validate:"required"`
	HomeTeamID   string    `json:"home_team_id" validate:"required,uuid4"`
	AwayTeamID   string    `json:"away_team_id" validate:"required,uuid4,nefield=HomeTeamID"`
	HomeScore    int       `json:"home_score" validate:"omitempty,min=0"`
	AwayScore    int       `json:"away_score" validate:"omitempty,min=0"`
	Status       string    `json:"status" validate:"omitempty,oneof=scheduled ongoing finished"`
	AuthUserData *middleware.AuthUserData
}

type RequestMatchDelete struct {
	ID           string `uri:"id" validate:"required"`
	AuthUserData *middleware.AuthUserData
}
