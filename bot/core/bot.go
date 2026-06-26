package core

import (
	"orion-event-bot/db"

	"github.com/disgoorg/disgo/bot"
)

type Bot struct {
	Client          *bot.Client
	DB              *db.DB
	MaintenanceMode bool
	AdminIDs        []string
}

func (b *Bot) IsAdmin(userID string) bool {
	for _, id := range b.AdminIDs {
		if id == userID {
			return true
		}
	}
	return false
}
