package handler

import (
	"net/http"
	"strconv"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"github.com/beercut-team/backend-boilerplate/internal/middleware"
	"github.com/beercut-team/backend-boilerplate/internal/service"
	"github.com/gin-gonic/gin"
)

type ChecklistHandler struct {
	svc service.ChecklistService
}

func NewChecklistHandler(svc service.ChecklistService) *ChecklistHandler {
	return &ChecklistHandler{svc: svc}
}

func (h *ChecklistHandler) GetByPatient(c *gin.Context) {
	patientID, err := strconv.ParseUint(c.Param("patientId"), 10, 32)
	if err != nil {
		BadRequest(c, "invalid patient_id")
		return
	}

	items, err := h.svc.GetByPatient(c.Request.Context(), uint(patientID))
	if err != nil {
		InternalError(c, "failed to get checklist")
		return
	}

	Success(c, http.StatusOK, items)
}

func (h *ChecklistHandler) UpdateItem(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "invalid id")
		return
	}

	var req domain.UpdateChecklistItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	item, err := h.svc.UpdateItem(c.Request.Context(), uint(id), req, userID)
	if err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	Success(c, http.StatusOK, item)
}

func (h *ChecklistHandler) ReviewItem(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "invalid id")
		return
	}

	var req domain.ReviewChecklistItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	reviewerID := middleware.GetUserID(c)
	item, err := h.svc.ReviewItem(c.Request.Context(), uint(id), req, reviewerID)
	if err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	Success(c, http.StatusOK, item)
}

func (h *ChecklistHandler) GetProgress(c *gin.Context) {
	patientID, err := strconv.ParseUint(c.Param("patientId"), 10, 32)
	if err != nil {
		BadRequest(c, "invalid patient_id")
		return
	}

	progress, err := h.svc.GetProgress(c.Request.Context(), uint(patientID))
	if err != nil {
		InternalError(c, "failed to get progress")
		return
	}

	Success(c, http.StatusOK, progress)
}
