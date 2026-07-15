package core

import (
	"fmt"
	"math/rand"
	"time"

	"orion-event-bot/db"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

func StartAutoDropTicker(b *Bot) {
	TriggerAutoDrop(b)
	go func() {
		for {
			delay := time.Duration(4)*time.Hour + time.Duration(rand.Intn(4*60))*time.Minute
			time.Sleep(delay)

			settings := db.GetSettings(b.DB.GormDB)
			if settings.AutoDropEnabled {
			}
		}
	}()
}

func TriggerAutoDrop(b *Bot) {
	if b.ChannelID == "" {
		fmt.Println("No ChannelID configured, skipping autodrop")
		return
	}

	channelID, err := snowflake.Parse(b.ChannelID)
	if err != nil {
		fmt.Println("Invalid ChannelID for autodrop:", err)
		return
	}

	dropID, err := db.CreateAutoDrop(b.DB.GormDB)
	if err != nil {
		fmt.Println("Error creating autodrop in DB:", err)
		return
	}

	loc := GetLocalizer("fr-FR")
	title := loc.MustLocalize(&i18n.LocalizeConfig{MessageID: "AutoDropTitle"})
	desc := loc.MustLocalize(&i18n.LocalizeConfig{MessageID: "AutoDropDescription"})
	btnText := loc.MustLocalize(&i18n.LocalizeConfig{MessageID: "AutoDropClaimButton"})

	msg := discord.NewMessageCreateV2(
		discord.NewContainer(
			discord.NewSection(
				discord.NewTextDisplay(title),
				discord.NewTextDisplay(desc),
			).WithAccessory(discord.NewThumbnail("https://raw.githubusercontent.com/googlefonts/noto-emoji/main/png/512/emoji_u1f381.png")),
			discord.NewActionRow(
				discord.NewButton(discord.ButtonStyleSuccess, btnText, fmt.Sprintf("autodrop-claim-%d", dropID), "", snowflake.ID(0)),
			),
		).WithAccentColor(0xF1C40F),
	)

	_, err = b.Client.Rest.CreateMessage(channelID, msg)
	if err != nil {
		fmt.Println("Error sending autodrop message:", err)
	}
}
