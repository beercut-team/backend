package domain

import "time"

type SyncQueue struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"index;not null" json:"user_id"`
	Entity     string    `gorm:"type:varchar(50);not null;index" json:"entity"`
	EntityID   uint      `gorm:"not null" json:"entity_id"`
	Action     string    `gorm:"type:varchar(20);not null" json:"action"` // CREATE, UPDATE, DELETE
	Payload    string    `gorm:"type:text" json:"payload"`
	ClientTime time.Time `json:"client_time"`
	ServerTime time.Time `gorm:"autoCreateTime" json:"server_time"`
	Synced     bool      `gorm:"default:false;index" json:"synced"`
}

type SyncPushRequest struct {
	Mutations []SyncMutation `json:"mutations" binding:"required"`
}

type SyncMutation struct {
	Entity     string      `json:"entity" binding:"required"`
	EntityID   uint        `json:"entity_id"`
	Action     string      `json:"action" binding:"required"`
	Payload    interface{} `json:"payload"`
	ClientTime string      `json:"client_time" binding:"required"`
}

type SyncPullResponse struct {
	Changes []SyncQueue `json:"changes"`
	Since   string      `json:"since"`
}
