package events

import (
	"fmt"

	"orion-event-bot/bot/core"
	"orion-event-bot/bot/handlers"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
)

func init() {
	handlers.RegisterEvent(handlers.Event{
		Name:     "Ready",
		ExecOnce: true,
		Execute: func(b *core.Bot, e bot.Event) {
			event, ok := e.(*events.Ready)
			if !ok {
				return
			}

			fmt.Println("Bot is connected as", event.User.Username)

			if err := handlers.SyncCommands(b); err != nil {
				fmt.Println("Failed to sync commands:", err)
			} else {
				fmt.Println("Successfully synced", len(handlers.Commands), "commands.")
			}
		},
	})
}
