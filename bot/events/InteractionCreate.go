package events

import (
	"strings"

	"orion-event-bot/bot/core"
	"orion-event-bot/bot/handlers"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
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
			if cmd, exists := handlers.Commands[commandName]; exists && cmd.Execute != nil {
				cmd.Execute(b, event)
			}
		},
	})

	handlers.RegisterEvent(handlers.Event{
		Name:     "ComponentInteractionCreate",
		ExecOnce: false,
		Execute: func(b *core.Bot, e bot.Event) {
			event, ok := e.(*events.ComponentInteractionCreate)
			if !ok {
				return
			}

			customID := event.Data.CustomID()
			parts := strings.SplitN(customID, "-", 2)
			if len(parts) == 0 {
				return
			}
			commandName := parts[0]

			if cmd, exists := handlers.Commands[commandName]; exists {
				switch event.Data.Type() {
				case discord.ComponentTypeButton:
					if cmd.ExecuteButton != nil {
						cmd.ExecuteButton(b, event)
					}
				case discord.ComponentTypeStringSelectMenu, discord.ComponentTypeUserSelectMenu, discord.ComponentTypeRoleSelectMenu, discord.ComponentTypeMentionableSelectMenu, discord.ComponentTypeChannelSelectMenu:
					if cmd.ExecuteMenu != nil {
						cmd.ExecuteMenu(b, event)
					}
				}
			}
		},
	})

	handlers.RegisterEvent(handlers.Event{
		Name:     "ModalSubmitInteractionCreate",
		ExecOnce: false,
		Execute: func(b *core.Bot, e bot.Event) {
			event, ok := e.(*events.ModalSubmitInteractionCreate)
			if !ok {
				return
			}

			customID := event.Data.CustomID
			parts := strings.SplitN(customID, "-", 2)
			if len(parts) == 0 {
				return
			}
			commandName := parts[0]

			if cmd, exists := handlers.Commands[commandName]; exists && cmd.ExecuteModal != nil {
				cmd.ExecuteModal(b, event)
			}
		},
	})

	handlers.RegisterEvent(handlers.Event{
		Name:     "AutocompleteInteractionCreate",
		ExecOnce: false,
		Execute: func(b *core.Bot, e bot.Event) {
			event, ok := e.(*events.AutocompleteInteractionCreate)
			if !ok {
				return
			}

			commandName := event.Data.CommandName
			if cmd, exists := handlers.Commands[commandName]; exists && cmd.ExecuteAutocomplete != nil {
				cmd.ExecuteAutocomplete(b, event)
			}
		},
	})
}
