package schema

import (
	"github.com/Alwanly/management-sport/pkg/middleware"
)

type RequestPlayerCreate struct {
	TeamID       string                   `json:"team_id" validate:"required"`
	Name         string                   `json:"name" validate:"required,min=1,max=255"`
	HeightCM     int                      `json:"height_cm" validate:"omitempty,min=0"`
	WeightKG     int                      `json:"weight_kg" validate:"omitempty,min=0"`
	Position     string                   `json:"position" validate:"required,oneof=GK CB LB DM CM RM LM RW LW CF"`
	ShirtNumber  int                      `json:"shirt_number" validate:"required,min=1,max=99"`
	AuthUserData *middleware.AuthUserData `swaggerignore:"true"`
}

type RequestPlayerGet struct {
	ID           string                   `uri:"id" validate:"required" swaggerignore:"true"`
	AuthUserData *middleware.AuthUserData `swaggerignore:"true"`
}

type RequestPlayerList struct {
	Page         int                      `form:"page" validate:"required,min=1"`
	PageSize     int                      `form:"page_size" validate:"required,min=1,max=100"`
	SortBy       string                   `form:"sort_by" validate:"omitempty,oneof=name shirt_number position"`
	SortOrder    string                   `form:"sort_order" validate:"omitempty,oneof=asc desc"`
	AuthUserData *middleware.AuthUserData `swaggerignore:"true"`
}

type RequestPlayerUpdate struct {
	ID           string                   `uri:"id" validate:"required" swaggerignore:"true"`
	Name         string                   `json:"name" validate:"required,min=1,max=255"`
	HeightCM     int                      `json:"height_cm" validate:"omitempty,min=0"`
	WeightKG     int                      `json:"weight_kg" validate:"omitempty,min=0"`
	Position     string                   `json:"position" validate:"required,oneof=GK DF MF FW"`
	ShirtNumber  int                      `json:"shirt_number" validate:"required,min=1,max=99"`
	AuthUserData *middleware.AuthUserData `swaggerignore:"true"`
}

type RequestPlayerDelete struct {
	ID           string                   `uri:"id" validate:"required" swaggerignore:"true"`
	AuthUserData *middleware.AuthUserData `swaggerignore:"true"`
}

type RequestPlayerUpdateStatus struct {
	ID           string                   `uri:"id" validate:"required" swaggerignore:"true"`
	Status       string                   `json:"status" validate:"required,oneof=ongoing finished injured suspended"`
	AuthUserData *middleware.AuthUserData `swaggerignore:"true"`
}
