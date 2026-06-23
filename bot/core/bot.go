package core

import (
	"orion-event-bot/db"

	"github.com/disgoorg/disgo/bot"
)

type Bot struct {
	Client *bot.Client
	DB     *db.DB
}
