package core

import (
	"fmt"
	"orion-event-bot/db"

	"github.com/disgoorg/disgo/bot"
)

type Bot struct {
	Client    *bot.Client
	DB        *db.DB
	AdminIDs  []string
	StaffIDs  []string
	GuildID   string
	ChannelID string
}

func (b *Bot) IsMaintenanceMode() bool {
	var event db.Event
	res := b.DB.GormDB.First(&event)
	if res.Error != nil {
		fmt.Println("maintenance check error")
		return false
	}
	return !event.IsActive
}

func (b *Bot) IsAdmin(userID string) bool {
	for _, id := range b.AdminIDs {
		if id == userID {
			return true
		}
	}
	return false
}

func (b *Bot) IsStaff(userID string) bool {
	for _, id := range b.StaffIDs {
		if id == userID {
			return true
		}
	}
	return false
}
