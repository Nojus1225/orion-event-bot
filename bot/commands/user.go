package commands

import (
	"fmt"
	"strings"
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
		Name:        "user",
		Description: "Get user event data",
		Create: discord.SlashCommandCreate{
			Name:        "user",
			Description: "Get user event data",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionUser{
					Name:        "user",
					Description: "The user",
				},
			},
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			loc := core.GetLocalizer(event.Locale().String())

			u := event.SlashCommandInteractionData().User("user")
			if u.ID == 0 {
				u = event.User()
			}

			var user db.User
			if err := b.DB.GormDB.Where("id = ?", u.ID.String()).First(&user).Error; err != nil {
				notParticipatingTxt, _ := loc.Localize(&i18n.LocalizeConfig{
					MessageID:    "UserNotParticipating",
					TemplateData: map[string]interface{}{"User": u.ID.String()},
				})
				msg := discord.NewMessageCreateV2(
					discord.NewContainer(
						discord.NewSection(
							discord.NewTextDisplay("# 👤 User"),
							discord.NewTextDisplay(notParticipatingTxt),
						).WithAccessory(discord.NewThumbnail(u.EffectiveAvatarURL())),
					).WithAccentColor(0xED4245),
				)
				_ = event.CreateMessage(msg)
				return
			}

			var energyResetText string
			now := time.Now()
			energyReset := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
			energyResetText = fmt.Sprintf("<t:%d:R>", energyReset.Unix())

			var invItems []string
			for _, itemID := range user.Inventory {
				if item, ok := vars.Items[vars.ItemID(itemID)]; ok {
					invItems = append(invItems, fmt.Sprintf("%s %s", item.Emoji, item.Name))
				} else {
					invItems = append(invItems, itemID)
				}
			}
			var inventoryText string
			if len(invItems) == 0 {
				inventoryText, _ = loc.Localize(&i18n.LocalizeConfig{
					MessageID: "MeInventoryEmpty",
				})
			} else {
				inventoryText = strings.Join(invItems, ", ")
			}

			meStats, _ := loc.Localize(&i18n.LocalizeConfig{
				MessageID: "MeStats",
				TemplateData: map[string]interface{}{
					"Snowflakes":  user.Snowflakes,
					"Energy":      user.Energy,
					"EnergyReset": energyResetText,
					"Position":    user.Position,
					"Inventory":   inventoryText,
				},
			})

			msg := discord.NewMessageCreateV2(
				discord.NewContainer(
					discord.NewSection(
						discord.NewTextDisplayf("# <@%s>", u.ID.String()),
						discord.NewTextDisplay(meStats),
					).WithAccessory(discord.NewThumbnail(*u.AvatarURL())),
				).WithAccentColor(0x5865F2),
			)

			err := event.CreateMessage(msg)
			if err != nil {
				fmt.Println("Error sending ping message:", err)
			}
		},
	})
}
