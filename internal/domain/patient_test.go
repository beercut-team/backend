package domain

import (
	"testing"
)

func TestValidateStatusTransition(t *testing.T) {
	tests := []struct {
		name     string
		from     PatientStatus
		to       PatientStatus
		expected bool
	}{
		// Valid transitions
		{"Draft to InProgress", PatientStatusDraft, PatientStatusInProgress, true},
		{"InProgress to PendingReview", PatientStatusInProgress, PatientStatusPendingReview, true},
		{"PendingReview to Approved", PatientStatusPendingReview, PatientStatusApproved, true},
		{"PendingReview to NeedsCorrection", PatientStatusPendingReview, PatientStatusNeedsCorrection, true},
		{"NeedsCorrection to InProgress", PatientStatusNeedsCorrection, PatientStatusInProgress, true},
		{"Approved to Scheduled", PatientStatusApproved, PatientStatusScheduled, true},
		{"Scheduled to Completed", PatientStatusScheduled, PatientStatusCompleted, true},

		// Cancel from any status
		{"Draft to Cancelled", PatientStatusDraft, PatientStatusCancelled, true},
		{"InProgress to Cancelled", PatientStatusInProgress, PatientStatusCancelled, true},
		{"Approved to Cancelled", PatientStatusApproved, PatientStatusCancelled, true},

		// Invalid transitions
		{"Draft to Approved", PatientStatusDraft, PatientStatusApproved, false},
		{"Draft to Scheduled", PatientStatusDraft, PatientStatusScheduled, false},
		{"InProgress to Scheduled", PatientStatusInProgress, PatientStatusScheduled, false},
		{"Approved to InProgress", PatientStatusApproved, PatientStatusInProgress, false},
		{"Completed to InProgress", PatientStatusCompleted, PatientStatusInProgress, false},
		{"Cancelled to InProgress", PatientStatusCancelled, PatientStatusInProgress, false},

		// No transitions from final states
		{"Completed to Scheduled", PatientStatusCompleted, PatientStatusScheduled, false},
		{"Cancelled to Approved", PatientStatusCancelled, PatientStatusApproved, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateStatusTransition(tt.from, tt.to)
			if result != tt.expected {
				t.Errorf("ValidateStatusTransition(%s, %s) = %v, want %v",
					tt.from, tt.to, result, tt.expected)
			}
		})
	}
}

func TestGetStatusDisplayName(t *testing.T) {
	tests := []struct {
		status   PatientStatus
		expected string
	}{
		{PatientStatusDraft, "Черновик"},
		{PatientStatusInProgress, "В процессе подготовки"},
		{PatientStatusPendingReview, "Ожидает проверки хирурга"},
		{PatientStatusApproved, "Одобрено, готов к операции"},
		{PatientStatusNeedsCorrection, "Требуется доработка"},
		{PatientStatusScheduled, "Операция запланирована"},
		{PatientStatusCompleted, "Операция завершена"},
		{PatientStatusCancelled, "Отменено"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			result := GetStatusDisplayName(tt.status)
			if result != tt.expected {
				t.Errorf("GetStatusDisplayName(%s) = %s, want %s",
					tt.status, result, tt.expected)
			}
		})
	}
}

func TestGenerateAccessCode(t *testing.T) {
	// Generate multiple codes and check uniqueness
	codes := make(map[string]bool)
	for i := 0; i < 100; i++ {
		code := GenerateAccessCode()

		// Check length (4 bytes = 8 hex chars)
		if len(code) != 8 {
			t.Errorf("GenerateAccessCode() length = %d, want 8", len(code))
		}

		// Check uniqueness
		if codes[code] {
			t.Errorf("GenerateAccessCode() generated duplicate: %s", code)
		}
		codes[code] = true
	}
}
