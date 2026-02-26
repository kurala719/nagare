package repository

import (
	"nagare/internal/database"
	"nagare/internal/model"
)

func AddPacketAnalysisDAO(pa *model.PacketAnalysis) error {
	return database.DB.Create(pa).Error
}

func UpdatePacketAnalysisDAO(pa *model.PacketAnalysis) error {
	return database.DB.Save(pa).Error
}

func GetPacketAnalysisByIDDAO(id uint) (model.PacketAnalysis, error) {
	var pa model.PacketAnalysis
	err := database.DB.First(&pa, id).Error
	return pa, err
}

func GetAllPacketAnalysesDAO(userID uint) ([]model.PacketAnalysis, error) {
	var pas []model.PacketAnalysis
	db := database.DB
	if userID > 0 {
		db = db.Where("user_id = ?", userID)
	}
	err := db.Order("id desc").Find(&pas).Error
	return pas, err
}

func DeletePacketAnalysisDAO(id uint) error {
	return database.DB.Delete(&model.PacketAnalysis{}, id).Error
}
