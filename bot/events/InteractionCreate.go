package events

import (
	"fmt"
	"strings"
	"time"

	"orion-event-bot/bot/core"
	"orion-event-bot/bot/handlers"
	"orion-event-bot/db"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"gorm.io/gorm/clause"
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

			if b.ChannelID != "" && event.Channel().ID().String() != b.ChannelID && !b.IsAdmin(event.User().ID.String()) {
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

			var user db.User
			userID := event.User().ID.String()
			now := time.Now().UTC()

			if err := b.DB.GormDB.Clauses(clause.OnConflict{DoNothing: true}).Create(&db.User{
				ID:              userID,
				Position:        1,
				Energy:          3,
				LastEnergyReset: now,
			}).Error; err != nil {
				fmt.Println("DB error:", err)
				return
			}
			if err := b.DB.GormDB.Where("id = ?", userID).First(&user).Error; err != nil {
				fmt.Println("DB error:", err)
				return
			}

			lastResetDay := user.LastEnergyReset.UTC().Format("2006-01-02")
			today := now.Format("2006-01-02")

			if lastResetDay != today {
				user.Energy = 3
				user.LodgeDonatedToday = 0
				user.LastEnergyReset = now

				if err := b.DB.GormDB.Save(&user).Error; err != nil {
					fmt.Println("Save reset error:", err)
					return
				}
			}

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

			if b.ChannelID != "" && event.Channel().ID().String() != b.ChannelID && !b.IsAdmin(event.User().ID.String()) {
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

			if b.ChannelID != "" && event.Channel().ID().String() != b.ChannelID && !b.IsAdmin(event.User().ID.String()) {
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

			if b.ChannelID != "" && event.Channel().ID().String() != b.ChannelID && !b.IsAdmin(event.User().ID.String()) {
				return
			}

			commandName := event.Data.CommandName
			if cmd, exists := handlers.Commands[commandName]; exists && cmd.ExecuteAutocomplete != nil {
				cmd.ExecuteAutocomplete(b, event)
			}
		},
	})
}
