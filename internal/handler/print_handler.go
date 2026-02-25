package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/beercut-team/backend-boilerplate/internal/service"
	"github.com/gin-gonic/gin"
)

type PrintHandler struct {
	pdfSvc service.PDFService
}

func NewPrintHandler(pdfSvc service.PDFService) *PrintHandler {
	return &PrintHandler{pdfSvc: pdfSvc}
}

func (h *PrintHandler) RoutingSheet(c *gin.Context) {
	patientID, err := strconv.ParseUint(c.Param("patientId"), 10, 32)
	if err != nil {
		BadRequest(c, "invalid patient_id")
		return
	}

	buf, err := h.pdfSvc.GenerateRoutingSheet(c.Request.Context(), uint(patientID))
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=routing_sheet_%d.pdf", patientID))
	c.Data(http.StatusOK, "application/pdf", buf.Bytes())
}

func (h *PrintHandler) ChecklistReport(c *gin.Context) {
	patientID, err := strconv.ParseUint(c.Param("patientId"), 10, 32)
	if err != nil {
		BadRequest(c, "invalid patient_id")
		return
	}

	buf, err := h.pdfSvc.GenerateChecklistReport(c.Request.Context(), uint(patientID))
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=checklist_report_%d.pdf", patientID))
	c.Data(http.StatusOK, "application/pdf", buf.Bytes())
}
