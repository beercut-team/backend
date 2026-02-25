package handler

import (
	"net/http"
	"strconv"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"github.com/beercut-team/backend-boilerplate/internal/middleware"
	"github.com/beercut-team/backend-boilerplate/internal/service"
	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	svc service.NotificationService
}

func NewNotificationHandler(svc service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

func (h *NotificationHandler) List(c *gin.Context) {
	p := GetPagination(c)
	userID := middleware.GetUserID(c)

	notifications, total, err := h.svc.List(c.Request.Context(), userID, p.Offset(), p.Limit)
	if err != nil {
		InternalError(c, "не удалось получить уведомления")
		return
	}

	SuccessWithMeta(c, http.StatusOK, notifications, NewMeta(p.Page, p.Limit, total))
}

func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "неверный id")
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.svc.MarkAsRead(c.Request.Context(), uint(id), userID); err != nil {
		InternalError(c, "не удалось отметить как прочитанное")
		return
	}

	Success(c, http.StatusOK, domain.MessageResponse{Message: "уведомление отмечено как прочитанное"})
}

func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if err := h.svc.MarkAllAsRead(c.Request.Context(), userID); err != nil {
		InternalError(c, "не удалось отметить все как прочитанные")
		return
	}

	Success(c, http.StatusOK, domain.MessageResponse{Message: "все уведомления отмечены как прочитанные"})
}

func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	userID := middleware.GetUserID(c)
	count, err := h.svc.UnreadCount(c.Request.Context(), userID)
	if err != nil {
		InternalError(c, "не удалось подсчитать непрочитанные")
		return
	}

	Success(c, http.StatusOK, gin.H{"count": count})
}
