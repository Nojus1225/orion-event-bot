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
					err := event.CreateMessage(discord.NewMessageCreate().WithEphemeral(true).WithContent(fmt.Sprintf(":x: You don't have any '%s' to sell!", targetItem)))
					if err != nil {
						fmt.Println("Error:", err)
					}
					return
				}

				price := count * sellPrices[targetItem]
				content = fmt.Sprintf("You are about to sell **%d** %s for **%d** coins. Click the button to confirm.", count, targetItem, price)
				buttonLabel = "Sell " + targetItem
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

				if total == 0 {
					err := event.CreateMessage(discord.NewMessageCreate().WithEphemeral(true).WithContent(":x: You don't have anything to sell!"))
					if err != nil {
						fmt.Println("Error:", err)
					}
					return
				}

				content = fmt.Sprintf("You are about to sell all your items for **%d** coins. Click the button to confirm.", total)
				buttonLabel = "Sell All"
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
			if (event.Message.InteractionMetadata != nil && event.Message.InteractionMetadata.User.ID != event.User().ID) ||
				(event.Message.Interaction != nil && event.Message.Interaction.User.ID != event.User().ID) {
				_ = event.CreateMessage(discord.NewMessageCreate().WithContent("You cannot interact with components on someone else's command.").WithEphemeral(true))
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
					err := event.CreateMessage(discord.NewMessageCreate().WithEphemeral(true).WithContent(fmt.Sprintf(":x: You don't have any '%s' to sell!", targetItem)))
					if err != nil {
						fmt.Println("Error:", err)
					}
					return
				}
				coinsGenerated = count * sellPrices[targetItem]
			}

			if coinsGenerated == 0 {
				err := event.CreateMessage(discord.NewMessageCreate().WithEphemeral(true).WithContent(":x: You don't have anything to sell!"))
				if err != nil {
					fmt.Println("Error:", err)
				}
				return
			}

			data.Coins += coinsGenerated

			if err := b.DB.GormDB.Save(&data).Error; err != nil {
				fmt.Println("DB error:", err)
			}

			msgUpdate := discord.NewMessageUpdateV2(
				discord.NewContainer(
					discord.NewTextDisplay(fmt.Sprintf(":white_check_mark: Successfully sold your items for **%d** coins!", coinsGenerated)),
				).WithAccentColor(0x57F287),
			)

			if err := event.UpdateMessage(msgUpdate); err != nil {
				fmt.Println("Error updating sell message:", err)
			}
		},
	})
}
