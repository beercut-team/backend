package repository

import (
	"context"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"gorm.io/gorm"
)

type PatientRepository interface {
	Create(ctx context.Context, patient *domain.Patient) error
	FindByID(ctx context.Context, id uint) (*domain.Patient, error)
	FindByAccessCode(ctx context.Context, code string) (*domain.Patient, error)
	FindAll(ctx context.Context, filters PatientFilters, offset, limit int) ([]domain.Patient, int64, error)
	Update(ctx context.Context, patient *domain.Patient) error
	UpdateStatus(ctx context.Context, id uint, status domain.PatientStatus) error
	CreateStatusHistory(ctx context.Context, h *domain.PatientStatusHistory) error
	FindStatusHistory(ctx context.Context, patientID uint) ([]domain.PatientStatusHistory, error)
	CountByStatus(ctx context.Context, doctorID *uint) (map[domain.PatientStatus]int64, error)
}

type PatientFilters struct {
	DoctorID   *uint
	SurgeonID  *uint
	DistrictID *uint
	Status     *domain.PatientStatus
	Search     string
	MinStatus  []domain.PatientStatus
}

type patientRepository struct {
	db *gorm.DB
}

func NewPatientRepository(db *gorm.DB) PatientRepository {
	return &patientRepository{db: db}
}

func (r *patientRepository) Create(ctx context.Context, patient *domain.Patient) error {
	return r.db.WithContext(ctx).Create(patient).Error
}

func (r *patientRepository) FindByID(ctx context.Context, id uint) (*domain.Patient, error) {
	var patient domain.Patient
	if err := r.db.WithContext(ctx).Preload("Doctor").Preload("Surgeon").Preload("District").First(&patient, id).Error; err != nil {
		return nil, err
	}
	return &patient, nil
}

func (r *patientRepository) FindByAccessCode(ctx context.Context, code string) (*domain.Patient, error) {
	var patient domain.Patient
	if err := r.db.WithContext(ctx).Where("access_code = ?", code).First(&patient).Error; err != nil {
		return nil, err
	}
	return &patient, nil
}

func (r *patientRepository) FindAll(ctx context.Context, filters PatientFilters, offset, limit int) ([]domain.Patient, int64, error) {
	var patients []domain.Patient
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Patient{})

	if filters.DoctorID != nil {
		query = query.Where("doctor_id = ?", *filters.DoctorID)
	}
	if filters.SurgeonID != nil {
		query = query.Where("surgeon_id = ?", *filters.SurgeonID)
	}
	if filters.DistrictID != nil {
		query = query.Where("district_id = ?", *filters.DistrictID)
	}
	if filters.Status != nil {
		query = query.Where("status = ?", *filters.Status)
	}
	if len(filters.MinStatus) > 0 {
		query = query.Where("status IN ?", filters.MinStatus)
	}
	if filters.Search != "" {
		query = query.Where("first_name ILIKE ? OR last_name ILIKE ? OR access_code ILIKE ?",
			"%"+filters.Search+"%", "%"+filters.Search+"%", "%"+filters.Search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("Doctor").Preload("District").
		Offset(offset).Limit(limit).Order("updated_at DESC").Find(&patients).Error; err != nil {
		return nil, 0, err
	}

	return patients, total, nil
}

func (r *patientRepository) Update(ctx context.Context, patient *domain.Patient) error {
	return r.db.WithContext(ctx).Save(patient).Error
}

func (r *patientRepository) UpdateStatus(ctx context.Context, id uint, status domain.PatientStatus) error {
	return r.db.WithContext(ctx).Model(&domain.Patient{}).Where("id = ?", id).Update("status", status).Error
}

func (r *patientRepository) CreateStatusHistory(ctx context.Context, h *domain.PatientStatusHistory) error {
	return r.db.WithContext(ctx).Create(h).Error
}

func (r *patientRepository) FindStatusHistory(ctx context.Context, patientID uint) ([]domain.PatientStatusHistory, error) {
	var history []domain.PatientStatusHistory
	err := r.db.WithContext(ctx).Where("patient_id = ?", patientID).Order("created_at ASC").Find(&history).Error
	return history, err
}

func (r *patientRepository) CountByStatus(ctx context.Context, doctorID *uint) (map[domain.PatientStatus]int64, error) {
	type result struct {
		Status domain.PatientStatus
		Count  int64
	}
	var results []result

	query := r.db.WithContext(ctx).Model(&domain.Patient{}).Select("status, count(*) as count").Group("status")
	if doctorID != nil {
		query = query.Where("doctor_id = ?", *doctorID)
	}

	if err := query.Find(&results).Error; err != nil {
		return nil, err
	}

	counts := make(map[domain.PatientStatus]int64)
	for _, r := range results {
		counts[r.Status] = r.Count
	}
	return counts, nil
}
