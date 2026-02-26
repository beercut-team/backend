package service

import (
	"context"
	"fmt"
	"time"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"github.com/beercut-team/backend-boilerplate/internal/repository"
	"github.com/google/uuid"
)

type IntegrationsService interface {
	// EMIAS
	ValidateForEMIAS(ctx context.Context, patientID uint) (*domain.IntegrationValidationResult, error)
	ExportToEMIAS(ctx context.Context, patientID uint) (*domain.EMIASExportResponse, error)
	CreateEMIASCase(ctx context.Context, req domain.EMIASExportRequest) (*domain.EMIASExportResponse, error)
	GetEMIASStatus(ctx context.Context, patientID uint) (*domain.EMIASStatusResponse, error)

	// RIAMS
	ValidateForRIAMS(ctx context.Context, patientID uint, regionCode string) (*domain.IntegrationValidationResult, error)
	ExportToRIAMS(ctx context.Context, patientID uint, regionCode string) (*domain.RIAMSExportResponse, error)
	GetRIAMSStatus(ctx context.Context, patientID uint) (*domain.RIAMSStatusResponse, error)
	GetRIAMSRegions(ctx context.Context) []domain.RIAMSRegion
}

type integrationsService struct {
	patientRepo repository.PatientRepository
}

func NewIntegrationsService(patientRepo repository.PatientRepository) IntegrationsService {
	return &integrationsService{patientRepo: patientRepo}
}

// EMIAS methods
func (s *integrationsService) ValidateForEMIAS(ctx context.Context, patientID uint) (*domain.IntegrationValidationResult, error) {
	patient, err := s.patientRepo.FindByID(ctx, patientID)
	if err != nil {
		return nil, err
	}

	result := &domain.IntegrationValidationResult{Valid: true}

	// Required fields
	if patient.FirstName == "" || patient.LastName == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "ФИО пациента обязательно")
	}
	if patient.DateOfBirth.IsZero() {
		result.Valid = false
		result.Errors = append(result.Errors, "Дата рождения обязательна")
	}
	if patient.SNILs == "" {
		result.Warnings = append(result.Warnings, "СНИЛС не указан")
	}
	if patient.OMSPolicy == "" {
		result.Warnings = append(result.Warnings, "Полис ОМС не указан")
	}

	return result, nil
}

func (s *integrationsService) ExportToEMIAS(ctx context.Context, patientID uint) (*domain.EMIASExportResponse, error) {
	patient, err := s.patientRepo.FindByID(ctx, patientID)
	if err != nil {
		return &domain.EMIASExportResponse{Success: false, Error: "пациент не найден"}, err
	}

	// Mock implementation - generate external ID
	externalID := fmt.Sprintf("EMIAS-%s", uuid.New().String()[:8])

	// Update patient metadata
	if patient.MedicalMetadata == nil {
		patient.MedicalMetadata = &domain.MedicalStandardsMetadata{}
	}
	if patient.MedicalMetadata.Integrations == nil {
		patient.MedicalMetadata.Integrations = &domain.IntegrationMetadata{}
	}
	if patient.MedicalMetadata.Integrations.EMIAS == nil {
		patient.MedicalMetadata.Integrations.EMIAS = &domain.EMIASMetadata{}
	}

	patient.MedicalMetadata.Integrations.EMIAS.PatientID = externalID
	patient.MedicalMetadata.Integrations.EMIAS.LastSyncAt = time.Now()
	patient.MedicalMetadata.Integrations.EMIAS.SyncStatus = "synced"

	if err := s.patientRepo.Update(ctx, patient); err != nil {
		return &domain.EMIASExportResponse{Success: false, Error: "не удалось обновить метаданные"}, err
	}

	return &domain.EMIASExportResponse{
		Success:    true,
		ExternalID: externalID,
		Message:    "Пациент успешно экспортирован в ЕМИАС",
	}, nil
}

func (s *integrationsService) CreateEMIASCase(ctx context.Context, req domain.EMIASExportRequest) (*domain.EMIASExportResponse, error) {
	patient, err := s.patientRepo.FindByID(ctx, req.PatientID)
	if err != nil {
		return &domain.EMIASExportResponse{Success: false, Error: "пациент не найден"}, err
	}

	// Mock implementation
	caseID := fmt.Sprintf("CASE-%s", uuid.New().String()[:8])

	if patient.MedicalMetadata == nil {
		patient.MedicalMetadata = &domain.MedicalStandardsMetadata{}
	}
	if patient.MedicalMetadata.Integrations == nil {
		patient.MedicalMetadata.Integrations = &domain.IntegrationMetadata{}
	}
	if patient.MedicalMetadata.Integrations.EMIAS == nil {
		patient.MedicalMetadata.Integrations.EMIAS = &domain.EMIASMetadata{}
	}

	patient.MedicalMetadata.Integrations.EMIAS.CaseID = caseID
	patient.MedicalMetadata.Integrations.EMIAS.LastSyncAt = time.Now()
	patient.MedicalMetadata.Integrations.EMIAS.SyncStatus = "synced"

	if err := s.patientRepo.Update(ctx, patient); err != nil {
		return &domain.EMIASExportResponse{Success: false, Error: "не удалось обновить метаданные"}, err
	}

	return &domain.EMIASExportResponse{
		Success:    true,
		ExternalID: caseID,
		Message:    "Случай успешно создан в ЕМИАС",
	}, nil
}

func (s *integrationsService) GetEMIASStatus(ctx context.Context, patientID uint) (*domain.EMIASStatusResponse, error) {
	patient, err := s.patientRepo.FindByID(ctx, patientID)
	if err != nil {
		return &domain.EMIASStatusResponse{Success: false}, err
	}

	if patient.MedicalMetadata == nil ||
		patient.MedicalMetadata.Integrations == nil ||
		patient.MedicalMetadata.Integrations.EMIAS == nil {
		return &domain.EMIASStatusResponse{
			Success: true,
			Status:  "not_synced",
		}, nil
	}

	emias := patient.MedicalMetadata.Integrations.EMIAS
	return &domain.EMIASStatusResponse{
		Success:    true,
		PatientID:  emias.PatientID,
		CaseID:     emias.CaseID,
		Status:     emias.SyncStatus,
		LastSyncAt: emias.LastSyncAt,
	}, nil
}

// RIAMS methods
func (s *integrationsService) ValidateForRIAMS(ctx context.Context, patientID uint, regionCode string) (*domain.IntegrationValidationResult, error) {
	patient, err := s.patientRepo.FindByID(ctx, patientID)
	if err != nil {
		return nil, err
	}

	result := &domain.IntegrationValidationResult{Valid: true}

	if patient.FirstName == "" || patient.LastName == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "ФИО пациента обязательно")
	}
	if patient.DateOfBirth.IsZero() {
		result.Valid = false
		result.Errors = append(result.Errors, "Дата рождения обязательна")
	}
	if regionCode == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "Код региона обязателен")
	}

	return result, nil
}

func (s *integrationsService) ExportToRIAMS(ctx context.Context, patientID uint, regionCode string) (*domain.RIAMSExportResponse, error) {
	patient, err := s.patientRepo.FindByID(ctx, patientID)
	if err != nil {
		return &domain.RIAMSExportResponse{Success: false, Error: "пациент не найден"}, err
	}

	// Mock implementation
	externalID := fmt.Sprintf("RIAMS-%s-%s", regionCode, uuid.New().String()[:8])

	if patient.MedicalMetadata == nil {
		patient.MedicalMetadata = &domain.MedicalStandardsMetadata{}
	}
	if patient.MedicalMetadata.Integrations == nil {
		patient.MedicalMetadata.Integrations = &domain.IntegrationMetadata{}
	}
	if patient.MedicalMetadata.Integrations.RIAMS == nil {
		patient.MedicalMetadata.Integrations.RIAMS = &domain.RIAMSMetadata{}
	}

	patient.MedicalMetadata.Integrations.RIAMS.PatientID = externalID
	patient.MedicalMetadata.Integrations.RIAMS.RegionCode = regionCode
	patient.MedicalMetadata.Integrations.RIAMS.LastSyncAt = time.Now()
	patient.MedicalMetadata.Integrations.RIAMS.SyncStatus = "synced"

	if err := s.patientRepo.Update(ctx, patient); err != nil {
		return &domain.RIAMSExportResponse{Success: false, Error: "не удалось обновить метаданные"}, err
	}

	return &domain.RIAMSExportResponse{
		Success:    true,
		ExternalID: externalID,
		Message:    "Пациент успешно экспортирован в РИАМС",
	}, nil
}

func (s *integrationsService) GetRIAMSStatus(ctx context.Context, patientID uint) (*domain.RIAMSStatusResponse, error) {
	patient, err := s.patientRepo.FindByID(ctx, patientID)
	if err != nil {
		return &domain.RIAMSStatusResponse{Success: false}, err
	}

	if patient.MedicalMetadata == nil ||
		patient.MedicalMetadata.Integrations == nil ||
		patient.MedicalMetadata.Integrations.RIAMS == nil {
		return &domain.RIAMSStatusResponse{
			Success: true,
			Status:  "not_synced",
		}, nil
	}

	riams := patient.MedicalMetadata.Integrations.RIAMS
	return &domain.RIAMSStatusResponse{
		Success:    true,
		PatientID:  riams.PatientID,
		RegionCode: riams.RegionCode,
		Status:     riams.SyncStatus,
		LastSyncAt: riams.LastSyncAt,
	}, nil
}

func (s *integrationsService) GetRIAMSRegions(ctx context.Context) []domain.RIAMSRegion {
	return []domain.RIAMSRegion{
		{Code: "77", Name: "Москва"},
		{Code: "78", Name: "Санкт-Петербург"},
		{Code: "50", Name: "Московская область"},
		{Code: "47", Name: "Ленинградская область"},
		{Code: "23", Name: "Краснодарский край"},
		{Code: "61", Name: "Ростовская область"},
		{Code: "66", Name: "Свердловская область"},
		{Code: "54", Name: "Новосибирская область"},
		{Code: "74", Name: "Челябинская область"},
		{Code: "16", Name: "Республика Татарстан"},
	}
}
