package db

import (
	"time"

	"gorm.io/gorm"
)

type ActionType string

const (
	ActionFouille ActionType = "fouille"
	ActionSell    ActionType = "sell"
	ActionShop    ActionType = "shop"
	ActionUpgrade ActionType = "upgrade"
	ActionMalus   ActionType = "malus"
)

type ActionLog struct {
	ID         uint `gorm:"primaryKey"`
	CreatedAt  time.Time
	UserID     string
	ActionType ActionType
	Details    string
}

func LogAction(db *gorm.DB, userID string, actionType ActionType, details string) {
	log := ActionLog{
		UserID:     userID,
		ActionType: actionType,
		Details:    details,
	}
	db.Create(&log)
}
