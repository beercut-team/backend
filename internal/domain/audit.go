package domain

import "time"

type AuditLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"index" json:"user_id"`
	Action     string    `gorm:"not null;index" json:"action"`
	Entity     string    `gorm:"not null;index" json:"entity"`
	EntityID   uint      `json:"entity_id"`
	OldValue   string    `gorm:"type:text" json:"old_value,omitempty"`
	NewValue   string    `gorm:"type:text" json:"new_value,omitempty"`
	IP         string    `json:"ip"`
	CreatedAt  time.Time `json:"created_at"`
}
