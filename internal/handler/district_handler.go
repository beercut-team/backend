package handler

import (
	"net/http"
	"strconv"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"github.com/beercut-team/backend-boilerplate/internal/service"
	"github.com/gin-gonic/gin"
)

type DistrictHandler struct {
	svc service.DistrictService
}

func NewDistrictHandler(svc service.DistrictService) *DistrictHandler {
	return &DistrictHandler{svc: svc}
}

func (h *DistrictHandler) Create(c *gin.Context) {
	var req domain.CreateDistrictRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	district, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		Error(c, http.StatusConflict, err.Error())
		return
	}

	Success(c, http.StatusCreated, district)
}

func (h *DistrictHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "неверный id")
		return
	}

	district, err := h.svc.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		NotFound(c, err.Error())
		return
	}

	Success(c, http.StatusOK, district)
}

func (h *DistrictHandler) List(c *gin.Context) {
	p := GetPagination(c)
	search := c.Query("search")

	districts, total, err := h.svc.List(c.Request.Context(), search, p.Offset(), p.Limit)
	if err != nil {
		InternalError(c, "не удалось получить список районов")
		return
	}

	SuccessWithMeta(c, http.StatusOK, districts, NewMeta(p.Page, p.Limit, total))
}

func (h *DistrictHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "неверный id")
		return
	}

	var req domain.UpdateDistrictRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	district, err := h.svc.Update(c.Request.Context(), uint(id), req)
	if err != nil {
		NotFound(c, err.Error())
		return
	}

	Success(c, http.StatusOK, district)
}

func (h *DistrictHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "неверный id")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), uint(id)); err != nil {
		NotFound(c, err.Error())
		return
	}

	Success(c, http.StatusOK, domain.MessageResponse{Message: "район удалён"})
}
