package repository

import (
	"nagare/internal/database"
	"nagare/internal/model"
)

// ============= Playbook DAOs =============

func CreatePlaybookDAO(pb *model.AnsiblePlaybook) error {
	return database.DB.Create(pb).Error
}

func UpdatePlaybookDAO(id uint, pb model.AnsiblePlaybook) error {
	return database.DB.Model(&model.AnsiblePlaybook{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":        pb.Name,
		"description": pb.Description,
		"content":     pb.Content,
		"tags":        pb.Tags,
	}).Error
}

func GetPlaybookByIDDAO(id uint) (model.AnsiblePlaybook, error) {
	var pb model.AnsiblePlaybook
	err := database.DB.First(&pb, id).Error
	return pb, err
}

func ListPlaybooksDAO(query string) ([]model.AnsiblePlaybook, error) {
	var pbs []model.AnsiblePlaybook
	db := database.DB
	if query != "" {
		db = db.Where("name LIKE ? OR description LIKE ? OR tags LIKE ?", "%"+query+"%", "%"+query+"%", "%"+query+"%")
	}
	err := db.Find(&pbs).Error
	return pbs, err
}

func DeletePlaybookDAO(id uint) error {
	return database.DB.Delete(&model.AnsiblePlaybook{}, id).Error
}

// ============= Job DAOs =============

func CreateAnsibleJobDAO(job *model.AnsibleJob) error {
	return database.DB.Create(job).Error
}

func UpdateAnsibleJobDAO(id uint, status string, output string) error {
	return database.DB.Model(&model.AnsibleJob{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status": status,
		"output": output,
	}).Error
}

func AppendAnsibleJobOutputDAO(id uint, additionalOutput string) error {
	// For large outputs, we might want a different strategy, but for now simple append
	return database.DB.Model(&model.AnsibleJob{}).Where("id = ?", id).
		Update("output", database.DB.Raw("CONCAT(output, ?)", additionalOutput)).Error
}

func GetAnsibleJobByIDDAO(id uint) (model.AnsibleJob, error) {
	var job model.AnsibleJob
	err := database.DB.Preload("Playbook").First(&job, id).Error
	return job, err
}

func ListAnsibleJobsDAO(playbookID uint, limit int) ([]model.AnsibleJob, error) {
	var jobs []model.AnsibleJob
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
