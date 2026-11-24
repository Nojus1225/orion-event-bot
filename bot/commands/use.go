package commands

import (
	"fmt"
	"strings"

	"orion-event-bot/bot/core"
	"orion-event-bot/bot/handlers"
	"orion-event-bot/bot/vars"
	"orion-event-bot/db"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

func fuzzyMatch(text, query string) bool {
	if query == "" {
		return true
	}
	text = strings.ToLower(text)
	query = strings.ToLower(query)
	if strings.Contains(text, query) {
		return true
	}
	qi := 0
	for _, r := range text {
		if qi < len(query) && byte(r) == query[qi] {
			qi++
		}
	}
	return qi == len(query)
}

func init() {
	handlers.RegisterCommand(handlers.Command{
		Name:        "use",
		Description: "Use an object from your inventory",
		Create: discord.SlashCommandCreate{
			Name:        "use",
			Description: "Use an object from your inventory",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:         "object",
					Description:  "The object to use.",
					Required:     true,
					Autocomplete: true,
				},
			},
		},
		ExecuteAutocomplete: func(b *core.Bot, event *events.AutocompleteInteractionCreate) {
			option, exist := event.AutocompleteInteraction.Data.Options["object"]
			if !exist {
				return
			}
			query := strings.TrimSpace(option.String())

			var user db.User
			if err := b.DB.GormDB.Where("id = ?", event.User().ID.String()).First(&user).Error; err != nil {
				_ = event.AutocompleteResult([]discord.AutocompleteChoice{})
				return
			}

			counts := make(map[string]int)
			var order []string
			for _, id := range user.Inventory {
				if counts[id] == 0 {
					order = append(order, id)
				}
				counts[id]++
			}

			var choices []discord.AutocompleteChoice
			for _, id := range order {
				item, ok := vars.Items[vars.ItemID(id)]
				if !ok {
					continue
				}
				if fuzzyMatch(string(item.ID), query) || fuzzyMatch(item.Name, query) || fuzzyMatch(item.Description, query) {
					name := fmt.Sprintf("%s %s (x%d)", item.Emoji, item.Name, counts[id])
					choices = append(choices, discord.AutocompleteChoiceString{
						Name:  name,
						Value: string(item.ID),
					})
				}
				if len(choices) >= 25 {
					break
				}
			}

			_ = event.AutocompleteResult(choices)
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			loc := core.GetLocalizer(event.Locale().Code())
			itemId := event.SlashCommandInteractionData().String("object")
			item, exist := vars.Items[vars.ItemID(itemId)]
			if !exist {
				notFoundTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "UseItemNotFound"})
				msg := discord.NewMessageCreateV2(
					discord.NewContainer(
						discord.NewSection(
							discord.NewTextDisplay("# Winter Rush"),
							discord.NewTextDisplay(notFoundTxt),
						).WithAccessory(discord.NewThumbnail(*event.User().AvatarURL())),
					).WithAccentColor(0xED4245),
				)
				_ = event.CreateMessage(msg)
				return
			}

			var user db.User
			if err := b.DB.GormDB.Where("id = ?", event.User().ID.String()).First(&user).Error; err != nil {
				fmt.Println("DB err:", err)
				return
			}

			var allPlayers []*db.User
			if err := b.DB.GormDB.Find(&allPlayers).Error; err != nil {
				fmt.Println("DB err:", err)
				return
			}

			ctx := vars.ItemContext{
				DB:         b.DB.GormDB,
				User:       &user,
				AllPlayers: allPlayers,
			}

			res, err := item.Use(&ctx)

			var msg discord.MessageCreate

			if err != nil {
				errTxt, _ := loc.Localize(&i18n.LocalizeConfig{
					MessageID:    "UseError",
					TemplateData: map[string]interface{}{"Error": err.Error()},
				})
				msg = discord.NewMessageCreateV2(
					discord.NewContainer(
						discord.NewSection(
							discord.NewTextDisplay("# Winter Rush"),
							discord.NewTextDisplay(errTxt),
						).WithAccessory(discord.NewThumbnail(*event.User().AvatarURL())),
					).WithAccentColor(0xED4245),
				)
			} else {
				for i, id := range user.Inventory {
					if id == itemId {
						user.Inventory = append(user.Inventory[:i], user.Inventory[i+1:]...)
						break
					}
				}
				if err := b.DB.GormDB.Save(&user).Error; err != nil {
					fmt.Println("DB err:", err)
					return
				}

				msg = discord.NewMessageCreateV2(
					discord.NewContainer(
						discord.NewSection(
							discord.NewTextDisplay("# Winter Rush"),
							discord.NewTextDisplay(res),
						).WithAccessory(discord.NewThumbnail(*event.User().AvatarURL())),
					).WithAccentColor(0x5865F2),
				)
			}

			err = event.CreateMessage(msg)
			if err != nil {
				fmt.Println("Error sending ping message:", err)
			}
		},
	})
}
