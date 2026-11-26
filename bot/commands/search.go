package commands

import (
	"fmt"
	"math/rand"

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
		Name:        "search",
		Description: "Search in the snow",
		Create: discord.SlashCommandCreate{
			Name:        "search",
			Description: "Search in the snow",
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			loc := core.GetLocalizer(event.Locale().Code())

			var user db.User
			if err := b.DB.GormDB.Where("id = ?", event.User().ID.String()).First(&user).Error; err != nil {
				fmt.Println("DB err:", err)
				return
			}

			if user.Energy <= 0 {
				noEnergyText, _ := loc.Localize(&i18n.LocalizeConfig{
					MessageID: "SearchNoEnergy",
				})
				msg := discord.NewMessageCreateV2(
					discord.NewContainer(
						discord.NewSection(
							discord.NewTextDisplay(func() string { t, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "SearchTitle"}); return t }()),
							discord.NewTextDisplay(noEnergyText),
						).WithAccessory(discord.NewThumbnail(*event.User().AvatarURL())),
					).WithAccentColor(0xED4245),
				)
				if err := event.CreateMessage(msg); err != nil {
					fmt.Println("Error sending message:", err)
				}
				return
			}

			var message string

			if n := rand.Intn(10); n < 6 {
				nSnowflakes := rand.Intn(25) + 15
				user.Snowflakes += nSnowflakes
				message, _ = loc.Localize(&i18n.LocalizeConfig{
					MessageID: "SearchFoundSnowflakes",
					TemplateData: map[string]interface{}{
						"Snowflakes": nSnowflakes,
					},
				})
			} else {
				if len(user.Inventory) >= 2 {
					user.Snowflakes += 25
					message, _ = loc.Localize(&i18n.LocalizeConfig{
						MessageID: "SearchInventoryFull",
					})
				} else {
					rItem := vars.RollRandomItem()

					user.Inventory = append(user.Inventory, string(rItem.ID))
					message, _ = loc.Localize(&i18n.LocalizeConfig{
						MessageID: "SearchFoundItem",
						TemplateData: map[string]interface{}{
							"Item": rItem.Name,
						},
					})
				}
			}

			user.Energy -= 1

			if err := b.DB.GormDB.Save(&user).Error; err != nil {
				fmt.Println("DB error:", err)
				return
			}

			msg := discord.NewMessageCreateV2(
				discord.NewContainer(
					discord.NewSection(
						discord.NewTextDisplay(func() string { t, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "SearchTitle"}); return t }()),
						discord.NewTextDisplay(message),
					).WithAccessory(discord.NewThumbnail(*event.User().AvatarURL())),
				).WithAccentColor(0x5865F2),
			)

			err := event.CreateMessage(msg)
			if err != nil {
				fmt.Println("Error sending ping message:", err)
			}
		},
	})
}
