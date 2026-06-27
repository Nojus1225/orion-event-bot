package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"orion-event-bot/bot/core"
	"orion-event-bot/bot/handlers"
	"orion-event-bot/db"

	_ "orion-event-bot/bot/commands"
	_ "orion-event-bot/bot/events"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/rest"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	core.InitI18n()

	token := os.Getenv("DISCORD_TOKEN")

	database := db.New()

	adminIDsStr := os.Getenv("ADMIN_IDS")
	var adminIDs []string
	if adminIDsStr != "" {
		adminIDs = strings.Split(adminIDsStr, ",")
		for i := range adminIDs {
			adminIDs[i] = strings.TrimSpace(adminIDs[i])
		}
	}

	guildID := os.Getenv("GUILD_ID")
	channelID := os.Getenv("CHANNEL_ID")

	coreBot := &core.Bot{
		DB:       database,
		AdminIDs: adminIDs,
		GuildID:  guildID,
		ChannelID: channelID,
	}

	client, err := disgo.New(token,
		bot.WithGatewayConfigOpts(
			gateway.WithIntents(
				gateway.IntentGuilds,
				gateway.IntentGuildMessages,
				gateway.IntentDirectMessages,
				gateway.IntentMessageContent,
			),
		),
		bot.WithRestConfigOpts(
			rest.WithDefaultAllowedMentions(discord.AllowedMentions{
				Parse: []discord.AllowedMentionType{},
			}),
		),
	)
	if err != nil {
		panic(err)
	}

	coreBot.Client = client

	handlers.SetupEvents(coreBot)
	core.StartAutoDropTicker(coreBot)

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		client.Close(ctx)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err = client.OpenGateway(ctx); err != nil {
		panic(err)
	}

	fmt.Println("Bot is now running. Press CTRL+C to exit.")
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	<-s
	fmt.Println("Shutting down bot...")
}
