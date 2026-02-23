package schema

import (
	"github.com/Alwanly/management-sport/pkg/middleware"
)

type RequestGoalCreate struct {
	MatchID      string                   `json:"match_id" validate:"required"`
	PlayerID     string                   `json:"player_id" validate:"required"`
	MinuteScored int                      `json:"minute_scored" validate:"required,min=0"`
	AuthUserData *middleware.AuthUserData `swaggerignore:"true"`
}

type RequestGoalGet struct {
	ID           string                   `uri:"id" validate:"required" swaggerignore:"true"`
	AuthUserData *middleware.AuthUserData `swaggerignore:"true"`
}

type RequestGoalList struct {
	Page         int                      `form:"page" validate:"required,min=1"`
	PageSize     int                      `form:"page_size" validate:"required,min=1,max=100"`
	SortBy       string                   `form:"sort_by" validate:"omitempty,oneof=minute_scored created_at"`
	SortOrder    string                   `form:"sort_order" validate:"omitempty,oneof=asc desc"`
	AuthUserData *middleware.AuthUserData `swaggerignore:"true"`
}

type RequestGoalDelete struct {
	ID           string                   `uri:"id" validate:"required" swaggerignore:"true"`
	AuthUserData *middleware.AuthUserData `swaggerignore:"true"`
}
