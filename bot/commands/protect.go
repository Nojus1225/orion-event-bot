package commands

import (
	"fmt"
	"strconv"
	"time"

	"orion-event-bot/bot/core"
	"orion-event-bot/bot/handlers"
	"orion-event-bot/bot/vars"
	"orion-event-bot/db"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

func init() {
	handlers.RegisterCommand(handlers.Command{
		Name:        "protect",
		Description: "Use a protection or utility item",
		Create: discord.SlashCommandCreate{
			Name:        "protect",
			Description: "Use a protection or utility item",
			DescriptionLocalizations: map[discord.Locale]string{
				discord.LocaleFrench: "Utiliser une protection ou un objet",
			},
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:         "item",
					Description:  "The item to use",
					Required:     true,
					Autocomplete: true,
				},
			},
		},
		ExecuteAutocomplete: func(b *core.Bot, event *events.AutocompleteInteractionCreate) {
			loc := core.GetLocalizer(event.Locale().String())
			var choices []discord.AutocompleteChoice

			var data db.User
			if err := b.DB.GormDB.FirstOrCreate(&data, db.User{ID: strconv.FormatUint(uint64(event.User().ID), 10)}).Error; err != nil {
				fmt.Println("DB error:", err)
			}

			for _, option := range event.Data.Options {
				if option.Name == "item" && option.Focused {
					items := []db.Item{
						db.ShopCremeSolaire,
						db.ShopPansement,
						db.ShopBouteilleEau,
						db.ShopDrapeauBaignade,
						db.ShopCremeSolaire1000,
					}
					for _, itemID := range items {
						qty := db.GetItemQuantity(b.DB.GormDB, data.ID, itemID)
						if qty > 0 {
							item := vars.GetItem(itemID)
							if item != nil {
								itemLoc, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: item.Name})
								if itemLoc == "" {
									itemLoc = item.Name
								}
								label := fmt.Sprintf("%s (%d)", itemLoc, qty)
								choices = append(choices, discord.AutocompleteChoiceString{
									Name:  label,
									Value: string(itemID),
								})
							}
						}
					}
				}
			}

			if err := event.AutocompleteResult(choices); err != nil {
				fmt.Println("Error sending autocomplete message:", err)
			}
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			loc := core.GetLocalizer(event.Locale().String())
			cmdData := event.SlashCommandInteractionData()

			userID := strconv.FormatUint(uint64(event.User().ID), 10)

			var data db.User
			if err := b.DB.GormDB.FirstOrCreate(&data, db.User{ID: userID}).Error; err != nil {
				return
			}

			itemType := db.Item(cmdData.String("item"))
			now := time.Now().Unix()

			if db.GetItemQuantity(b.DB.GormDB, userID, itemType) <= 0 {
				msgStr, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "ProtectNoItem"})
				if msgStr == "" {
					msgStr = "Vous ne possédez pas cet objet."
				}
				_ = event.CreateMessage(discord.NewMessageCreate().WithContent(msgStr).WithEphemeral(true))
				return
			}

			used := false
			var responseMessageID string

			switch itemType {
			case db.ShopCremeSolaire:
				if data.StatusMalus == db.MalusSunBurn {
					data.StatusMalus = db.MalusNone
					data.CDMalus = 0
					used = true
					responseMessageID = "ProtectSuccessCreme"
				} else {
					responseMessageID = "ProtectNotNeeded"
				}
			case db.ShopPansement:
				if data.StatusMalus == db.MalusCut {
					data.StatusMalus = db.MalusNone
					data.CDMalus = 0
					used = true
					responseMessageID = "ProtectSuccessPansement"
				} else {
					responseMessageID = "ProtectNotNeeded"
				}
			case db.ShopBouteilleEau:
				if int64(data.CDFouille) > now {
					data.CDFouille = 0
					used = true
					responseMessageID = "ProtectSuccessEau"
				} else {
					responseMessageID = "ProtectNotNeeded"
				}
			case db.ShopDrapeauBaignade:
				data.FlagActiveUntil = now + 4*3600
				used = true
				responseMessageID = "ProtectSuccessDrapeau"
			case db.ShopCremeSolaire1000:
				data.SabotageImmunityExpiresAt = now + 4*3600
				used = true
				responseMessageID = "ProtectSuccessCreme100"
			}

			if used {
				b.DB.GormDB.Save(&data)
				db.RemoveItem(b.DB.GormDB, userID, itemType, 1)
				db.LogAction(b.DB.GormDB, userID, db.ActionProtect, fmt.Sprintf(`{"item":"%s"}`, itemType))
			}

			msgStr, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: responseMessageID})
			if msgStr == "" {
				if responseMessageID == "ProtectNotNeeded" {
					msgStr = "Vous n'avez pas besoin d'utiliser cet objet actuellement."
				} else {
					msgStr = "Objet utilisé avec succès !"
				}
			}

			msg := discord.NewMessageCreateV2(
				discord.NewContainer(
					discord.NewTextDisplay(msgStr),
				).WithAccentColor(0x2ECC71),
			).WithEphemeral(true)

			err := event.CreateMessage(msg)
			if err != nil {
				fmt.Println("Error sending protect message:", err)
			}
		},
	})
}
