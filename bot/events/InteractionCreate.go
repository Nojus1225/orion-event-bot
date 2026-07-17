package events

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	"orion-event-bot/bot/core"
	"orion-event-bot/bot/handlers"
	"orion-event-bot/db"

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

			if strings.HasPrefix(customID, "autodrop-claim-") {
				dropIDStr := strings.TrimPrefix(customID, "autodrop-claim-")
				var dropID uint
				if _, err := fmt.Sscanf(dropIDStr, "%d", &dropID); err == nil {
					userID := event.User().ID.String()
					success := db.ClaimDrop(b.DB.GormDB, dropID, userID)

					if !success {
						loc := core.GetLocalizer("en-US")
						tooLate := loc.MustLocalize(&i18n.LocalizeConfig{MessageID: "AutoDropTooLate"})
						_ = event.CreateMessage(discord.NewMessageCreate().WithEphemeral(true).WithContent(tooLate))
						return
					}

					var data db.User
					b.DB.GormDB.FirstOrCreate(&data, db.User{ID: userID})

					random, err := rand.Int(rand.Reader, big.NewInt(1000))
					if err != nil {
						fmt.Println("Error generating random reward:", err)
						return
					}
					roll := int(random.Int64()) + 1

					loc := core.GetLocalizer("en-US")

					var lootMsg string

					if roll <= 450 {
						data.Coins += 3
						b.DB.GormDB.Save(&data)
						lootMsg = loc.MustLocalize(&i18n.LocalizeConfig{MessageID: "AutoDropLootCoinsSmall"})
					} else if roll <= 750 {
						db.IncrementItem(b.DB.GormDB, userID, db.ItemCoquillage, 3)
						db.IncrementItem(b.DB.GormDB, userID, db.ItemCannette, 1)
						lootMsg = loc.MustLocalize(&i18n.LocalizeConfig{MessageID: "AutoDropLootShells"})
					} else if roll <= 900 {
						data.Coins += 12
						b.DB.GormDB.Save(&data)
						lootMsg = loc.MustLocalize(&i18n.LocalizeConfig{MessageID: "AutoDropLootCoinsMedium"})
					} else if roll <= 980 {
						db.IncrementItem(b.DB.GormDB, userID, db.ItemLunettes, 1)
						db.IncrementItem(b.DB.GormDB, userID, db.ItemPince, 2)
						lootMsg = loc.MustLocalize(&i18n.LocalizeConfig{MessageID: "AutoDropLootFirstAid"})
					} else if roll <= 998 {
						data.Coins += 50
						b.DB.GormDB.Save(&data)
						lootMsg = loc.MustLocalize(&i18n.LocalizeConfig{MessageID: "AutoDropLootJackpot"})
					} else {
						db.IncrementItem(b.DB.GormDB, userID, db.ItemMsgBouteille, 1)
						db.IncrementItem(b.DB.GormDB, userID, db.ItemMontre, 1)
						db.IncrementItem(b.DB.GormDB, userID, db.ItemDent, 1)
						lootMsg = loc.MustLocalize(&i18n.LocalizeConfig{MessageID: "AutoDropLootPirate"})
					}

					db.IncrementStat(b.DB.GormDB, userID, db.StatObjetsTrouve, 1)
					db.LogAction(b.DB.GormDB, userID, db.ActionFouille, fmt.Sprintf(`{"result":"autodrop", "loot":"%s"}`, lootMsg))

					claimedTitle := loc.MustLocalize(&i18n.LocalizeConfig{MessageID: "AutoDropClaimedTitle"})
					claimedDesc := loc.MustLocalize(&i18n.LocalizeConfig{
						MessageID: "AutoDropClaimedDescription",
						TemplateData: map[string]interface{}{
							"User": userID,
							"Loot": lootMsg,
						},
					})

					msgUpdate := discord.NewMessageUpdateV2(
						discord.NewContainer(
							discord.NewSection(
								discord.NewTextDisplay(claimedTitle),
								discord.NewTextDisplay(claimedDesc),
							).WithAccessory(discord.NewThumbnail("https://raw.githubusercontent.com/googlefonts/noto-emoji/main/png/512/emoji_u1f381.png")),
						).WithAccentColor(0x2ECC71),
					)

					_ = event.UpdateMessage(msgUpdate)
				}
				return
			}

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
