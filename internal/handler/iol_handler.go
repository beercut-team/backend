package handler

import (
	"net/http"
	"strconv"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"github.com/beercut-team/backend-boilerplate/internal/middleware"
	"github.com/beercut-team/backend-boilerplate/internal/service"
	"github.com/gin-gonic/gin"
)

type IOLHandler struct {
	svc service.IOLService
}

func NewIOLHandler(svc service.IOLService) *IOLHandler {
	return &IOLHandler{svc: svc}
}

func (h *IOLHandler) Calculate(c *gin.Context) {
	var req domain.IOLCalculationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	calc, err := h.svc.Calculate(c.Request.Context(), req, userID)
	if err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	Success(c, http.StatusOK, calc)
}

func (h *IOLHandler) History(c *gin.Context) {
	patientID, err := strconv.ParseUint(c.Param("patientId"), 10, 32)
	if err != nil {
		BadRequest(c, "неверный patient_id")
		return
	}

	calcs, err := h.svc.GetHistory(c.Request.Context(), uint(patientID))
	if err != nil {
		InternalError(c, "не удалось получить историю")
		return
	}

	Success(c, http.StatusOK, calcs)
}
