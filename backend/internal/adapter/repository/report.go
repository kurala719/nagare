package repository

import (
	"nagare/internal/core/domain"
	"nagare/internal/database"
)

// GetReportConfigDAO retrieves the report configuration
func GetReportConfigDAO() (domain.ReportConfig, error) {
	var config domain.ReportConfig
	err := database.DB.First(&config).Error
	return config, err
}

// UpdateReportConfigDAO updates the report configuration
func UpdateReportConfigDAO(config domain.ReportConfig) error {
	if config.ID == 0 {
		return database.DB.Create(&config).Error
	}
	return database.DB.Save(&config).Error
}

// AddReportDAO adds a new report record
func AddReportDAO(report *domain.Report) error {
	return database.DB.Create(report).Error
}

// GetReportByIDDAO retrieves a report by ID
func GetReportByIDDAO(id uint) (domain.Report, error) {
	var report domain.Report
	err := database.DB.First(&report, id).Error
	return report, err
}

// ListReportsDAO retrieves reports with pagination
func ListReportsDAO(reportType string, status *int, limit, offset int) ([]domain.Report, int64, error) {
	var reports []domain.Report
	var total int64
	query := database.DB.Model(&domain.Report{})

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
	return database.DB.Model(&domain.Report{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateReportContentDAO updates the content data of a report
func UpdateReportContentDAO(id uint, content string) error {
	return database.DB.Model(&domain.Report{}).Where("id = ?", id).Update("content_data", content).Error
}

// DeleteReportDAO deletes a report record
func DeleteReportDAO(id uint) error {
	return database.DB.Delete(&domain.Report{}, id).Error
}
