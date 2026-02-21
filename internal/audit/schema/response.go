package schema

import "time"

type ResponseAuditItem struct {
	ID         string    `json:"id"`
	EntityType string    `json:"entity_type"`
	EntityID   string    `json:"entity_id"`
	Action     string    `json:"action"`
	UserID     string    `json:"user_id"`
	UserRole   string    `json:"user_role"`
	CreatedAt  time.Time `json:"created_at"`
}

type ResponseAuditGet struct {
	ID         string    `json:"id"`
	EntityType string    `json:"entity_type"`
	EntityID   string    `json:"entity_id"`
	Action     string    `json:"action"`
	OldValue   string    `json:"old_value"`
	NewValue   string    `json:"new_value"`
	UserID     string    `json:"user_id"`
	UserRole   string    `json:"user_role"`
	CreatedAt  time.Time `json:"created_at"`
}
