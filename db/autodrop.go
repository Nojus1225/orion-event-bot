package db

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

type AutoDrop struct {
	ID        uint         `gorm:"primaryKey;autoIncrement"`
	MessageID string       `gorm:"index"`
	ClaimedBy *string      `gorm:"default:null"`
	ClaimedAt sql.NullTime
	CreatedAt time.Time
}

func CreateAutoDrop(db *gorm.DB) (uint, error) {
	drop := AutoDrop{
		CreatedAt: time.Now(),
	}
	err := db.Create(&drop).Error
	return drop.ID, err
}

func ClaimDrop(db *gorm.DB, dropID uint, userID string) bool {
	res := db.Model(&AutoDrop{}).
		Where("id = ? AND claimed_by IS NULL", dropID).
		Updates(map[string]interface{}{
			"claimed_by": userID,
			"claimed_at": time.Now(),
		})

	if res.Error != nil {
		return false
	}
	return res.RowsAffected > 0
}
