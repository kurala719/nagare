package api

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"nagare/internal/service"

	"github.com/gin-gonic/gin"
)

// ListReportsCtrl handles GET /analysis/reports
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

// GetReportCtrl handles GET /analysis/reports/:id
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

// GetReportContentCtrl handles GET /analysis/reports/:id/content
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

// GenerateDailyReportCtrl handles POST /analysis/reports/generations/daily
func GenerateDailyReportCtrl(c *gin.Context) {
	report, err := service.GenerateDailyReportServ()
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

// GenerateWeeklyReportCtrl handles POST /analysis/reports/generations/weekly
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

// GenerateMonthlyReportCtrl handles POST /analysis/reports/generations/monthly
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

// GenerateCustomReportCtrl handles POST /analysis/reports/generations/custom
func GenerateCustomReportCtrl(c *gin.Context) {
	var req struct {
		Title     string `json:"title"`
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, "invalid request body")
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		respondBadRequest(c, "invalid start_time format (RFC3339 required)")
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		respondBadRequest(c, "invalid end_time format (RFC3339 required)")
		return
	}

	report, err := service.GenerateCustomReportServ(req.Title, startTime, endTime)
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

// DeleteReportCtrl handles DELETE /analysis/reports/:id
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

// DownloadReportCtrl handles GET /analysis/reports/:id/file
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
