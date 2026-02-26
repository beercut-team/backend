package domain

import "time"

// EMIAS integration types
type EMIASExportRequest struct {
	PatientID    uint   `json:"patient_id"`
	SurgeryDate  string `json:"surgery_date,omitempty"`
	ProcedureCode string `json:"procedure_code,omitempty"`
	DiagnosisCode string `json:"diagnosis_code,omitempty"`
}

type EMIASExportResponse struct {
	Success    bool   `json:"success"`
	ExternalID string `json:"external_id,omitempty"`
	Message    string `json:"message,omitempty"`
	Error      string `json:"error,omitempty"`
}

type EMIASStatusResponse struct {
	Success    bool      `json:"success"`
	PatientID  string    `json:"patient_id,omitempty"`
	CaseID     string    `json:"case_id,omitempty"`
	Status     string    `json:"status,omitempty"` // synced, pending, error
	LastSyncAt time.Time `json:"last_sync_at,omitempty"`
}

// RIAMS integration types
type RIAMSExportRequest struct {
	PatientID  uint   `json:"patient_id"`
	RegionCode string `json:"region_code"` // required
}

type RIAMSExportResponse struct {
	Success    bool   `json:"success"`
	ExternalID string `json:"external_id,omitempty"`
	Message    string `json:"message,omitempty"`
	Error      string `json:"error,omitempty"`
}

type RIAMSStatusResponse struct {
	Success    bool      `json:"success"`
	PatientID  string    `json:"patient_id,omitempty"`
	RegionCode string    `json:"region_code,omitempty"`
	Status     string    `json:"status,omitempty"` // synced, pending, error
	LastSyncAt time.Time `json:"last_sync_at,omitempty"`
}

type RIAMSRegion struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// Validation result
type IntegrationValidationResult struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}
