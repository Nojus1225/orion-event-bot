package db

import (
	"gorm.io/gorm"
)

func IncrementItem(db *gorm.DB, userID string, itemID Item, amount int) int {
	var item UserItem
	err := db.Where(UserItem{UserID: userID, ItemID: string(itemID)}).FirstOrCreate(&item).Error
	if err == nil {
		item.Quantity += amount
		db.Save(&item)
		return item.Quantity
	}
	return 0
}

func GetItemQuantity(db *gorm.DB, userID string, itemID Item) int {
	var item UserItem
	err := db.Where("user_id = ? AND item_id = ?", userID, string(itemID)).First(&item).Error
	if err == nil {
		return item.Quantity
	}
	return 0
}

func RemoveItem(db *gorm.DB, userID string, itemID Item, amount int) {
	var item UserItem
	err := db.Where("user_id = ? AND item_id = ?", userID, string(itemID)).First(&item).Error
	if err == nil {
		item.Quantity -= amount
		if item.Quantity <= 0 {
			db.Delete(&item)
		} else {
			db.Save(&item)
		}
	}
}

func IncrementStat(db *gorm.DB, userID string, statName StatName, amount int) {
	var stat UserStat
	err := db.Where(UserStat{UserID: userID, StatName: string(statName)}).FirstOrCreate(&stat).Error
	if err == nil {
		stat.Value += amount
		db.Save(&stat)
	}
}

func GetStat(db *gorm.DB, userID string, statName StatName) int {
	var stat UserStat
	err := db.Where("user_id = ? AND stat_name = ?", userID, string(statName)).First(&stat).Error
	if err == nil {
		return stat.Value
	}
	return 0
}
