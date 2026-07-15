package commands

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"orion-event-bot/bot/core"
	"orion-event-bot/bot/handlers"
	"orion-event-bot/db"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var sellPrices = map[string]int{
	"Shell":               1,
	"Bottle":              1,
	"Can":                 1,
	"Glasses":             4,
	"Coins":               4,
	"Crab Claw":           4,
	"Message in a Bottle": 15,
	"Watch":               15,
	"Shark Tooth":         15,
}

var sellItemsMap = map[string]db.Item{
	"Shell":               db.ItemCoquillage,
	"Bottle":              db.ItemBouteille,
	"Can":                 db.ItemCannette,
	"Glasses":             db.ItemLunettes,
	"Coins":               db.ItemPiece,
	"Crab Claw":           db.ItemPince,
	"Message in a Bottle": db.ItemMsgBouteille,
	"Watch":               db.ItemMontre,
	"Shark Tooth":         db.ItemDent,
}

func getSellMultiplier(tier db.Tier) float64 {
	switch tier {
	case db.TierSac:
		return 1.2
	case db.TierTamis:
		return 1.5
	case db.TierDetecteur:
		return 2.0
	default:
		return 1.0
	}
}

func init() {
	handlers.RegisterCommand(handlers.Command{
		Name:        "sell",
		Description: "Sell your items.",
		Create: discord.SlashCommandCreate{
			Name:        "sell",
			Description: "Sell all your items.",
			NameLocalizations: map[discord.Locale]string{
				discord.LocaleFrench: "vendre",
			},
			DescriptionLocalizations: map[discord.Locale]string{
				discord.LocaleFrench: "Vendre tous vos objets.",
			},
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:         "item",
					Description:  "Sell all stock of this item only",
					NameLocalizations: map[discord.Locale]string{
						discord.LocaleFrench: "objet",
					},
					DescriptionLocalizations: map[discord.Locale]string{
						discord.LocaleFrench: "Vendre tout le stock de cet objet uniquement",
					},
					Required:     false,
					Autocomplete: true,
				},
			},
		},
		ExecuteAutocomplete: func(b *core.Bot, event *events.AutocompleteInteractionCreate) {
			var choices []discord.AutocompleteChoice
			currentItem := ""
			for _, option := range event.Data.Options {
				if option.Name == "item" {
					if len(option.Value) > 0 {
						_ = json.Unmarshal(option.Value, &currentItem)
					}
					break
				}
			}

			var data db.User
			if err := b.DB.GormDB.FirstOrCreate(&data, db.User{ID: strconv.FormatUint(uint64(event.User().ID), 10)}).Error; err != nil {
				fmt.Println("DB error:", err)
			}

			addChoice := func(name string, count int) {
				if count > 0 {
					choiceName := fmt.Sprintf("%s (%d)", name, count)
					if currentItem == "" || strings.Contains(strings.ToLower(name), strings.ToLower(currentItem)) {
						choices = append(choices, discord.AutocompleteChoiceString{
							Name:  choiceName,
							Value: name,
						})
					}
				}
			}

			for name, itemID := range sellItemsMap {
				addChoice(name, db.GetItemQuantity(b.DB.GormDB, data.ID, itemID))
			}

			if err := event.AutocompleteResult(choices); err != nil {
				fmt.Println("Error sending autocomplete message:", err)
			}
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			loc := core.GetLocalizer(event.Locale().String())
			var data db.User
			if err := b.DB.GormDB.FirstOrCreate(&data, db.User{ID: strconv.FormatUint(uint64(event.User().ID), 10)}).Error; err != nil {
				fmt.Println("DB error:", err)
			}

			var targetItem string
			if option, ok := event.SlashCommandInteractionData().OptString("item"); ok {
				targetItem = option
			}

			var content string
			var buttonLabel string
			var customID string

			if targetItem != "" {
				count := 0
				if itemID, ok := sellItemsMap[targetItem]; ok {
					count = db.GetItemQuantity(b.DB.GormDB, data.ID, itemID)
				}

				if count == 0 {
					itemLoc, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: targetItem})
					if itemLoc == "" {
						itemLoc = targetItem
					}
					noItemTxt, _ := loc.Localize(&i18n.LocalizeConfig{
						MessageID:    "SellNoSpecificItem",
						TemplateData: map[string]interface{}{"Item": itemLoc},
					})
					err := event.CreateMessage(discord.NewMessageCreate().WithEphemeral(true).WithContent(noItemTxt))
					if err != nil {
						fmt.Println("Error:", err)
					}
					return
				}

				price := int(float64(count*sellPrices[targetItem]) * getSellMultiplier(data.Tier))
				itemLoc, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: targetItem})
				if itemLoc == "" {
					itemLoc = targetItem
				}
				content, _ = loc.Localize(&i18n.LocalizeConfig{
					MessageID:    "SellConfirm",
					TemplateData: map[string]interface{}{"Quantity": count, "Item": itemLoc, "Coins": price},
				})
				buttonLabel = "Sell " + itemLoc
				customID = "sell-" + targetItem
			} else {
				total := 0
				for name, itemID := range sellItemsMap {
					count := db.GetItemQuantity(b.DB.GormDB, data.ID, itemID)
					total += count * sellPrices[name]
				}

				total = int(float64(total) * getSellMultiplier(data.Tier))

				if total == 0 {
					noItemsTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "SellNoItems"})
					err := event.CreateMessage(discord.NewMessageCreate().WithEphemeral(true).WithContent(noItemsTxt))
					if err != nil {
						fmt.Println("Error:", err)
					}
					return
				}

				content, _ = loc.Localize(&i18n.LocalizeConfig{
					MessageID:    "SellAllConfirm",
					TemplateData: map[string]interface{}{"Coins": total},
				})
				buttonLabel, _ = loc.Localize(&i18n.LocalizeConfig{MessageID: "SellAll"})
				customID = "sell-all"
			}

			msg := discord.NewMessageCreateV2(
				discord.NewContainer(
					discord.NewSection(
						discord.NewTextDisplay(content),
					).WithAccessory(discord.NewThumbnail("https://raw.githubusercontent.com/googlefonts/noto-emoji/main/png/512/emoji_u1f4b0.png")),
					discord.NewActionRow(
						discord.NewButton(discord.ButtonStyleSuccess, buttonLabel, customID, "", snowflake.ID(0)).WithEmoji(discord.NewComponentEmoji("💰")),
					),
				).WithAccentColor(0x5865F2),
			)

			err := event.CreateMessage(msg)
			if err != nil {
				fmt.Println("Error sending sell message:", err)
			}
		},
		ExecuteButton: func(b *core.Bot, event *events.ComponentInteractionCreate) {
			loc := core.GetLocalizer(event.Locale().String())
			notYourComponentTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "NotYourComponent"})
			if (event.Message.InteractionMetadata != nil && event.Message.InteractionMetadata.User.ID != event.User().ID) ||
				(event.Message.Interaction != nil && event.Message.Interaction.User.ID != event.User().ID) {
				_ = event.CreateMessage(discord.NewMessageCreate().WithContent(notYourComponentTxt).WithEphemeral(true))
				return
			}
			customID := event.Data.CustomID()
			parts := strings.SplitN(customID, "-", 2)
			if len(parts) != 2 || parts[0] != "sell" {
				return
			}
			action := parts[1]

			var data db.User
			if err := b.DB.GormDB.FirstOrCreate(&data, db.User{ID: strconv.FormatUint(uint64(event.User().ID), 10)}).Error; err != nil {
				fmt.Println("DB error:", err)
			}

			coinsGenerated := 0

			if action == "all" {
				for name, itemID := range sellItemsMap {
					count := db.GetItemQuantity(b.DB.GormDB, data.ID, itemID)
					if count > 0 {
						coinsGenerated += count * sellPrices[name]
						db.RemoveItem(b.DB.GormDB, data.ID, itemID, count)
					}
				}
				coinsGenerated = int(float64(coinsGenerated) * getSellMultiplier(data.Tier))
			} else {
				targetItem := action
				count := 0
				if itemID, ok := sellItemsMap[targetItem]; ok {
					count = db.GetItemQuantity(b.DB.GormDB, data.ID, itemID)
					if count > 0 {
						db.RemoveItem(b.DB.GormDB, data.ID, itemID, count)
					}
				}
				if count == 0 {
					itemLoc, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: targetItem})
					if itemLoc == "" {
						itemLoc = targetItem
					}
					noItemTxt, _ := loc.Localize(&i18n.LocalizeConfig{
						MessageID:    "SellNoSpecificItem",
						TemplateData: map[string]interface{}{"Item": itemLoc},
					})
					err := event.CreateMessage(discord.NewMessageCreate().WithEphemeral(true).WithContent(noItemTxt))
					if err != nil {
						fmt.Println("Error:", err)
					}
					return
				}
				coinsGenerated = int(float64(count*sellPrices[targetItem]) * getSellMultiplier(data.Tier))
			}

			if coinsGenerated == 0 {
				noItemsTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "SellNoItems"})
				err := event.CreateMessage(discord.NewMessageCreate().WithEphemeral(true).WithContent(noItemsTxt))
				if err != nil {
					fmt.Println("Error:", err)
				}
				return
			}

			data.Coins += coinsGenerated
			if err := b.DB.GormDB.Save(&data).Error; err != nil {
				fmt.Println("DB error:", err)
			}

			db.IncrementStat(b.DB.GormDB, data.ID, db.StatTotalCoins, coinsGenerated)
			db.IncrementStat(b.DB.GormDB, data.ID, db.StatVentesTotal, 1)

			db.LogAction(b.DB.GormDB, data.ID, db.ActionSell, fmt.Sprintf(`{"action":"%s", "coins":%d}`, action, coinsGenerated))

			successMsg, _ := loc.Localize(&i18n.LocalizeConfig{
				MessageID:    "SellSuccessMsg",
				TemplateData: map[string]interface{}{"Coins": coinsGenerated},
			})
			msgUpdate := discord.NewMessageUpdateV2(
				discord.NewContainer(
					discord.NewTextDisplay(successMsg),
				).WithAccentColor(0x57F287),
			)

			if err := event.UpdateMessage(msgUpdate); err != nil {
				fmt.Println("Error updating sell message:", err)
			}
		},
	})
}
