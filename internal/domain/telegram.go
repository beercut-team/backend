package domain

import "time"

type TelegramBinding struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	PatientID  uint      `gorm:"index;not null" json:"patient_id"`
	ChatID     int64     `gorm:"uniqueIndex;not null" json:"chat_id"`
	AccessCode string    `gorm:"index;not null" json:"access_code"`
	IsActive   bool      `gorm:"default:true" json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
}
