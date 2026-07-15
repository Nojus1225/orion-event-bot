package commands

import (
	"fmt"

	"orion-event-bot/bot/core"
	"orion-event-bot/bot/handlers"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

func init() {
	handlers.RegisterCommand(handlers.Command{
		Name:        "event",
		Description: "Information about the summer event",
		Create: discord.SlashCommandCreate{
			Name:        "event",
			Description: "Information about the summer event",

			DescriptionLocalizations: map[discord.Locale]string{
				discord.LocaleFrench: "Informations sur l'événement estival",
			},
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			loc := core.GetLocalizer(event.Locale().String())

			eventTitle, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "EventTitle"})
			eventDesc, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "EventDesc"})

			msg := discord.NewMessageCreateV2(
				discord.NewContainer(
					discord.NewSection(
						discord.NewTextDisplay(eventTitle),
						discord.NewTextDisplay(eventDesc),
					).WithAccessory(discord.NewThumbnail("https://raw.githubusercontent.com/googlefonts/noto-emoji/main/png/512/emoji_u1f3dd.png")),
				).WithAccentColor(0x5865F2),
			)

			if err := event.CreateMessage(msg); err != nil {
				fmt.Println("Error sending event message:", err)
			}
		},
	})
}
