package commands

import (
	"fmt"
	"time"

	"orion-event-bot/bot/core"
	"orion-event-bot/bot/handlers"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func init() {
	handlers.RegisterCommand(handlers.Command{
		Name:        "ping",
		Description: "Replies with pong",
		Create: discord.SlashCommandCreate{
			Name:        "ping",
			Description: "Replies with pong",
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			wsPing := b.Client.Gateway.Latency().Milliseconds()
			dbPing := 0

			botPing := time.Since(event.ID().Time()).Milliseconds()
			avatarURL := event.User().EffectiveAvatarURL()

			msg := discord.NewMessageCreateV2(
				discord.NewContainer(
					discord.NewSection(
						discord.NewTextDisplay("# Orion Hosting"),
						discord.NewTextDisplay(fmt.Sprintf("**Bot:** %dms\n**WS:** %dms\n**DB:** %dms", botPing, wsPing, dbPing)),
					).WithAccessory(discord.NewThumbnail(avatarURL)),
				).WithAccentColor(0x5865F2),
			)

			err := event.CreateMessage(msg)
			if err != nil {
				fmt.Println("Error sending ping message:", err)
			}
		},
	})
}
