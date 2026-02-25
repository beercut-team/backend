package domain

import "time"

type Comment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	PatientID uint      `gorm:"index;not null" json:"patient_id"`
	AuthorID  uint      `gorm:"not null" json:"author_id"`
	Author    *User     `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	ParentID  *uint     `gorm:"index" json:"parent_id"`
	Body      string    `gorm:"type:text;not null" json:"body"`
	IsUrgent  bool      `gorm:"default:false" json:"is_urgent"`
	IsRead    bool      `gorm:"default:false" json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateCommentRequest struct {
	PatientID uint   `json:"patient_id" binding:"required"`
	ParentID  *uint  `json:"parent_id"`
	Body      string `json:"body" binding:"required"`
	IsUrgent  bool   `json:"is_urgent"`
}
