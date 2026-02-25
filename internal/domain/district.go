package domain

import "time"

type District struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"not null;uniqueIndex" json:"name"`
	Region    string    `gorm:"not null" json:"region"`
	Code      string    `gorm:"uniqueIndex" json:"code"`
	Timezone  string    `gorm:"default:'Europe/Moscow'" json:"timezone"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateDistrictRequest struct {
	Name     string `json:"name" binding:"required"`
	Region   string `json:"region" binding:"required"`
	Code     string `json:"code" binding:"required"`
	Timezone string `json:"timezone"`
}

type UpdateDistrictRequest struct {
	Name     *string `json:"name"`
	Region   *string `json:"region"`
	Code     *string `json:"code"`
	Timezone *string `json:"timezone"`
}
