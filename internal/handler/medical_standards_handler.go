package handler

import (
	"net/http"
	"strconv"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"github.com/beercut-team/backend-boilerplate/internal/service"
	"github.com/gin-gonic/gin"
)

type MedicalStandardsHandler struct {
	medicalSvc service.MedicalStandardsService
}

func NewMedicalStandardsHandler(medicalSvc service.MedicalStandardsService) *MedicalStandardsHandler {
	return &MedicalStandardsHandler{medicalSvc: medicalSvc}
}

// UpdateMedicalMetadata updates medical metadata for a patient
// POST /api/v1/patients/:id/medical-metadata
func (h *MedicalStandardsHandler) UpdateMedicalMetadata(c *gin.Context) {
	patientID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "неверный ID пациента"})
		return
	}

	var req domain.UpdateMedicalMetadataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "неверный формат данных"})
		return
	}

	if err := h.medicalSvc.UpdateMedicalMetadata(c.Request.Context(), uint(patientID), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "медицинские метаданные обновлены"})
}

// SearchICD10Codes searches ICD-10 diagnosis codes
// GET /api/v1/medical-codes/icd10/search?q=катаракта
func (h *MedicalStandardsHandler) SearchICD10Codes(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "параметр q обязателен"})
		return
	}

	results := h.medicalSvc.SearchICD10Codes(c.Request.Context(), query)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results,
		"count":   len(results),
	})
}

// SearchSNOMEDCodes searches SNOMED-CT procedure codes
// GET /api/v1/medical-codes/snomed/search?q=факоэмульсификация
func (h *MedicalStandardsHandler) SearchSNOMEDCodes(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "параметр q обязателен"})
		return
	}

	results := h.medicalSvc.SearchSNOMEDCodes(c.Request.Context(), query)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results,
		"count":   len(results),
	})
}

// SearchLOINCCodes searches LOINC observation codes
// GET /api/v1/medical-codes/loinc/search?q=длина оси
func (h *MedicalStandardsHandler) SearchLOINCCodes(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "параметр q обязателен"})
		return
	}

	results := h.medicalSvc.SearchLOINCCodes(c.Request.Context(), query)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results,
		"count":   len(results),
	})
}
