package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// Medical code types
type ICD10Code struct {
	Code    string `json:"code"`
	Display string `json:"display"`
	System  string `json:"system"`
}

type SNOMEDCode struct {
	Code    string `json:"code"`
	Display string `json:"display"`
	System  string `json:"system"`
}

type LOINCCode struct {
	Code       string    `json:"code"`
	Display    string    `json:"display"`
	System     string    `json:"system"`
	Value      string    `json:"value,omitempty"`
	Unit       string    `json:"unit,omitempty"`
	ObservedAt time.Time `json:"observed_at,omitempty"`
}

type MedicalExtension struct {
	URL   string      `json:"url"`
	Value interface{} `json:"value"`
}

// Integration metadata
type EMIASMetadata struct {
	PatientID  string    `json:"patient_id,omitempty"`
	CaseID     string    `json:"case_id,omitempty"`
	LastSyncAt time.Time `json:"last_sync_at,omitempty"`
	SyncStatus string    `json:"sync_status,omitempty"` // synced, pending, error
}

type RIAMSMetadata struct {
	PatientID  string    `json:"patient_id,omitempty"`
	RegionCode string    `json:"region_code,omitempty"`
	LastSyncAt time.Time `json:"last_sync_at,omitempty"`
	SyncStatus string    `json:"sync_status,omitempty"` // synced, pending, error
}

// Main metadata structure
type MedicalStandardsMetadata struct {
	DiagnosisCodes []ICD10Code       `json:"diagnosis_codes,omitempty"`
	ProcedureCodes []SNOMEDCode      `json:"procedure_codes,omitempty"`
	Observations   []LOINCCode       `json:"observations,omitempty"`
	FHIRResourceID string            `json:"fhir_resource_id,omitempty"`
	Extensions     []MedicalExtension `json:"extensions,omitempty"`
	Integrations   *IntegrationMetadata `json:"integrations,omitempty"`
}

type IntegrationMetadata struct {
	EMIAS *EMIASMetadata `json:"emias,omitempty"`
	RIAMS *RIAMSMetadata `json:"riams,omitempty"`
}

// Request/Response types
type UpdateMedicalMetadataRequest struct {
	DiagnosisCodes []ICD10Code       `json:"diagnosis_codes,omitempty"`
	ProcedureCodes []SNOMEDCode      `json:"procedure_codes,omitempty"`
	Observations   []LOINCCode       `json:"observations,omitempty"`
	Extensions     []MedicalExtension `json:"extensions,omitempty"`
}

type SearchMedicalCodesRequest struct {
	Query string `json:"query" form:"q" binding:"required"`
}

type MedicalCodeSearchResult struct {
	Code    string `json:"code"`
	Display string `json:"display"`
	System  string `json:"system"`
}

// Scan implements sql.Scanner for JSONB support
func (m *MedicalStandardsMetadata) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to unmarshal JSONB value")
	}

	return json.Unmarshal(bytes, m)
}

// Value implements driver.Valuer for JSONB support
func (m MedicalStandardsMetadata) Value() (driver.Value, error) {
	if m.DiagnosisCodes == nil && m.ProcedureCodes == nil && m.Observations == nil {
		return nil, nil
	}
	return json.Marshal(m)
}
