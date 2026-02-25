package domain

import "time"

type SurgeryStatus string

const (
	SurgeryStatusScheduled SurgeryStatus = "SCHEDULED"
	SurgeryStatusCompleted SurgeryStatus = "COMPLETED"
	SurgeryStatusCancelled SurgeryStatus = "CANCELLED"
)

type Surgery struct {
	ID            uint          `gorm:"primaryKey" json:"id"`
	PatientID     uint          `gorm:"index;not null" json:"patient_id"`
	Patient       *Patient      `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
	SurgeonID     uint          `gorm:"index;not null" json:"surgeon_id"`
	Surgeon       *User         `gorm:"foreignKey:SurgeonID" json:"surgeon,omitempty"`
	ScheduledDate time.Time     `gorm:"not null" json:"scheduled_date"`
	OperationType OperationType `gorm:"type:varchar(30);not null" json:"operation_type"`
	Eye           string        `gorm:"type:varchar(5)" json:"eye"`
	Status        SurgeryStatus `gorm:"type:varchar(20);default:'SCHEDULED';not null" json:"status"`
	Notes         string        `gorm:"type:text" json:"notes"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

type CreateSurgeryRequest struct {
	PatientID     uint   `json:"patient_id" binding:"required"`
	ScheduledDate string `json:"scheduled_date" binding:"required"`
	Notes         string `json:"notes"`
}

type UpdateSurgeryRequest struct {
	ScheduledDate *string `json:"scheduled_date"`
	Status        *string `json:"status"`
	Notes         *string `json:"notes"`
}
