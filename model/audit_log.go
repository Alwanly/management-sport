package model

import (
	"time"
)

type AuditAction string

const (
	AuditActionCreate AuditAction = "create"
	AuditActionUpdate AuditAction = "update"
	AuditActionDelete AuditAction = "delete"
)

type AuditLog struct {
	ID         string      `gorm:"primaryKey;column:id;type:varchar(36);not null"`
	EntityType string      `gorm:"column:entity_type;type:varchar(100);not null;index"`
	EntityID   string      `gorm:"column:entity_id;type:varchar(36);not null;index"`
	Action     AuditAction `gorm:"column:action;type:varchar(20);not null"`
	OldValue   string      `gorm:"column:old_value;type:jsonb"`
	NewValue   string      `gorm:"column:new_value;type:jsonb"`
	UserID     string      `gorm:"column:user_id;type:varchar(36);not null"`
	UserRole   string      `gorm:"column:user_role;type:varchar(50)"`
	CreatedAt  time.Time   `gorm:"column:created_at;type:timestamptz;not null;index"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

type AuditLogs []AuditLog
