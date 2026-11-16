package events

import (
	"encoding/json"
	"regexp"
	"strings"

	"orion-event-bot/bot/core"
	"orion-event-bot/bot/handlers"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

var mentionRegex = regexp.MustCompile(`<@!?(\d+)>`)

func init() {
	handlers.RegisterEvent(handlers.Event{
		Name:     "MessageCreate",
		ExecOnce: false,
		Execute: func(b *core.Bot, e bot.Event) {
			event, ok := e.(*events.MessageCreate)
			if !ok {
				return
			}

			if event.Message.Author.Bot {
				return
			}

			content := event.Message.Content

			if !strings.HasPrefix(content, ",") {
				return
			}

			if !b.IsAdmin(event.Message.Author.ID.String()) {
				return
			}

			parts := strings.Split(content, " ")
			command := strings.ToLower(strings.TrimPrefix(parts[0], ","))
			args := parts[1:]

			switch command {
			case "sql":
				if len(args) == 0 {
					return
				}
				query := strings.Join(args, " ")
				var results []map[string]interface{}
				if err := b.DB.GormDB.Raw(query).Scan(&results).Error; err != nil {
					_, _ = event.Client().Rest.CreateMessage(event.ChannelID, discord.NewMessageCreate().WithContent("Error: "+err.Error()))
					return
				}

				if len(results) == 0 {
					_, _ = event.Client().Rest.CreateMessage(event.ChannelID, discord.NewMessageCreate().WithContent("Success, no results."))
					return
				}

				j, err := json.MarshalIndent(results, "", "  ")
				if err != nil {
					_, _ = event.Client().Rest.CreateMessage(event.ChannelID, discord.NewMessageCreate().WithContent("Error parsing results: "+err.Error()))
					return
				}

				msgStr := string(j)
				if len(msgStr) > 1900 {
					msgStr = msgStr[:1900] + "\n... (truncated)"
				}
				_, _ = event.Client().Rest.CreateMessage(event.ChannelID, discord.NewMessageCreate().WithContent("```json\n"+msgStr+"\n```"))
			}
		},
	})
}
