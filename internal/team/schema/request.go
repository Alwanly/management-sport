package schema

import (
	"github.com/Alwanly/management-sport/model"
	"github.com/Alwanly/management-sport/pkg/middleware"
)

type RequestTeamCreate struct {
	Name         string                   `form:"name" validate:"required,min=3,max=255"`
	FoundedYear  int                      `form:"founded_year" validate:"omitempty,min=1800,max=2100"`
	Address      string                   `form:"address" validate:"omitempty,max=500"`
	City         string                   `form:"city" validate:"omitempty,max=255"`
	AuthUserData *middleware.AuthUserData `swaggerignore:"true"`
}

type RequestTeamGet struct {
	ID           string                   `uri:"id" validate:"required"`
	AuthUserData *middleware.AuthUserData `swaggerignore:"true"`
}

type RequestTeamList struct {
	Page         int                      `form:"page" validate:"required,min=1"`
	PageSize     int                      `form:"page_size" validate:"required,min=1,max=100"`
	SortBy       string                   `form:"sort_by" validate:"omitempty,oneof=name city founded_year"`
	SortOrder    string                   `form:"sort_order" validate:"omitempty,oneof=asc desc"`
	AuthUserData *middleware.AuthUserData `swaggerignore:"true"`
}

type RequestTeamUpdate struct {
	ID           string                   `uri:"id" validate:"required" swaggerignore:"true"`
	Name         string                   `form:"name" validate:"required,min=3,max=255"`
	FoundedYear  int                      `form:"founded_year" validate:"omitempty,min=1800,max=2100"`
	Address      string                   `form:"address" validate:"omitempty,max=500"`
	City         string                   `form:"city" validate:"omitempty,max=255"`
	AuthUserData *middleware.AuthUserData `swaggerignore:"true"`
}

type RequestTeamDelete struct {
	ID           string                   `uri:"id" validate:"required"`
	AuthUserData *middleware.AuthUserData `swaggerignore:"true"`
}

func (r *RequestTeamList) ToResponse(teams []model.Team) []ResponseTeamItem {
	responseTeams := make([]ResponseTeamItem, len(teams))
	for i, team := range teams {
		responseTeams[i] = ResponseTeamItem{
			ID:          team.ID,
			Name:        team.Name,
			LogoURL:     team.LogoURL,
			FoundedYear: team.FoundedYear,
			City:        team.City,
		}
	}
	return responseTeams
}
