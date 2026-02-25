package handler

import (
	"net/http"
	"strconv"

	"github.com/beercut-team/backend-boilerplate/internal/middleware"
	"github.com/beercut-team/backend-boilerplate/internal/service"
	"github.com/gin-gonic/gin"
)

type MediaHandler struct {
	svc service.MediaService
}

func NewMediaHandler(svc service.MediaService) *MediaHandler {
	return &MediaHandler{svc: svc}
}

func (h *MediaHandler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		BadRequest(c, "файл обязателен")
		return
	}
	defer file.Close()

	patientIDStr := c.PostForm("patient_id")
	patientID, err := strconv.ParseUint(patientIDStr, 10, 32)
	if err != nil {
		BadRequest(c, "неверный patient_id")
		return
	}

	category := c.PostForm("category")
	if category == "" {
		category = "general"
	}

	userID := middleware.GetUserID(c)
	media, err := h.svc.Upload(
		c.Request.Context(),
		uint(patientID),
		userID,
		header.Filename,
		header.Header.Get("Content-Type"),
		category,
		header.Size,
		file,
	)
	if err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	Success(c, http.StatusCreated, media)
}

func (h *MediaHandler) GetByPatient(c *gin.Context) {
	patientID, err := strconv.ParseUint(c.Param("patientId"), 10, 32)
	if err != nil {
		BadRequest(c, "неверный patient_id")
		return
	}

	media, err := h.svc.GetByPatient(c.Request.Context(), uint(patientID))
	if err != nil {
		InternalError(c, "не удалось получить медиафайлы")
		return
	}

	Success(c, http.StatusOK, media)
}

func (h *MediaHandler) Download(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "неверный id")
		return
	}

	url, err := h.svc.GetDownloadURL(c.Request.Context(), uint(id))
	if err != nil {
		NotFound(c, err.Error())
		return
	}

	Success(c, http.StatusOK, gin.H{"url": url})
}

func (h *MediaHandler) Thumbnail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "неверный id")
		return
	}

	url, err := h.svc.GetThumbnailURL(c.Request.Context(), uint(id))
	if err != nil {
		NotFound(c, err.Error())
		return
	}

	Success(c, http.StatusOK, gin.H{"url": url})
}

func (h *MediaHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "неверный id")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), uint(id)); err != nil {
		NotFound(c, err.Error())
		return
	}

	Success(c, http.StatusOK, gin.H{"message": "удалено"})
}
