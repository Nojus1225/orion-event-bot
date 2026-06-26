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
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:         "item",
					Description:  "Sell all stock of this item only",
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

			addChoice("Shell", data.ItemCoquillage)
			addChoice("Bottle", data.ItemBouteille)
			addChoice("Can", data.ItemCannette)
			addChoice("Glasses", data.ItemLunettes)
			addChoice("Coins", data.ItemPiece)
			addChoice("Crab Claw", data.ItemPince)
			addChoice("Message in a Bottle", data.ItemMsgBouteille)
			addChoice("Watch", data.ItemMontre)
			addChoice("Shark Tooth", data.ItemDent)

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
				switch targetItem {
				case "Shell":
					count = data.ItemCoquillage
				case "Bottle":
					count = data.ItemBouteille
				case "Can":
					count = data.ItemCannette
				case "Glasses":
					count = data.ItemLunettes
				case "Coins":
					count = data.ItemPiece
				case "Crab Claw":
					count = data.ItemPince
				case "Message in a Bottle":
					count = data.ItemMsgBouteille
				case "Watch":
					count = data.ItemMontre
				case "Shark Tooth":
					count = data.ItemDent
				}

				if count == 0 {
					itemLoc, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: targetItem})
					if itemLoc == "" {
						itemLoc = targetItem
					}
					noItemTxt, _ := loc.Localize(&i18n.LocalizeConfig{
						MessageID: "SellNoSpecificItem",
						TemplateData: map[string]interface{}{"Item": itemLoc},
					})
					err := event.CreateMessage(discord.NewMessageCreate().WithEphemeral(true).WithContent(noItemTxt))
					if err != nil {
						fmt.Println("Error:", err)
					}
					return
				}

				price := int(float64(count * sellPrices[targetItem]) * getSellMultiplier(data.Tier))
				itemLoc, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: targetItem})
				if itemLoc == "" {
					itemLoc = targetItem
				}
				content, _ = loc.Localize(&i18n.LocalizeConfig{
					MessageID: "SellConfirm",
					TemplateData: map[string]interface{}{"Quantity": count, "Item": itemLoc, "Coins": price},
				})
				buttonLabel = "Sell " + itemLoc
				customID = "sell-" + targetItem
			} else {
				total := 0
				total += data.ItemCoquillage * sellPrices["Shell"]
				total += data.ItemBouteille * sellPrices["Bottle"]
				total += data.ItemCannette * sellPrices["Can"]
				total += data.ItemLunettes * sellPrices["Glasses"]
				total += data.ItemPiece * sellPrices["Coins"]
				total += data.ItemPince * sellPrices["Crab Claw"]
				total += data.ItemMsgBouteille * sellPrices["Message in a Bottle"]
				total += data.ItemMontre * sellPrices["Watch"]
				total += data.ItemDent * sellPrices["Shark Tooth"]

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
					MessageID: "SellAllConfirm",
					TemplateData: map[string]interface{}{"Coins": total},
				})
				buttonLabel, _ = loc.Localize(&i18n.LocalizeConfig{MessageID: "SellAll"})
				customID = "sell-all"
			}

			msg := discord.NewMessageCreateV2(
				discord.NewContainer(
					discord.NewTextDisplay(content),
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
				coinsGenerated += data.ItemCoquillage * sellPrices["Shell"]
				data.ItemCoquillage = 0
				coinsGenerated += data.ItemBouteille * sellPrices["Bottle"]
				data.ItemBouteille = 0
				coinsGenerated += data.ItemCannette * sellPrices["Can"]
				data.ItemCannette = 0
				coinsGenerated += data.ItemLunettes * sellPrices["Glasses"]
				data.ItemLunettes = 0
				coinsGenerated += data.ItemPiece * sellPrices["Coins"]
				data.ItemPiece = 0
				coinsGenerated += data.ItemPince * sellPrices["Crab Claw"]
				data.ItemPince = 0
				coinsGenerated += data.ItemMsgBouteille * sellPrices["Message in a Bottle"]
				data.ItemMsgBouteille = 0
				coinsGenerated += data.ItemMontre * sellPrices["Watch"]
				data.ItemMontre = 0
				coinsGenerated += data.ItemDent * sellPrices["Shark Tooth"]
				data.ItemDent = 0

				coinsGenerated = int(float64(coinsGenerated) * getSellMultiplier(data.Tier))
			} else {
				targetItem := action
				count := 0
				switch targetItem {
				case "Shell":
					count = data.ItemCoquillage
					data.ItemCoquillage = 0
				case "Bottle":
					count = data.ItemBouteille
					data.ItemBouteille = 0
				case "Can":
					count = data.ItemCannette
					data.ItemCannette = 0
				case "Glasses":
					count = data.ItemLunettes
					data.ItemLunettes = 0
				case "Coins":
					count = data.ItemPiece
					data.ItemPiece = 0
				case "Crab Claw":
					count = data.ItemPince
					data.ItemPince = 0
				case "Message in a Bottle":
					count = data.ItemMsgBouteille
					data.ItemMsgBouteille = 0
				case "Watch":
					count = data.ItemMontre
					data.ItemMontre = 0
				case "Shark Tooth":
					count = data.ItemDent
					data.ItemDent = 0
				}
				if count == 0 {
					itemLoc, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: targetItem})
					if itemLoc == "" {
						itemLoc = targetItem
					}
					noItemTxt, _ := loc.Localize(&i18n.LocalizeConfig{
						MessageID: "SellNoSpecificItem",
						TemplateData: map[string]interface{}{"Item": itemLoc},
					})
					err := event.CreateMessage(discord.NewMessageCreate().WithEphemeral(true).WithContent(noItemTxt))
					if err != nil {
						fmt.Println("Error:", err)
					}
					return
				}
				coinsGenerated = int(float64(count * sellPrices[targetItem]) * getSellMultiplier(data.Tier))
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
			data.StatsTotalCoins += coinsGenerated

			if err := b.DB.GormDB.Save(&data).Error; err != nil {
				fmt.Println("DB error:", err)
			}

			successMsg, _ := loc.Localize(&i18n.LocalizeConfig{
				MessageID: "SellSuccessMsg",
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
