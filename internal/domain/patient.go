package domain

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

type PatientStatus string

const (
	PatientStatusNew              PatientStatus = "NEW"
	PatientStatusPreparation      PatientStatus = "PREPARATION"
	PatientStatusReviewNeeded     PatientStatus = "REVIEW_NEEDED"
	PatientStatusApproved         PatientStatus = "APPROVED"
	PatientStatusSurgeryScheduled PatientStatus = "SURGERY_SCHEDULED"
	PatientStatusCompleted        PatientStatus = "COMPLETED"
	PatientStatusRejected         PatientStatus = "REJECTED"
)

type OperationType string

const (
	OperationPhacoemulsification OperationType = "PHACOEMULSIFICATION"
	OperationAntiglaucoma        OperationType = "ANTIGLAUCOMA"
	OperationVitrectomy           OperationType = "VITRECTOMY"
)

type Patient struct {
	ID             uint          `gorm:"primaryKey" json:"id"`
	AccessCode     string        `gorm:"uniqueIndex;not null" json:"access_code"`
	FirstName      string        `gorm:"not null" json:"first_name"`
	LastName       string        `gorm:"not null" json:"last_name"`
	MiddleName     string        `json:"middle_name"`
	DateOfBirth    time.Time     `json:"date_of_birth"`
	Phone          string        `json:"phone"`
	Email          string        `json:"email"`
	Address        string        `json:"address"`
	SNILs          string        `json:"snils"`
	PassportSeries string        `json:"passport_series"`
	PassportNumber string        `json:"passport_number"`
	PolicyNumber   string        `json:"policy_number"`
	Diagnosis      string        `gorm:"type:text" json:"diagnosis"`
	OperationType  OperationType `gorm:"type:varchar(30);not null" json:"operation_type"`
	Eye            string        `gorm:"type:varchar(5)" json:"eye"` // OD, OS, OU
	Status         PatientStatus `gorm:"type:varchar(30);default:'NEW';not null;index" json:"status"`
	DoctorID       uint          `gorm:"index;not null" json:"doctor_id"`
	Doctor         *User         `gorm:"foreignKey:DoctorID" json:"doctor,omitempty"`
	SurgeonID      *uint         `gorm:"index" json:"surgeon_id"`
	Surgeon        *User         `gorm:"foreignKey:SurgeonID" json:"surgeon,omitempty"`
	DistrictID     uint          `gorm:"index" json:"district_id"`
	District       *District     `gorm:"foreignKey:DistrictID" json:"district,omitempty"`
	Notes          string        `gorm:"type:text" json:"notes"`
	SurgeryDate    *time.Time    `json:"surgery_date"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

type PatientStatusHistory struct {
	ID        uint          `gorm:"primaryKey" json:"id"`
	PatientID uint          `gorm:"index;not null" json:"patient_id"`
	FromStatus PatientStatus `gorm:"type:varchar(30)" json:"from_status"`
	ToStatus  PatientStatus `gorm:"type:varchar(30);not null" json:"to_status"`
	ChangedBy uint          `json:"changed_by"`
	Comment   string        `gorm:"type:text" json:"comment"`
	CreatedAt time.Time     `json:"created_at"`
}

func GenerateAccessCode() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// --- Requests ---

type CreatePatientRequest struct {
	FirstName      string        `json:"first_name" binding:"required"`
	LastName       string        `json:"last_name" binding:"required"`
	MiddleName     string        `json:"middle_name"`
	DateOfBirth    string        `json:"date_of_birth"`
	Phone          string        `json:"phone"`
	Email          string        `json:"email"`
	Address        string        `json:"address"`
	SNILs          string        `json:"snils"`
	PassportSeries string        `json:"passport_series"`
	PassportNumber string        `json:"passport_number"`
	PolicyNumber   string        `json:"policy_number"`
	Diagnosis      string        `json:"diagnosis"`
	OperationType  OperationType `json:"operation_type" binding:"required"`
	Eye            string        `json:"eye" binding:"required"`
	DistrictID     uint          `json:"district_id" binding:"required"`
	Notes          string        `json:"notes"`
}

type UpdatePatientRequest struct {
	FirstName      *string `json:"first_name"`
	LastName       *string `json:"last_name"`
	MiddleName     *string `json:"middle_name"`
	Phone          *string `json:"phone"`
	Email          *string `json:"email"`
	Address        *string `json:"address"`
	Diagnosis      *string `json:"diagnosis"`
	Notes          *string `json:"notes"`
	SNILs          *string `json:"snils"`
	PassportSeries *string `json:"passport_series"`
	PassportNumber *string `json:"passport_number"`
	PolicyNumber   *string `json:"policy_number"`
}

type PatientStatusRequest struct {
	Status  PatientStatus `json:"status" binding:"required"`
	Comment string        `json:"comment"`
}

type PatientPublicResponse struct {
	AccessCode    string                 `json:"access_code"`
	FirstName     string                 `json:"first_name"`
	LastName      string                 `json:"last_name"`
	Status        PatientStatus          `json:"status"`
	SurgeryDate   *time.Time             `json:"surgery_date"`
	StatusHistory []PatientStatusHistory `json:"status_history"`
}
