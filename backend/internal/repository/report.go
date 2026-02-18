package repository

import (
	"fmt"

	"nagare/internal/database"
	"nagare/internal/model"
)

// AddReportDAO creates a new report record
func AddReportDAO(report model.Report) (model.Report, error) {
	if err := database.DB.Create(&report).Error; err != nil {
		return model.Report{}, fmt.Errorf("failed to create report: %w", err)
	}
	return report, nil
}

// GetReportByIDDAO retrieves a report by ID
func GetReportByIDDAO(id uint) (model.Report, error) {
	var report model.Report
	if err := database.DB.First(&report, id).Error; err != nil {
		return model.Report{}, err
	}
	return report, nil
}

// UpdateReportDAO updates a report
func UpdateReportDAO(id uint, updates map[string]interface{}) error {
	return database.DB.Model(&model.Report{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteReportDAO deletes a report
func DeleteReportDAO(id uint) error {
	return database.DB.Delete(&model.Report{}, id).Error
}

// ListReportsDAO lists reports with pagination and filtering
func ListReportsDAO(reportType *string, status *int, limit int, offset int) ([]model.Report, error) {
	var reports []model.Report
	query := database.DB

	if reportType != nil && *reportType != "" {
		query = query.Where("report_type = ?", *reportType)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&reports).Error; err != nil {
		return nil, err
	}
	return reports, nil
}

// CountReportsDAO counts reports with optional filters
func CountReportsDAO(reportType *string, status *int) (int64, error) {
	var count int64
	query := database.DB

	if reportType != nil && *reportType != "" {
		query = query.Where("report_type = ?", *reportType)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Model(&model.Report{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// AddReportItemDAO adds a report item (host data)
func AddReportItemDAO(item model.ReportItem) (model.ReportItem, error) {
	if err := database.DB.Create(&item).Error; err != nil {
		return model.ReportItem{}, fmt.Errorf("failed to add report item: %w", err)
	}
	return item, nil
}

// GetReportItemsDAO retrieves all items for a report
func GetReportItemsDAO(reportID uint) ([]model.ReportItem, error) {
	var items []model.ReportItem
	if err := database.DB.Where("report_id = ?", reportID).Order("rank ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// GetReportConfigDAO retrieves the report configuration
func GetReportConfigDAO() (model.ReportConfig, error) {
	var config model.ReportConfig
	// There should be at most one config record
	if err := database.DB.First(&config).Error; err != nil {
		return model.ReportConfig{}, err
	}
	return config, nil
}

// UpdateReportConfigDAO updates the report configuration
func UpdateReportConfigDAO(config model.ReportConfig) error {
	if config.ID == 0 {
		// First time creating config
		return database.DB.Create(&config).Error
	}
	return database.DB.Model(&config).Updates(config).Error
}

// GetLatestReportByTypeDAO retrieves the most recent report of a given type
func GetLatestReportByTypeDAO(reportType string) (model.Report, error) {
	var report model.Report
	if err := database.DB.Where("report_type = ? AND status = ?", reportType, 1).
		Order("generated_at DESC").First(&report).Error; err != nil {
		return model.Report{}, err
	}
	return report, nil
}

// GetReportsByDateRangeDAO retrieves reports within a date range
func GetReportsByDateRangeDAO(startDate, endDate string) ([]model.Report, error) {
	var reports []model.Report
	if err := database.DB.Where("generated_at >= ? AND generated_at <= ? AND status = ?", startDate, endDate, 1).
		Order("generated_at DESC").Find(&reports).Error; err != nil {
		return nil, err
	}
	return reports, nil
}
