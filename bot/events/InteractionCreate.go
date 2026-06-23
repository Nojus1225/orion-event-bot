package events

import (
	"orion-event-bot/bot/core"
	"orion-event-bot/bot/handlers"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
)

func init() {
	handlers.RegisterEvent(handlers.Event{
		Name:     "ApplicationCommandInteractionCreate",
		ExecOnce: false,
		Execute: func(b *core.Bot, e bot.Event) {
			event, ok := e.(*events.ApplicationCommandInteractionCreate)
			if !ok {
				return
			}

			commandName := event.Data.CommandName()
			if cmd, exists := handlers.Commands[commandName]; exists {
				cmd.Execute(b, event)
			}
		},
	})
}
