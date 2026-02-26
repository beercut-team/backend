package domain

import "time"

type ChecklistItemStatus string

const (
	ChecklistStatusPending    ChecklistItemStatus = "PENDING"
	ChecklistStatusInProgress ChecklistItemStatus = "IN_PROGRESS"
	ChecklistStatusCompleted  ChecklistItemStatus = "COMPLETED"
	ChecklistStatusRejected   ChecklistItemStatus = "REJECTED"
	ChecklistStatusExpired    ChecklistItemStatus = "EXPIRED"
)

type ChecklistTemplate struct {
	ID            uint          `gorm:"primaryKey" json:"id"`
	OperationType OperationType `gorm:"type:varchar(30);not null;index" json:"operation_type"`
	Name          string        `gorm:"not null" json:"name"`
	Description   string        `gorm:"type:text" json:"description"`
	Category      string        `gorm:"type:varchar(50)" json:"category"`
	IsRequired    bool          `gorm:"default:true" json:"is_required"`
	ExpiresInDays int           `json:"expires_in_days"`
	SortOrder     int           `json:"sort_order"`
	CreatedAt     time.Time     `json:"created_at"`
}

type ChecklistItem struct {
	ID           uint                `gorm:"primaryKey" json:"id"`
	PatientID    uint                `gorm:"index;not null" json:"patient_id"`
	TemplateID   uint                `gorm:"index" json:"template_id"`
	Template     *ChecklistTemplate  `gorm:"foreignKey:TemplateID" json:"template,omitempty"`
	Name         string              `gorm:"not null" json:"name"`
	Description  string              `gorm:"type:text" json:"description"`
	Category     string              `gorm:"type:varchar(50)" json:"category"`
	IsRequired   bool                `gorm:"default:true" json:"is_required"`
	Status       ChecklistItemStatus `gorm:"type:varchar(20);default:'PENDING';not null;index" json:"status"`
	Result       string              `gorm:"type:text" json:"result"`
	Notes        string              `gorm:"type:text" json:"notes"`
	CompletedAt  *time.Time          `json:"completed_at"`
	CompletedBy  *uint               `json:"completed_by"`
	ReviewedBy   *uint               `json:"reviewed_by"`
	ReviewNote   string              `gorm:"type:text" json:"review_note"`
	ExpiresAt    *time.Time          `json:"expires_at"`
	MediaID      *uint               `json:"media_id"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
}

// --- Requests ---

type CreateChecklistItemRequest struct {
	PatientID     uint   `json:"patient_id" binding:"required"`
	Name          string `json:"name" binding:"required"`
	Description   string `json:"description"`
	Category      string `json:"category"`
	IsRequired    bool   `json:"is_required"`
	ExpiresInDays int    `json:"expires_in_days"`
}

type UpdateChecklistItemRequest struct {
	Status string  `json:"status"`
	Result *string `json:"result"`
	Notes  *string `json:"notes"`
}

type ReviewChecklistItemRequest struct {
	Status     string `json:"status" binding:"required"` // COMPLETED or REJECTED
	ReviewNote string `json:"review_note"`
}

// --- Template definitions ---

type TemplateDefinition struct {
	Name          string
	Description   string
	Category      string
	IsRequired    bool
	ExpiresInDays int
	SortOrder     int
}

func GetChecklistTemplates(opType OperationType) []TemplateDefinition {
	common := []TemplateDefinition{
		{Name: "Общий анализ крови", Description: "Клинический анализ крови", Category: "Анализы", IsRequired: true, ExpiresInDays: 14, SortOrder: 1},
		{Name: "Общий анализ мочи", Description: "Общий анализ мочи", Category: "Анализы", IsRequired: true, ExpiresInDays: 14, SortOrder: 2},
		{Name: "Биохимический анализ крови", Description: "Глюкоза, АЛТ, АСТ, билирубин, креатинин, мочевина", Category: "Анализы", IsRequired: true, ExpiresInDays: 14, SortOrder: 3},
		{Name: "Коагулограмма", Description: "МНО, АЧТВ, фибриноген", Category: "Анализы", IsRequired: true, ExpiresInDays: 14, SortOrder: 4},
		{Name: "Анализ на ВИЧ", Description: "Anti-HIV 1/2", Category: "Анализы", IsRequired: true, ExpiresInDays: 90, SortOrder: 5},
		{Name: "Анализ на гепатит B", Description: "HBsAg", Category: "Анализы", IsRequired: true, ExpiresInDays: 90, SortOrder: 6},
		{Name: "Анализ на гепатит C", Description: "Anti-HCV", Category: "Анализы", IsRequired: true, ExpiresInDays: 90, SortOrder: 7},
		{Name: "Анализ на сифилис", Description: "RW", Category: "Анализы", IsRequired: true, ExpiresInDays: 90, SortOrder: 8},
		{Name: "ЭКГ", Description: "Электрокардиограмма с заключением", Category: "Обследования", IsRequired: true, ExpiresInDays: 14, SortOrder: 9},
		{Name: "Флюорография", Description: "Или рентген грудной клетки", Category: "Обследования", IsRequired: true, ExpiresInDays: 365, SortOrder: 10},
		{Name: "Заключение терапевта", Description: "Заключение терапевта об отсутствии противопоказаний", Category: "Заключения", IsRequired: true, ExpiresInDays: 14, SortOrder: 11},
		{Name: "Заключение эндокринолога", Description: "При наличии сахарного диабета", Category: "Заключения", IsRequired: false, ExpiresInDays: 30, SortOrder: 12},
	}

	switch opType {
	case OperationPhacoemulsification:
		specific := []TemplateDefinition{
			{Name: "Биометрия глаза (IOL Master / A-scan)", Description: "Биометрические данные для расчёта ИОЛ", Category: "Офтальмология", IsRequired: true, ExpiresInDays: 90, SortOrder: 13},
			{Name: "Кератометрия", Description: "Данные кератометрии", Category: "Офтальмология", IsRequired: true, ExpiresInDays: 90, SortOrder: 14},
			{Name: "OCT макулярной зоны", Description: "Оптическая когерентная томография", Category: "Офтальмология", IsRequired: false, ExpiresInDays: 90, SortOrder: 15},
		}
		return append(common, specific...)

	case OperationAntiglaucoma:
		specific := []TemplateDefinition{
			{Name: "Периметрия", Description: "Поля зрения", Category: "Офтальмология", IsRequired: true, ExpiresInDays: 30, SortOrder: 13},
			{Name: "Тонометрия", Description: "Измерение ВГД в динамике", Category: "Офтальмология", IsRequired: true, ExpiresInDays: 14, SortOrder: 14},
			{Name: "OCT диска зрительного нерва", Description: "OCT ДЗН", Category: "Офтальмология", IsRequired: true, ExpiresInDays: 90, SortOrder: 15},
			{Name: "Гониоскопия", Description: "Исследование угла передней камеры", Category: "Офтальмология", IsRequired: true, ExpiresInDays: 90, SortOrder: 16},
		}
		return append(common, specific...)

	case OperationVitrectomy:
		specific := []TemplateDefinition{
			{Name: "B-скан", Description: "УЗИ глазного яблока", Category: "Офтальмология", IsRequired: true, ExpiresInDays: 30, SortOrder: 13},
			{Name: "OCT макулярной зоны", Description: "Оптическая когерентная томография", Category: "Офтальмология", IsRequired: true, ExpiresInDays: 30, SortOrder: 14},
			{Name: "ЭФИ", Description: "Электрофизиологическое исследование", Category: "Офтальмология", IsRequired: false, ExpiresInDays: 90, SortOrder: 15},
		}
		return append(common, specific...)
	}

	return common
}
