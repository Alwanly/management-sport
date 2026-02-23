package schema

import (
	"time"

	"github.com/Alwanly/management-sport/pkg/middleware"
)

type RequestAuditList struct {
	Page         int                      `form:"page" validate:"required,min=1"`
	PageSize     int                      `form:"page_size" validate:"required,min=1,max=100"`
	From         time.Time                `form:"from" validate:"omitempty"`
	To           time.Time                `form:"to" validate:"omitempty"`
	Entity       string                   `form:"entity" validate:"omitempty"`
	Action       string                   `form:"action" validate:"omitempty"`
	AuthUserData *middleware.AuthUserData `swaggerignore:"true"`
}

type RequestAuditGet struct {
	ID           string                   `uri:"id" validate:"required" swaggerignore:"true"`
	AuthUserData *middleware.AuthUserData `swaggerignore:"true"`
}
