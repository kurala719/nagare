package api

import (
	"net/http"
	"os"
	"strconv"

	"nagare/internal/service"

	"github.com/gin-gonic/gin"
)

// ListReportsCtrl handles GET /reports
func ListReportsCtrl(c *gin.Context) {
	reportType := parseOptionalString(c, "type")

	status, err := parseOptionalInt(c, "status")
	if err != nil {
		respondBadRequest(c, "invalid status")
		return
	}

	limit := 50
	if l, err := parseOptionalInt(c, "limit"); err == nil && l != nil {
		if *l > 0 && *l <= 500 {
			limit = *l
		}
	}

	offset := 0
	if o, err := parseOptionalInt(c, "offset"); err == nil && o != nil && *o > 0 {
		offset = *o
	}

	var rType string
	if reportType != nil {
		rType = *reportType
	}

	reports, err := service.ListReportsServ(rType, status, limit, offset)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, reports)
}

// GetReportCtrl handles GET /reports/:id
func GetReportCtrl(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respondBadRequest(c, "invalid report ID")
		return
	}

	report, err := service.GetReportServ(uint(id))
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, report)
}

// GetReportContentCtrl handles GET /reports/:id/content
func GetReportContentCtrl(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respondBadRequest(c, "invalid report ID")
		return
	}

	content, err := service.GetReportContentServ(uint(id))
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, content)
}

// GenerateWeeklyReportCtrl handles POST /reports/generate/weekly
func GenerateWeeklyReportCtrl(c *gin.Context) {
	report, err := service.GenerateWeeklyReportServ()
	if err != nil {
		respondError(c, err)
		return
	}

	resp := service.ReportResp{
		ID:         report.ID,
		ReportType: report.ReportType,
		Title:      report.Title,
		Status:     "pending",
		StatusCode: report.Status,
	}

	respondSuccess(c, http.StatusCreated, resp)
}

// GenerateMonthlyReportCtrl handles POST /reports/generate/monthly
func GenerateMonthlyReportCtrl(c *gin.Context) {
	report, err := service.GenerateMonthlyReportServ()
	if err != nil {
		respondError(c, err)
		return
	}

	resp := service.ReportResp{
		ID:         report.ID,
		ReportType: report.ReportType,
		Title:      report.Title,
		Status:     "pending",
		StatusCode: report.Status,
	}

	respondSuccess(c, http.StatusCreated, resp)
}

// DeleteReportCtrl handles DELETE /reports/:id
func DeleteReportCtrl(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respondBadRequest(c, "invalid report ID")
		return
	}

	if err := service.DeleteReportServ(uint(id)); err != nil {
		respondError(c, err)
		return
	}

	respondSuccessMessage(c, http.StatusOK, "report deleted")
}

// DownloadReportCtrl handles GET /reports/:id/download
func DownloadReportCtrl(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respondBadRequest(c, "invalid report ID")
		return
	}

	report, err := service.GetReportServ(uint(id))
	if err != nil {
		respondError(c, err)
		return
	}

	if report.StatusCode != 1 {
		respondBadRequest(c, "report is not ready for download (status: "+report.Status+")")
		return
	}

	filePath, err := service.GetReportFilePathServ(uint(id))
	if err != nil {
		respondError(c, err)
		return
	}

	// Check if file exists
	if _, err := os.Stat(filePath); err != nil {
		respondBadRequest(c, "report file not found on disk")
		return
	}

	// Serve the PDF file
	c.File(filePath)
}
