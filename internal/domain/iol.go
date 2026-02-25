package domain

import "time"

type IOLCalculation struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	PatientID     uint      `gorm:"index;not null" json:"patient_id"`
	Eye           string    `gorm:"type:varchar(5);not null" json:"eye"`
	AxialLength   float64   `gorm:"not null" json:"axial_length"`
	Keratometry1  float64   `gorm:"not null" json:"keratometry1"`
	Keratometry2  float64   `gorm:"not null" json:"keratometry2"`
	ACD           float64   `json:"acd"`
	TargetRefraction float64 `json:"target_refraction"`
	Formula       string    `gorm:"type:varchar(20);not null" json:"formula"`
	IOLPower      float64   `json:"iol_power"`
	PredictedRefraction float64 `json:"predicted_refraction"`
	AConstant     float64   `json:"a_constant"`
	CalculatedBy  uint      `json:"calculated_by"`
	Warnings      string    `gorm:"type:text" json:"warnings"`
	CreatedAt     time.Time `json:"created_at"`
}

type IOLCalculationRequest struct {
	PatientID      uint    `json:"patient_id" binding:"required"`
	Eye            string  `json:"eye" binding:"required"`
	AxialLength    float64 `json:"axial_length" binding:"required"`
	Keratometry1   float64 `json:"keratometry1" binding:"required"`
	Keratometry2   float64 `json:"keratometry2" binding:"required"`
	ACD            float64 `json:"acd"`
	TargetRefraction float64 `json:"target_refraction"`
	Formula        string  `json:"formula" binding:"required"`
	AConstant      float64 `json:"a_constant"`
}
