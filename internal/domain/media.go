package domain

import "time"

type Media struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	PatientID    uint      `gorm:"index;not null" json:"patient_id"`
	UploadedBy   uint      `json:"uploaded_by"`
	FileName     string    `gorm:"not null" json:"file_name"`
	OriginalName string    `gorm:"not null" json:"original_name"`
	ContentType  string    `gorm:"not null" json:"content_type"`
	Size         int64     `json:"size"`
	StoragePath  string    `gorm:"not null" json:"storage_path"`
	ThumbnailPath string   `json:"thumbnail_path"`
	Category     string    `gorm:"type:varchar(50);index" json:"category"`
	CreatedAt    time.Time `json:"created_at"`
}
