package handler

import (
	"net/http"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"github.com/beercut-team/backend-boilerplate/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AdminHandler struct {
	authService service.AuthService
	db          *gorm.DB
}

func NewAdminHandler(authService service.AuthService, db *gorm.DB) *AdminHandler {
	return &AdminHandler{authService: authService, db: db}
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	users, err := h.authService.ListUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *AdminHandler) Stats(c *gin.Context) {
	var usersCount, patientsCount, districtsCount, surgeriesCount int64
	h.db.Model(&domain.User{}).Count(&usersCount)
	h.db.Model(&domain.Patient{}).Count(&patientsCount)
	h.db.Model(&domain.District{}).Count(&districtsCount)
	h.db.Model(&domain.Surgery{}).Count(&surgeriesCount)

	c.JSON(http.StatusOK, gin.H{
		"users":     usersCount,
		"patients":  patientsCount,
		"districts": districtsCount,
		"surgeries": surgeriesCount,
	})
}
