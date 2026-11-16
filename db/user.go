package db

import "time"

const MaxInventoryItems = 2

type User struct {
	ID         string `gorm:"primaryKey;type:varchar(255)"`
	Position   int
	Snowflakes int

	LodgeDonatedTotal int `gorm:"check:lodge_donated_today >= 0 AND lodge_donated_today <= 125"`
	LodgeDonatedToday int `gorm:"check:lodge_donated_today >= 0 AND lodge_donated_today <= 25"`

	Energy          int
	LastEnergyReset time.Time

	Inventory []string

	DebuffChairlift bool `gorm:"default:false"`
	BuffHelmet      bool `gorm:"default:false"`
	BuffWax         bool `gorm:"default:false"`

	FinishedAt time.Time
}

func (u *User) AddInventory(item string) bool {
	if len(u.Inventory) >= MaxInventoryItems {
		return false
	}

	u.Inventory = append(u.Inventory, item)
	return true
}
