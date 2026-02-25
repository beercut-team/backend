package handler

import (
	"net/http"
	"strconv"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"github.com/beercut-team/backend-boilerplate/internal/middleware"
	"github.com/beercut-team/backend-boilerplate/internal/service"
	"github.com/gin-gonic/gin"
)

type CommentHandler struct {
	svc service.CommentService
}

func NewCommentHandler(svc service.CommentService) *CommentHandler {
	return &CommentHandler{svc: svc}
}

func (h *CommentHandler) Create(c *gin.Context) {
	var req domain.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	authorID := middleware.GetUserID(c)
	comment, err := h.svc.Create(c.Request.Context(), req, authorID)
	if err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	Success(c, http.StatusCreated, comment)
}

func (h *CommentHandler) GetByPatient(c *gin.Context) {
	patientID, err := strconv.ParseUint(c.Param("patientId"), 10, 32)
	if err != nil {
		BadRequest(c, "invalid patient_id")
		return
	}

	comments, err := h.svc.GetByPatient(c.Request.Context(), uint(patientID))
	if err != nil {
		InternalError(c, "failed to get comments")
		return
	}

	Success(c, http.StatusOK, comments)
}

func (h *CommentHandler) MarkAsRead(c *gin.Context) {
	patientID, err := strconv.ParseUint(c.Param("patientId"), 10, 32)
	if err != nil {
		BadRequest(c, "invalid patient_id")
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.svc.MarkAsRead(c.Request.Context(), uint(patientID), userID); err != nil {
		InternalError(c, "failed to mark as read")
		return
	}

	Success(c, http.StatusOK, domain.MessageResponse{Message: "comments marked as read"})
}
