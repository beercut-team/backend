package handler

import (
	"net/http"
	"strconv"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"github.com/beercut-team/backend-boilerplate/internal/service"
	"github.com/gin-gonic/gin"
)

type IntegrationsHandler struct {
	integrationsSvc service.IntegrationsService
}

func NewIntegrationsHandler(integrationsSvc service.IntegrationsService) *IntegrationsHandler {
	return &IntegrationsHandler{integrationsSvc: integrationsSvc}
}

// EMIAS endpoints

// ExportToEMIAS exports patient to EMIAS
// POST /api/v1/integrations/emias/patients/:id/export
func (h *IntegrationsHandler) ExportToEMIAS(c *gin.Context) {
	patientID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "неверный ID пациента"})
		return
	}

	// Validate first
	validation, err := h.integrationsSvc.ValidateForEMIAS(c.Request.Context(), uint(patientID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	if !validation.Valid {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":  false,
			"error":    "валидация не пройдена",
			"errors":   validation.Errors,
			"warnings": validation.Warnings,
		})
		return
	}

	result, err := h.integrationsSvc.ExportToEMIAS(c.Request.Context(), uint(patientID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, result)
		return
	}

	c.JSON(http.StatusOK, result)
}

// CreateEMIASCase creates a case in EMIAS
// POST /api/v1/integrations/emias/patients/:id/case
func (h *IntegrationsHandler) CreateEMIASCase(c *gin.Context) {
	patientID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "неверный ID пациента"})
		return
	}

	var req domain.EMIASExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "неверный формат данных"})
		return
	}

	req.PatientID = uint(patientID)

	result, err := h.integrationsSvc.CreateEMIASCase(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, result)
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetEMIASStatus gets EMIAS sync status
// GET /api/v1/integrations/emias/patients/:id/status
func (h *IntegrationsHandler) GetEMIASStatus(c *gin.Context) {
	patientID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "неверный ID пациента"})
		return
	}

	result, err := h.integrationsSvc.GetEMIASStatus(c.Request.Context(), uint(patientID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// RIAMS endpoints

// ExportToRIAMS exports patient to RIAMS
// POST /api/v1/integrations/riams/patients/:id/export
func (h *IntegrationsHandler) ExportToRIAMS(c *gin.Context) {
	patientID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "неверный ID пациента"})
		return
	}

	var req domain.RIAMSExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "неверный формат данных"})
		return
	}

	if req.RegionCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "region_code обязателен"})
		return
	}

	// Validate first
	validation, err := h.integrationsSvc.ValidateForRIAMS(c.Request.Context(), uint(patientID), req.RegionCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	if !validation.Valid {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":  false,
			"error":    "валидация не пройдена",
			"errors":   validation.Errors,
			"warnings": validation.Warnings,
		})
		return
	}

	result, err := h.integrationsSvc.ExportToRIAMS(c.Request.Context(), uint(patientID), req.RegionCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, result)
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetRIAMSStatus gets RIAMS sync status
// GET /api/v1/integrations/riams/patients/:id/status
func (h *IntegrationsHandler) GetRIAMSStatus(c *gin.Context) {
	patientID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "неверный ID пациента"})
		return
	}

	result, err := h.integrationsSvc.GetRIAMSStatus(c.Request.Context(), uint(patientID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetRIAMSRegions gets list of supported RIAMS regions
// GET /api/v1/integrations/riams/regions
func (h *IntegrationsHandler) GetRIAMSRegions(c *gin.Context) {
	regions := h.integrationsSvc.GetRIAMSRegions(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    regions,
		"count":   len(regions),
	})
}
