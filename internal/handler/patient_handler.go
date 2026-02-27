package handler

import (
	"net/http"
	"strconv"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"github.com/beercut-team/backend-boilerplate/internal/middleware"
	"github.com/beercut-team/backend-boilerplate/internal/repository"
	"github.com/beercut-team/backend-boilerplate/internal/service"
	"github.com/gin-gonic/gin"
)

type PatientHandler struct {
	svc service.PatientService
}

func NewPatientHandler(svc service.PatientService) *PatientHandler {
	return &PatientHandler{svc: svc}
}

func (h *PatientHandler) Create(c *gin.Context) {
	var req domain.CreatePatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	doctorID := middleware.GetUserID(c)
	patient, err := h.svc.Create(c.Request.Context(), req, doctorID)
	if err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	Success(c, http.StatusCreated, patient)
}

func (h *PatientHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "неверный id")
		return
	}

	patient, err := h.svc.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		NotFound(c, err.Error())
		return
	}

	Success(c, http.StatusOK, patient)
}

func (h *PatientHandler) GetPublic(c *gin.Context) {
	code := c.Param("code")
	resp, err := h.svc.GetByAccessCode(c.Request.Context(), code)
	if err != nil {
		NotFound(c, err.Error())
		return
	}
	Success(c, http.StatusOK, resp)
}

func (h *PatientHandler) List(c *gin.Context) {
	p := GetPagination(c)
	role := middleware.GetUserRole(c)
	userID := middleware.GetUserID(c)

	filters := repository.PatientFilters{
		Search: c.Query("search"),
	}

	if s := c.Query("status"); s != "" {
		status := domain.PatientStatus(s)
		filters.Status = &status
	}

	// RBAC scoping
	switch role {
	case domain.RoleDistrictDoctor:
		filters.DoctorID = &userID
	case domain.RoleSurgeon:
		filters.MinStatus = []domain.PatientStatus{
			domain.PatientStatusPendingReview,
			domain.PatientStatusApproved,
			domain.PatientStatusScheduled,
			domain.PatientStatusCompleted,
			domain.PatientStatusNeedsCorrection,
		}
	}

	patients, total, err := h.svc.List(c.Request.Context(), filters, p.Offset(), p.Limit)
	if err != nil {
		InternalError(c, "не удалось получить список пациентов")
		return
	}

	SuccessWithMeta(c, http.StatusOK, patients, NewMeta(p.Page, p.Limit, total))
}

func (h *PatientHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "неверный id")
		return
	}

	var req domain.UpdatePatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	patient, err := h.svc.Update(c.Request.Context(), uint(id), req)
	if err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	Success(c, http.StatusOK, patient)
}

func (h *PatientHandler) ChangeStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "неверный id")
		return
	}

	var req domain.PatientStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.svc.ChangeStatus(c.Request.Context(), uint(id), req, userID); err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	Success(c, http.StatusOK, domain.MessageResponse{Message: "статус обновлён"})
}

func (h *PatientHandler) Dashboard(c *gin.Context) {
	role := middleware.GetUserRole(c)
	userID := middleware.GetUserID(c)

	var doctorID *uint
	if role == domain.RoleDistrictDoctor {
		doctorID = &userID
	}

	stats, err := h.svc.DashboardStats(c.Request.Context(), doctorID, role)
	if err != nil {
		InternalError(c, "не удалось получить статистику")
		return
	}

	Success(c, http.StatusOK, stats)
}

func (h *PatientHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "неверный id")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), uint(id)); err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	Success(c, http.StatusOK, domain.MessageResponse{Message: "пациент удалён"})
}

func (h *PatientHandler) RegenerateAccessCode(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "неверный id")
		return
	}

	patient, err := h.svc.RegenerateAccessCode(c.Request.Context(), uint(id))
	if err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	Success(c, http.StatusOK, patient)
}

func (h *PatientHandler) BatchUpdate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "неверный id")
		return
	}

	var req domain.BatchUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	response, err := h.svc.BatchUpdate(c.Request.Context(), uint(id), req, userID)
	if err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if response.Success {
		Success(c, http.StatusOK, response)
	} else {
		Success(c, http.StatusPartialContent, response)
	}
}
