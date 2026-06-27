package core

import (
	"orion-event-bot/db"

	"github.com/disgoorg/disgo/bot"
)

type Bot struct {
	Client          *bot.Client
	DB              *db.DB
	AdminIDs        []string
	GuildID         string
	ChannelID       string
}

func (b *Bot) IsMaintenanceMode() bool {
	settings := db.GetSettings(b.DB.GormDB)
	return settings.MaintenanceMode
}

func (b *Bot) IsAdmin(userID string) bool {
	for _, id := range b.AdminIDs {
		if id == userID {
			return true
		}
	}
	return false
}
