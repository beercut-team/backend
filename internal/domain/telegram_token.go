package domain

import "time"

// TelegramLoginToken - одноразовый токен для входа через Telegram
type TelegramLoginToken struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Token     string    `gorm:"uniqueIndex;not null" json:"token"`
	PatientID uint      `gorm:"not null;index" json:"patient_id"`
	Used      bool      `gorm:"default:false;not null" json:"used"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}
