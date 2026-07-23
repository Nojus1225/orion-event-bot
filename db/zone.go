package db

import "gorm.io/gorm"

type Zone struct {
	Name              string `gorm:"primaryKey;type:varchar(50)"`
	SabotageExpiresAt int64
	SabotageType      string `gorm:"type:varchar(50)"`
	ImmunityExpiresAt int64
	SaboteurID        string `gorm:"type:varchar(255)"`
}

func GetZone(db *gorm.DB, zoneName string) Zone {
	var zone Zone
	db.Where("name = ?", zoneName).FirstOrCreate(&zone, Zone{Name: zoneName})
	return zone
}

func SetZoneSabotage(db *gorm.DB, zoneName string, sabotageType string, expiresAt int64, immunityExpiresAt int64, saboteurID string) {
	zone := GetZone(db, zoneName)
	zone.SabotageType = sabotageType
	zone.SabotageExpiresAt = expiresAt
	zone.ImmunityExpiresAt = immunityExpiresAt
	zone.SaboteurID = saboteurID
	db.Save(&zone)
}
