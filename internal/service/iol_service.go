package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"github.com/beercut-team/backend-boilerplate/internal/repository"
	"github.com/beercut-team/backend-boilerplate/internal/service/formulas"
)

type IOLService interface {
	Calculate(ctx context.Context, req domain.IOLCalculationRequest, userID uint) (*domain.IOLCalculation, error)
	GetHistory(ctx context.Context, patientID uint) ([]domain.IOLCalculation, error)
}

type iolService struct {
	repo repository.IOLRepository
}

func NewIOLService(repo repository.IOLRepository) IOLService {
	return &iolService{repo: repo}
}

func (s *iolService) Calculate(ctx context.Context, req domain.IOLCalculationRequest, userID uint) (*domain.IOLCalculation, error) {
	avgK := (req.Keratometry1 + req.Keratometry2) / 2
	aConst := req.AConstant
	if aConst == 0 {
		aConst = 118.4 // default SRK/T A-constant
	}

	var power, predictedRef float64
	var warnings []string

	// Input validation with warnings
	if req.AxialLength < 20.0 || req.AxialLength > 30.0 {
		warnings = append(warnings, fmt.Sprintf("Axial length %.2f mm is outside normal range (20-30 mm)", req.AxialLength))
	}
	if avgK < 40.0 || avgK > 48.0 {
		warnings = append(warnings, fmt.Sprintf("Average K %.2f D is outside normal range (40-48 D)", avgK))
	}
	if req.ACD > 0 && (req.ACD < 2.0 || req.ACD > 4.5) {
		warnings = append(warnings, fmt.Sprintf("ACD %.2f mm is outside normal range (2.0-4.5 mm)", req.ACD))
	}

	switch strings.ToUpper(req.Formula) {
	case "SRKT", "SRK/T":
		power, predictedRef = formulas.SRKT(req.AxialLength, avgK, aConst, req.TargetRefraction)
	case "HAIGIS":
		if req.ACD == 0 {
			return nil, errors.New("ACD is required for Haigis formula")
		}
		power, predictedRef = formulas.Haigis(req.AxialLength, avgK, req.ACD, req.TargetRefraction)
	case "HOFFERQ", "HOFFER_Q":
		power, predictedRef = formulas.HofferQ(req.AxialLength, avgK, req.ACD, req.TargetRefraction)
	default:
		return nil, errors.New("unsupported formula, use: SRKT, HAIGIS, or HOFFERQ")
	}

	calc := &domain.IOLCalculation{
		PatientID:           req.PatientID,
		Eye:                 req.Eye,
		AxialLength:         req.AxialLength,
		Keratometry1:        req.Keratometry1,
		Keratometry2:        req.Keratometry2,
		ACD:                 req.ACD,
		TargetRefraction:    req.TargetRefraction,
		Formula:             strings.ToUpper(req.Formula),
		IOLPower:            power,
		PredictedRefraction: predictedRef,
		AConstant:           aConst,
		CalculatedBy:        userID,
		Warnings:            strings.Join(warnings, "; "),
	}

	if err := s.repo.Create(ctx, calc); err != nil {
		return nil, errors.New("failed to save calculation")
	}

	return calc, nil
}

func (s *iolService) GetHistory(ctx context.Context, patientID uint) ([]domain.IOLCalculation, error) {
	return s.repo.FindByPatient(ctx, patientID)
}
