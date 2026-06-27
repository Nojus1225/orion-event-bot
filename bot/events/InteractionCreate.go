package events

import (
	"fmt"
	"strings"

	"orion-event-bot/bot/core"
	"orion-event-bot/bot/handlers"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/nicksnyder/go-i18n/v2/i18n"
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

			if b.ChannelID != "" && event.Channel().ID().String() != b.ChannelID {
				msg := discord.NewMessageCreate().WithEphemeral(true).WithContent("⚠️ Cette commande ne peut être utilisée que dans le salon de l'événement.")
				_ = event.CreateMessage(msg)
				return
			}

			if b.IsMaintenanceMode() && !b.IsAdmin(event.User().ID.String()) {
				loc := core.GetLocalizer(event.Locale().String())
				msg, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "MaintenanceMessage"})
				if msg == "" {
					msg = "The event is complete."
				}
				err := event.CreateMessage(discord.NewMessageCreate().WithEphemeral(true).WithContent(msg))
				if err != nil {
					fmt.Println("Error sending maintenance message:", err)
				}
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

			if b.ChannelID != "" && event.Channel().ID().String() != b.ChannelID {
				msg := discord.NewMessageCreate().WithEphemeral(true).WithContent("⚠️ Cette interaction ne peut être utilisée que dans le salon de l'événement.")
				_ = event.CreateMessage(msg)
				return
			}

			if b.IsMaintenanceMode() && !b.IsAdmin(event.User().ID.String()) {
				loc := core.GetLocalizer(event.Locale().String())
				msg, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "MaintenanceMessage"})
				if msg == "" {
					msg = "L'événement est actuellement en maintenance. Merci de patienter !"
				}
				err := event.CreateMessage(discord.NewMessageCreate().WithEphemeral(true).WithContent(msg))
				if err != nil {
					fmt.Println("Error sending maintenance message:", err)
				}
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

			if b.ChannelID != "" && event.Channel().ID().String() != b.ChannelID {
				msg := discord.NewMessageCreate().WithEphemeral(true).WithContent("⚠️ Ce modal ne peut être soumis que dans le salon de l'événement.")
				_ = event.CreateMessage(msg)
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

			if b.ChannelID != "" && event.Channel().ID().String() != b.ChannelID {
				return
			}

			commandName := event.Data.CommandName
			if cmd, exists := handlers.Commands[commandName]; exists && cmd.ExecuteAutocomplete != nil {
				cmd.ExecuteAutocomplete(b, event)
			}
		},
	})
}
