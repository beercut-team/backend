package handler

import (
	"net/http"
	"strconv"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"github.com/beercut-team/backend-boilerplate/internal/middleware"
	"github.com/beercut-team/backend-boilerplate/internal/service"
	"github.com/gin-gonic/gin"
)

type SurgeryHandler struct {
	svc service.SurgeryService
}

func NewSurgeryHandler(svc service.SurgeryService) *SurgeryHandler {
	return &SurgeryHandler{svc: svc}
}

func (h *SurgeryHandler) Schedule(c *gin.Context) {
	var req domain.CreateSurgeryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	surgeonID := middleware.GetUserID(c)
	surgery, err := h.svc.Schedule(c.Request.Context(), req, surgeonID)
	if err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	Success(c, http.StatusCreated, surgery)
}

func (h *SurgeryHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "invalid id")
		return
	}

	surgery, err := h.svc.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		NotFound(c, err.Error())
		return
	}

	Success(c, http.StatusOK, surgery)
}

func (h *SurgeryHandler) List(c *gin.Context) {
	p := GetPagination(c)
	surgeonID := middleware.GetUserID(c)

	surgeries, total, err := h.svc.ListBySurgeon(c.Request.Context(), surgeonID, p.Offset(), p.Limit)
	if err != nil {
		InternalError(c, "failed to list surgeries")
		return
	}

	SuccessWithMeta(c, http.StatusOK, surgeries, NewMeta(p.Page, p.Limit, total))
}

func (h *SurgeryHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "invalid id")
		return
	}

	var req domain.UpdateSurgeryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	surgery, err := h.svc.Update(c.Request.Context(), uint(id), req)
	if err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	Success(c, http.StatusOK, surgery)
}
