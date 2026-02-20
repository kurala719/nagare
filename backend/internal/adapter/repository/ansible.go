package repository

import (
	"nagare/internal/core/domain"
	"nagare/internal/database"
)

// ============= Playbook DAOs =============

func CreatePlaybookDAO(pb *domain.AnsiblePlaybook) error {
	return database.DB.Create(pb).Error
}

func UpdatePlaybookDAO(id uint, pb domain.AnsiblePlaybook) error {
	return database.DB.Model(&domain.AnsiblePlaybook{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":        pb.Name,
		"description": pb.Description,
		"content":     pb.Content,
		"tags":        pb.Tags,
	}).Error
}

func GetPlaybookByIDDAO(id uint) (domain.AnsiblePlaybook, error) {
	var pb domain.AnsiblePlaybook
	err := database.DB.First(&pb, id).Error
	return pb, err
}

func ListPlaybooksDAO(query string) ([]domain.AnsiblePlaybook, error) {
	var pbs []domain.AnsiblePlaybook
	db := database.DB
	if query != "" {
		db = db.Where("name LIKE ? OR description LIKE ? OR tags LIKE ?", "%"+query+"%", "%"+query+"%", "%"+query+"%")
	}
	err := db.Find(&pbs).Error
	return pbs, err
}

func DeletePlaybookDAO(id uint) error {
	return database.DB.Delete(&domain.AnsiblePlaybook{}, id).Error
}

// ============= Job DAOs =============

func CreateAnsibleJobDAO(job *domain.AnsibleJob) error {
	return database.DB.Create(job).Error
}

func UpdateAnsibleJobDAO(id uint, status string, output string) error {
	return database.DB.Model(&domain.AnsibleJob{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status": status,
		"output": output,
	}).Error
}

func AppendAnsibleJobOutputDAO(id uint, additionalOutput string) error {
	// For large outputs, we might want a different strategy, but for now simple append
	return database.DB.Model(&domain.AnsibleJob{}).Where("id = ?", id).
		Update("output", database.DB.Raw("CONCAT(output, ?)", additionalOutput)).Error
}

func GetAnsibleJobByIDDAO(id uint) (domain.AnsibleJob, error) {
	var job domain.AnsibleJob
	err := database.DB.Preload("Playbook").First(&job, id).Error
	return job, err
}

func ListAnsibleJobsDAO(playbookID uint, limit int) ([]domain.AnsibleJob, error) {
	var jobs []domain.AnsibleJob
	query := database.DB.Order("created_at desc")
	if playbookID > 0 {
		query = query.Where("playbook_id = ?", playbookID)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Preload("Playbook").Find(&jobs).Error
	return jobs, err
}
