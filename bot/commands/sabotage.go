package commands

import (
	"fmt"
	"strconv"

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
		Name:        "sabotage",
		Description: "Sabotage a user/zone",
		Create: discord.SlashCommandCreate{
			Name:        "sabotage",
			Description: "Sabotage a user/zone",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionSubCommand{
					Name:        "user",
					Description: "Sabotage a user",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionUser{
							Name:        "user",
							Description: "The targeted user",
							Required:    true,
						},
						discord.ApplicationCommandOptionString{
							Name:         "sabotage",
							Description:  "The sabotage to be inflicted",
							Required:     true,
							Autocomplete: true,
						},
					},
				},
				discord.ApplicationCommandOptionSubCommand{
					Name:        "zone",
					Description: "Sabotage a zone",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionString{
							Name:         "zone",
							Description:  "The targeted zone",
							Required:     true,
							Autocomplete: true,
						},
						discord.ApplicationCommandOptionString{
							Name:         "sabotage",
							Description:  "The sabotage to be inflicted",
							Required:     true,
							Autocomplete: true,
						},
					},
				},
			},
		},
		ExecuteAutocomplete: func(b *core.Bot, event *events.AutocompleteInteractionCreate) {
			loc := core.GetLocalizer(event.Locale().String())
			var choices []discord.AutocompleteChoice

			subCommand := event.Data.SubCommandName
			if subCommand == nil {
				return
			}
			
			var data db.User
			if err := b.DB.GormDB.FirstOrCreate(&data, db.User{ID: strconv.FormatUint(uint64(event.User().ID), 10)}).Error; err != nil {
				fmt.Println("DB error:", err)
			}

			switch *subCommand {
			case "user":
				for _, option := range event.Data.Options {
					if option.Name == "sabotage" && option.Focused {
						items := []db.Item{db.ShopVolDePelle, db.ShopCrabeDePoche}
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
			case "zone":
				for _, option := range event.Data.Options {
					if option.Name == "zone" && option.Focused {
						zones := []string{"Deckchairs", "WatersEdge", "TheRocks"}
						for _, zone := range zones {
							zoneLoc, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: zone})
							if zoneLoc == "" {
								zoneLoc = zone
							}
							choices = append(choices, discord.AutocompleteChoiceString{
								Name:  zoneLoc,
								Value: zone,
							})
						}
					} else if option.Name == "sabotage" && option.Focused {
						items := []db.Item{db.ShopAlerteMeduses}
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
			}

			if err := event.AutocompleteResult(choices); err != nil {
				fmt.Println("Error sending autocomplete message:", err)
			}
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {

			msg := discord.NewMessageCreateV2(
				discord.NewContainer(
					discord.NewTextDisplay("oui"),
				).WithAccentColor(0x5865F2),
			)

			err := event.CreateMessage(msg)
			if err != nil {
				fmt.Println("Error sending ping message:", err)
			}
		},
	})
}
