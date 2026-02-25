package handler

import (
	"net/http"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"github.com/beercut-team/backend-boilerplate/internal/middleware"
	"github.com/beercut-team/backend-boilerplate/internal/service"
	"github.com/gin-gonic/gin"
)

type SyncHandler struct {
	svc service.SyncService
}

func NewSyncHandler(svc service.SyncService) *SyncHandler {
	return &SyncHandler{svc: svc}
}

func (h *SyncHandler) Push(c *gin.Context) {
	var req domain.SyncPushRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.svc.Push(c.Request.Context(), userID, req); err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	Success(c, http.StatusOK, domain.MessageResponse{Message: "синхронизация завершена"})
}

func (h *SyncHandler) Pull(c *gin.Context) {
	since := c.Query("since")
	if since == "" {
		BadRequest(c, "параметр since обязателен (ISO 8601)")
		return
	}

	userID := middleware.GetUserID(c)
	resp, err := h.svc.Pull(c.Request.Context(), userID, since)
	if err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	Success(c, http.StatusOK, resp)
}
