package commands

import (
	"fmt"
	"time"

	"orion-event-bot/bot/core"
	"orion-event-bot/bot/handlers"
	"orion-event-bot/db"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

func init() {
	handlers.RegisterCommand(handlers.Command{
		Name:        "me",
		Description: "Get your event data",
		Create: discord.SlashCommandCreate{
			Name:        "me",
			Description: "Get your event data",
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			loc := core.GetLocalizer(event.Locale().String())
			var user db.User
			if err := b.DB.GormDB.Where("id = ?", event.User().ID.String()).First(&user).Error; err != nil {
				fmt.Println("DB error:", err)
				return
			}

			var energyResetText string
			now := time.Now()
			energyReset := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
			energyResetText = fmt.Sprintf("<t:%d:R>", energyReset.Unix())

			meStats, _ := loc.Localize(&i18n.LocalizeConfig{
				MessageID: "MeStats",
				TemplateData: map[string]interface{}{
					"Snowflakes":  user.Snowflakes,
					"Energy":      user.Energy,
					"EnergyReset": energyResetText,
					"Position":    user.Position,
				},
			})

			msg := discord.NewMessageCreateV2(
				discord.NewContainer(
					discord.NewSection(
						discord.NewTextDisplayf("# <@%s>", event.User().ID.String()),
						discord.NewTextDisplay(meStats),
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
