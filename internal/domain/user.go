package domain

import "time"

type Role string

const (
	RoleDistrictDoctor Role = "DISTRICT_DOCTOR"
	RoleSurgeon        Role = "SURGEON"
	RolePatient        Role = "PATIENT"
	RoleAdmin          Role = "ADMIN"
	RoleCallCenter     Role = "CALL_CENTER"
)

type User struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Email          string    `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash   string    `gorm:"not null" json:"-"`
	Name           string    `gorm:"not null" json:"name"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	MiddleName     string    `json:"middle_name"`
	Phone          string    `gorm:"index" json:"phone"`
	Role           Role      `gorm:"type:varchar(20);default:'PATIENT';not null;index" json:"role"`
	DistrictID     *uint     `gorm:"index" json:"district_id"`
	District       *District `gorm:"foreignKey:DistrictID" json:"district,omitempty"`
	Specialization string    `json:"specialization"`
	LicenseNumber  string    `json:"license_number"`
	TelegramChatID *int64    `gorm:"index" json:"telegram_chat_id,omitempty"`
	IsActive       bool      `gorm:"default:true;not null" json:"is_active"`
	RefreshToken   string    `gorm:"index" json:"-"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// --- Requests ---

type RegisterRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=6"`
	Name       string `json:"name" binding:"required,min=2"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	MiddleName string `json:"middle_name"`
	Phone      string `json:"phone"`
	Role       Role   `json:"role"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type PatientLoginRequest struct {
	AccessCode string `json:"access_code" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// --- Responses ---

type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         UserResponse `json:"user"`
}

type UserResponse struct {
	ID             uint   `json:"id"`
	Email          string `json:"email"`
	Name           string `json:"name"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	MiddleName     string `json:"middle_name"`
	Phone          string `json:"phone"`
	Role           Role   `json:"role"`
	DistrictID     *uint  `json:"district_id"`
	Specialization string `json:"specialization"`
	LicenseNumber  string `json:"license_number"`
	IsActive       bool   `json:"is_active"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:             u.ID,
		Email:          u.Email,
		Name:           u.Name,
		FirstName:      u.FirstName,
		LastName:       u.LastName,
		MiddleName:     u.MiddleName,
		Phone:          u.Phone,
		Role:           u.Role,
		DistrictID:     u.DistrictID,
		Specialization: u.Specialization,
		LicenseNumber:  u.LicenseNumber,
		IsActive:       u.IsActive,
	}
}

func ValidRole(r Role) bool {
	switch r {
	case RoleDistrictDoctor, RoleSurgeon, RolePatient, RoleAdmin, RoleCallCenter:
		return true
	}
	return false
}
