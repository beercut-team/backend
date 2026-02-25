package domain

import "time"

type NotificationType string

const (
	NotifStatusChange     NotificationType = "STATUS_CHANGE"
	NotifNewComment       NotificationType = "NEW_COMMENT"
	NotifSurgeryScheduled NotificationType = "SURGERY_SCHEDULED"
	NotifChecklistExpiry  NotificationType = "CHECKLIST_EXPIRY"
	NotifSurgeryReminder  NotificationType = "SURGERY_REMINDER"
)

type Notification struct {
	ID        uint             `gorm:"primaryKey" json:"id"`
	UserID    uint             `gorm:"index;not null" json:"user_id"`
	Type      NotificationType `gorm:"type:varchar(30);not null" json:"type"`
	Title     string           `gorm:"not null" json:"title"`
	Body      string           `gorm:"type:text" json:"body"`
	EntityType string          `gorm:"type:varchar(30)" json:"entity_type"`
	EntityID  uint             `json:"entity_id"`
	IsRead    bool             `gorm:"default:false;index" json:"is_read"`
	CreatedAt time.Time        `json:"created_at"`
}

type CreateNotificationRequest struct {
	UserID     uint             `json:"user_id" binding:"required"`
	Type       NotificationType `json:"type" binding:"required"`
	Title      string           `json:"title" binding:"required"`
	Body       string           `json:"body"`
	EntityType string           `json:"entity_type"`
	EntityID   uint             `json:"entity_id"`
}
