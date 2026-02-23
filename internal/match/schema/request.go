package schema

import (
	"github.com/Alwanly/management-sport/pkg/middleware"
)

type RequestMatchCreate struct {
	MatchDate    string                   `json:"match_date" validate:"required,datetime=2006-01-02"`
	MatchTime    string                   `json:"match_time" validate:"required,datetime=15:04"`
	HomeTeamID   string                   `json:"home_team_id" validate:"required"`
	AwayTeamID   string                   `json:"away_team_id" validate:"required,nefield=HomeTeamID"`
	AuthUserData *middleware.AuthUserData `swaggerignore:"true"`
}

type RequestMatchGet struct {
	ID           string                   `uri:"id" validate:"required" swaggerignore:"true"`
	AuthUserData *middleware.AuthUserData `swaggerignore:"true"`
}

type RequestMatchList struct {
	Page         int                      `form:"page" validate:"required,min=1"`
	PageSize     int                      `form:"page_size" validate:"required,min=1,max=100"`
	SortBy       string                   `form:"sort_by" validate:"omitempty,oneof=match_date status"`
	SortOrder    string                   `form:"sort_order" validate:"omitempty,oneof=asc desc"`
	AuthUserData *middleware.AuthUserData `swaggerignore:"true"`
}

type RequestMatchUpdate struct {
	ID           string                   `uri:"id" validate:"required" swaggerignore:"true"`
	MatchDate    string                   `json:"match_date" validate:"required,datetime=2006-01-02"`
	MatchTime    string                   `json:"match_time" validate:"required,datetime=15:04"`
	HomeTeamID   string                   `json:"home_team_id" validate:"required"`
	AwayTeamID   string                   `json:"away_team_id" validate:"required,nefield=HomeTeamID"`
	HomeScore    int                      `json:"home_score" validate:"omitempty,min=0"`
	AwayScore    int                      `json:"away_score" validate:"omitempty,min=0"`
	Status       string                   `json:"status" validate:"omitempty,oneof=scheduled ongoing finished"`
	AuthUserData *middleware.AuthUserData `swaggerignore:"true"`
}

type RequestMatchDelete struct {
	ID           string                   `uri:"id" validate:"required" swaggerignore:"true"`
	AuthUserData *middleware.AuthUserData `swaggerignore:"true"`
}

type RequestMatchUpdateStatus struct {
	ID           string                   `uri:"id" validate:"required" swaggerignore:"true"`
	Status       string                   `json:"status" validate:"required,oneof=scheduled ongoing finished"`
	AuthUserData *middleware.AuthUserData `swaggerignore:"true"`
}
