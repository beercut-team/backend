package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"github.com/beercut-team/backend-boilerplate/internal/middleware"
	"github.com/beercut-team/backend-boilerplate/internal/service"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req domain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: err.Error()})
		return
	}

	resp, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusConflict, domain.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: err.Error()})
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req domain.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: err.Error()})
		return
	}

	resp, err := h.authService.Refresh(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID := middleware.GetUserID(c)

	if err := h.authService.Logout(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: "failed to logout"})
		return
	}

	c.JSON(http.StatusOK, domain.MessageResponse{Message: "logged out"})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID := middleware.GetUserID(c)

	resp, err := h.authService.Me(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, domain.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
