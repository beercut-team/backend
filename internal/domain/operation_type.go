package domain

import "time"

// OperationTypeModel - модель для управления типами операций в БД
type OperationTypeModel struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	Code             string    `gorm:"uniqueIndex;not null" json:"code"` // PHACOEMULSIFICATION, ANTIGLAUCOMA, VITRECTOMY
	Name             string    `gorm:"not null" json:"name"`
	Description      string    `gorm:"type:text" json:"description"`
	ChecklistTemplate string   `gorm:"type:text" json:"checklist_template"` // JSON array of checklist items
	IsActive         bool      `gorm:"default:true;not null" json:"is_active"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (OperationTypeModel) TableName() string {
	return "operation_types"
}

type CreateOperationTypeRequest struct {
	Code              string `json:"code" binding:"required"`
	Name              string `json:"name" binding:"required"`
	Description       string `json:"description"`
	ChecklistTemplate string `json:"checklist_template"`
}

type UpdateOperationTypeRequest struct {
	Name              *string `json:"name"`
	Description       *string `json:"description"`
	ChecklistTemplate *string `json:"checklist_template"`
	IsActive          *bool   `json:"is_active"`
}
