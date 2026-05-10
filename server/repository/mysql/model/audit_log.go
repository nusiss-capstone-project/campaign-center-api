package model

import "time"

// AuditLog maps to table audit_logs.
type AuditLog struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement"`
	EntityType   string    `gorm:"column:entity_type;size:64"`
	EntityID     int64     `gorm:"column:entity_id"`
	Action       string    `gorm:"column:action;size:64"`
	OperatorName string    `gorm:"column:operator_name;size:64"`
	DetailJSON   string    `gorm:"column:detail_json;type:text"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
