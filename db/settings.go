package db

import (
	"gorm.io/gorm"
)

type Settings struct {
	ID              uint `gorm:"primaryKey"`
	MaintenanceMode bool `gorm:"default:false"`
}

func GetSettings(db *gorm.DB) Settings {
	var settings Settings
	db.FirstOrCreate(&settings, Settings{ID: 1})
	return settings
}

func UpdateMaintenanceMode(db *gorm.DB, mode bool) error {
	return db.Model(&Settings{}).Where("id = ?", 1).Update("maintenance_mode", mode).Error
}
