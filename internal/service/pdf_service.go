package service

import (
	"bytes"
	"context"
	"fmt"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"github.com/beercut-team/backend-boilerplate/internal/repository"
	"github.com/go-pdf/fpdf"
)

type PDFService interface {
	GenerateRoutingSheet(ctx context.Context, patientID uint) (*bytes.Buffer, error)
	GenerateChecklistReport(ctx context.Context, patientID uint) (*bytes.Buffer, error)
}

type pdfService struct {
	patientRepo   repository.PatientRepository
	checklistRepo repository.ChecklistRepository
}

func NewPDFService(patientRepo repository.PatientRepository, checklistRepo repository.ChecklistRepository) PDFService {
	return &pdfService{patientRepo: patientRepo, checklistRepo: checklistRepo}
}

func (s *pdfService) GenerateRoutingSheet(ctx context.Context, patientID uint) (*bytes.Buffer, error) {
	patient, err := s.patientRepo.FindByID(ctx, patientID)
	if err != nil {
		return nil, fmt.Errorf("patient not found: %w", err)
	}

	items, err := s.checklistRepo.FindItemsByPatient(ctx, patientID)
	if err != nil {
		return nil, fmt.Errorf("failed to load checklist: %w", err)
	}

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Use built-in font (Cyrillic support would require adding a TTF font)
	pdf.SetFont("Helvetica", "B", 16)
	pdf.Cell(190, 10, "Routing Sheet / Marshrut list")
	pdf.Ln(15)

	pdf.SetFont("Helvetica", "", 11)
	pdf.Cell(190, 7, fmt.Sprintf("Patient: %s %s %s", patient.LastName, patient.FirstName, patient.MiddleName))
	pdf.Ln(7)
	pdf.Cell(190, 7, fmt.Sprintf("Access Code: %s", patient.AccessCode))
	pdf.Ln(7)
	pdf.Cell(190, 7, fmt.Sprintf("DOB: %s", patient.DateOfBirth.Format("02.01.2006")))
	pdf.Ln(7)
	pdf.Cell(190, 7, fmt.Sprintf("Operation: %s (%s)", patient.OperationType, patient.Eye))
	pdf.Ln(7)
	pdf.Cell(190, 7, fmt.Sprintf("Status: %s", patient.Status))
	pdf.Ln(7)
	pdf.Cell(190, 7, fmt.Sprintf("Diagnosis: %s", patient.Diagnosis))
	pdf.Ln(12)

	// Checklist summary
	pdf.SetFont("Helvetica", "B", 13)
	pdf.Cell(190, 10, "Checklist Items")
	pdf.Ln(10)

	pdf.SetFont("Helvetica", "B", 9)
	pdf.CellFormat(80, 7, "Item", "1", 0, "", false, 0, "")
	pdf.CellFormat(30, 7, "Category", "1", 0, "", false, 0, "")
	pdf.CellFormat(25, 7, "Required", "1", 0, "", false, 0, "")
	pdf.CellFormat(25, 7, "Status", "1", 0, "", false, 0, "")
	pdf.CellFormat(30, 7, "Expires", "1", 1, "", false, 0, "")

	pdf.SetFont("Helvetica", "", 8)
	for _, item := range items {
		required := "No"
		if item.IsRequired {
			required = "Yes"
		}
		expires := ""
		if item.ExpiresAt != nil {
			expires = item.ExpiresAt.Format("02.01.2006")
		}
		pdf.CellFormat(80, 6, item.Name, "1", 0, "", false, 0, "")
		pdf.CellFormat(30, 6, item.Category, "1", 0, "", false, 0, "")
		pdf.CellFormat(25, 6, required, "1", 0, "C", false, 0, "")
		pdf.CellFormat(25, 6, string(item.Status), "1", 0, "C", false, 0, "")
		pdf.CellFormat(30, 6, expires, "1", 1, "C", false, 0, "")
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}
	return &buf, nil
}

func (s *pdfService) GenerateChecklistReport(ctx context.Context, patientID uint) (*bytes.Buffer, error) {
	patient, err := s.patientRepo.FindByID(ctx, patientID)
	if err != nil {
		return nil, fmt.Errorf("patient not found: %w", err)
	}

	items, err := s.checklistRepo.FindItemsByPatient(ctx, patientID)
	if err != nil {
		return nil, fmt.Errorf("failed to load checklist: %w", err)
	}

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Helvetica", "B", 16)
	pdf.Cell(190, 10, "Checklist Report")
	pdf.Ln(15)

	pdf.SetFont("Helvetica", "", 11)
	pdf.Cell(190, 7, fmt.Sprintf("Patient: %s %s %s", patient.LastName, patient.FirstName, patient.MiddleName))
	pdf.Ln(7)
	pdf.Cell(190, 7, fmt.Sprintf("Operation: %s (%s)", patient.OperationType, patient.Eye))
	pdf.Ln(12)

	for _, item := range items {
		pdf.SetFont("Helvetica", "B", 10)
		statusIcon := "[ ]"
		if item.Status == domain.ChecklistStatusCompleted {
			statusIcon = "[X]"
		} else if item.Status == domain.ChecklistStatusRejected {
			statusIcon = "[!]"
		} else if item.Status == domain.ChecklistStatusExpired {
			statusIcon = "[-]"
		}
		pdf.Cell(190, 7, fmt.Sprintf("%s %s", statusIcon, item.Name))
		pdf.Ln(7)

		pdf.SetFont("Helvetica", "", 9)
		if item.Description != "" {
			pdf.Cell(190, 5, fmt.Sprintf("   Description: %s", item.Description))
			pdf.Ln(5)
		}
		if item.Result != "" {
			pdf.Cell(190, 5, fmt.Sprintf("   Result: %s", item.Result))
			pdf.Ln(5)
		}
		if item.ReviewNote != "" {
			pdf.Cell(190, 5, fmt.Sprintf("   Review Note: %s", item.ReviewNote))
			pdf.Ln(5)
		}
		pdf.Ln(3)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}
	return &buf, nil
}
