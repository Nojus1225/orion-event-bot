package commands

import (
	"errors"
	"fmt"

	"orion-event-bot/bot/core"
	"orion-event-bot/bot/handlers"
	"orion-event-bot/db"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"gorm.io/gorm"
)

func init() {
	handlers.RegisterCommand(handlers.Command{
		Name:        "lodge",
		Description: "Put snowflakes in the lodge",
		Create: discord.SlashCommandCreate{
			Name:        "lodge",
			Description: "Put snowflakes in the lodge",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionInt{
					Name:        "snowflakes",
					Description: "Number of snowflakes to put in the lodge",
					Required:    true,
				},
			},
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			loc := core.GetLocalizer(event.Locale().String())

			slash := event.SlashCommandInteractionData()
			nSnowflakes := slash.Int("snowflakes")

			var user db.User
			if err := b.DB.GormDB.Where("id = ?", event.User().ID.String()).First(&user).Error; err != nil {
				fmt.Println("DB error:", err)
				return
			}

			var evt db.Event
			if err := b.DB.GormDB.First(&evt).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				fmt.Println("DB error:", err)
			}

			var msg discord.MessageCreate

			if user.Snowflakes < nSnowflakes {
				notEnoughText, _ := loc.Localize(&i18n.LocalizeConfig{
					MessageID: "LodgeNotEnoughSnowflakes",
				})
				msg = discord.NewMessageCreateV2(
					discord.NewContainer(
						discord.NewSection(
							discord.NewTextDisplay("# Lodge"),
							discord.NewTextDisplay(notEnoughText),
						).WithAccessory(discord.NewThumbnail("https://abs.twimg.com/emoji/v2/72x72/1f3e1.png")),
					).WithAccentColor(0x5865F2),
				)
			} else if (user.LodgeDonatedToday + nSnowflakes) > 25 {
				maxDonationText, _ := loc.Localize(&i18n.LocalizeConfig{
					MessageID: "LodgeMaxDonation",
				})
				msg = discord.NewMessageCreateV2(
					discord.NewContainer(
						discord.NewSection(
							discord.NewTextDisplay("# Lodge"),
							discord.NewTextDisplay(maxDonationText),
						).WithAccessory(discord.NewThumbnail("https://abs.twimg.com/emoji/v2/72x72/1f3e1.png")),
					).WithAccentColor(0x5865F2),
				)
			} else {
				user.LodgeDonatedToday += nSnowflakes
				user.LodgeDonatedTotal += nSnowflakes
				user.Snowflakes -= nSnowflakes

				evt.LodgeTotal += nSnowflakes

				if err := b.DB.GormDB.Save(&user).Error; err != nil {
					fmt.Println("DB error:", err)
					return
				}

				if err := b.DB.GormDB.Save(&evt).Error; err != nil {
					fmt.Println("DB error:", err)
					return
				}

				successText, _ := loc.Localize(&i18n.LocalizeConfig{
					MessageID: "LodgeSuccess",
					TemplateData: map[string]interface{}{
						"Snowflakes": nSnowflakes,
					},
				})
				msg = discord.NewMessageCreateV2(
					discord.NewContainer(
						discord.NewSection(
							discord.NewTextDisplay("# Lodge"),
							discord.NewTextDisplay(successText),
						).WithAccessory(discord.NewThumbnail("https://abs.twimg.com/emoji/v2/72x72/1f3e1.png")),
					).WithAccentColor(0x5865F2),
				)
			}

			err := event.CreateMessage(msg)
			if err != nil {
				fmt.Println("Error sending ping message:", err)
			}
		},
	})
}
