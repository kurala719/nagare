package repository

import (
	"gorm.io/gorm"
	"nagare/internal/database"
	"nagare/internal/model"
)

// GetReportConfigDAO retrieves the report configuration
func GetReportConfigDAO() (model.ReportConfig, error) {
	var config model.ReportConfig
	err := database.DB.Limit(1).Find(&config).Error
	if err == nil && config.ID == 0 {
		return config, gorm.ErrRecordNotFound
	}
	return config, err
}

// UpdateReportConfigDAO updates the report configuration
func UpdateReportConfigDAO(config model.ReportConfig) error {
	if config.ID == 0 {
		return database.DB.Create(&config).Error
	}
	return database.DB.Save(&config).Error
}

// AddReportDAO adds a new report record
func AddReportDAO(report *model.Report) error {
	return database.DB.Create(report).Error
}

// GetReportByIDDAO retrieves a report by ID
func GetReportByIDDAO(id uint) (model.Report, error) {
	var report model.Report
	err := database.DB.First(&report, id).Error
	return report, err
}

// ListReportsDAO retrieves reports with pagination
func ListReportsDAO(reportType string, status *int, limit, offset int) ([]model.Report, int64, error) {
	var reports []model.Report
	var total int64
	query := database.DB.Model(&model.Report{})

	if reportType != "" {
		query = query.Where("report_type = ?", reportType)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("created_at desc").Limit(limit).Offset(offset).Find(&reports).Error
	return reports, total, err
}

// UpdateReportStatusDAO updates the status of a report
func UpdateReportStatusDAO(id uint, status int, filePath, downloadURL string) error {
	updates := map[string]interface{}{
		"status":       status,
		"file_path":    filePath,
		"download_url": downloadURL,
	}
	return database.DB.Model(&model.Report{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateReportContentDAO updates the content data of a report
func UpdateReportContentDAO(id uint, content string) error {
	return database.DB.Model(&model.Report{}).Where("id = ?", id).Update("content_data", content).Error
}

// DeleteReportDAO deletes a report record
func DeleteReportDAO(id uint) error {
	return database.DB.Delete(&model.Report{}, id).Error
}
