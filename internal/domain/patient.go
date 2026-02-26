package domain

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

type PatientStatus string

const (
	PatientStatusDraft           PatientStatus = "DRAFT"
	PatientStatusInProgress      PatientStatus = "IN_PROGRESS"
	PatientStatusPendingReview   PatientStatus = "PENDING_REVIEW"
	PatientStatusApproved        PatientStatus = "APPROVED"
	PatientStatusNeedsCorrection PatientStatus = "NEEDS_CORRECTION"
	PatientStatusScheduled       PatientStatus = "SCHEDULED"
	PatientStatusCompleted       PatientStatus = "COMPLETED"
	PatientStatusCancelled       PatientStatus = "CANCELLED"
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
	Status         PatientStatus `gorm:"type:varchar(30);default:'DRAFT';not null;index" json:"status"`
	DoctorID       uint          `gorm:"index;not null" json:"doctor_id"`
	Doctor         *User         `gorm:"foreignKey:DoctorID" json:"doctor,omitempty"`
	SurgeonID      *uint         `gorm:"index" json:"surgeon_id"`
	Surgeon        *User         `gorm:"foreignKey:SurgeonID" json:"surgeon,omitempty"`
	DistrictID     uint          `gorm:"index" json:"district_id"`
	District       *District     `gorm:"foreignKey:DistrictID" json:"district,omitempty"`
	Notes          string        `gorm:"type:text" json:"notes"`
	SurgeryDate    *time.Time    `json:"surgery_date"`

	// Medical standards and integrations
	MedicalMetadata *MedicalStandardsMetadata `gorm:"type:jsonb" json:"medical_metadata,omitempty"`
	OMSPolicy       string                    `gorm:"type:varchar(16)" json:"oms_policy,omitempty"`
	Gender          string                    `gorm:"type:varchar(10)" json:"gender,omitempty"` // male, female

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

// ValidateStatusTransition проверяет допустимость перехода между статусами
func ValidateStatusTransition(from, to PatientStatus) bool {
	// Можно отменить из любого статуса
	if to == PatientStatusCancelled {
		return true
	}

	// Определяем допустимые переходы
	validTransitions := map[PatientStatus][]PatientStatus{
		PatientStatusDraft:           {PatientStatusInProgress, PatientStatusCancelled},
		PatientStatusInProgress:      {PatientStatusPendingReview, PatientStatusCancelled},
		PatientStatusPendingReview:   {PatientStatusApproved, PatientStatusNeedsCorrection, PatientStatusCancelled},
		PatientStatusNeedsCorrection: {PatientStatusInProgress, PatientStatusCancelled},
		PatientStatusApproved:        {PatientStatusScheduled, PatientStatusCancelled},
		PatientStatusScheduled:       {PatientStatusCompleted, PatientStatusCancelled},
		PatientStatusCompleted:       {}, // Финальный статус
		PatientStatusCancelled:       {}, // Финальный статус
	}

	allowedStatuses, exists := validTransitions[from]
	if !exists {
		return false
	}

	for _, allowed := range allowedStatuses {
		if allowed == to {
			return true
		}
	}

	return false
}

// GetStatusDisplayName возвращает человекочитаемое название статуса
func GetStatusDisplayName(status PatientStatus) string {
	names := map[PatientStatus]string{
		PatientStatusDraft:           "Черновик",
		PatientStatusInProgress:      "В процессе подготовки",
		PatientStatusPendingReview:   "Ожидает проверки хирурга",
		PatientStatusApproved:        "Одобрено, готов к операции",
		PatientStatusNeedsCorrection: "Требуется доработка",
		PatientStatusScheduled:       "Операция запланирована",
		PatientStatusCompleted:       "Операция завершена",
		PatientStatusCancelled:       "Отменено",
	}
	if name, ok := names[status]; ok {
		return name
	}
	return string(status)
}

func GetOperationTypeDisplayName(opType OperationType) string {
	names := map[OperationType]string{
		OperationPhacoemulsification: "Факоэмульсификация катаракты",
		OperationAntiglaucoma:        "Антиглаукомная операция",
		OperationVitrectomy:          "Витрэктомия",
	}
	if name, ok := names[opType]; ok {
		return name
	}
	return string(opType)
}

func GetEyeDisplayName(eye string) string {
	names := map[string]string{
		"OD": "Правый глаз",
		"OS": "Левый глаз",
		"OU": "Оба глаза",
	}
	if name, ok := names[eye]; ok {
		return name
	}
	return eye
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

type BatchUpdateRequest struct {
	Patient   *UpdatePatientRequest  `json:"patient"`
	Status    *PatientStatusRequest  `json:"status"`
	Checklist []ChecklistItemUpdate  `json:"checklist"`
	Timestamp string                 `json:"timestamp"` // ISO8601 timestamp from client
}

type ChecklistItemUpdate struct {
	ID     uint    `json:"id" binding:"required"`
	Status *string `json:"status"`
	Result *string `json:"result"`
	Notes  *string `json:"notes"`
}

type BatchUpdateResponse struct {
	Success      bool                   `json:"success"`
	Patient      *Patient               `json:"patient,omitempty"`
	Conflicts    []string               `json:"conflicts,omitempty"`
	UpdatedItems int                    `json:"updated_items"`
	Message      string                 `json:"message"`
}

type PatientPublicResponse struct {
	AccessCode    string                 `json:"access_code"`
	FirstName     string                 `json:"first_name"`
	LastName      string                 `json:"last_name"`
	Status        PatientStatus          `json:"status"`
	SurgeryDate   *time.Time             `json:"surgery_date"`
	StatusHistory []PatientStatusHistory `json:"status_history"`
}
