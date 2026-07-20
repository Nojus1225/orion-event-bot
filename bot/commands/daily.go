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

type DailyReward struct {
	Item     db.Item
	Quantity int
}

var dailyItems = map[string]DailyReward{
	"2026-07-22": {Item: db.ItemCoquillage, Quantity: 3},
	"2026-07-23": {Item: db.ItemBouteille, Quantity: 5},
	"2026-07-24": {Item: db.ItemCannette, Quantity: 8},
	"2026-07-25": {Item: db.ItemLunettes, Quantity: 3},
	"2026-07-26": {Item: db.ItemPiece, Quantity: 4},
	"2026-07-27": {Item: db.ItemPince, Quantity: 5},
	"2026-07-28": {Item: db.ItemMsgBouteille, Quantity: 2},
	"2026-07-29": {Item: db.ItemMontre, Quantity: 3},
}

func getItemTranslationKey(item db.Item) string {
	switch item {
	case db.ItemCoquillage:
		return "Shell"
	case db.ItemBouteille:
		return "Bottle"
	case db.ItemCannette:
		return "Can"
	case db.ItemLunettes:
		return "Glasses"
	case db.ItemPiece:
		return "Coins"
	case db.ItemPince:
		return "CrabClaw"
	case db.ItemMsgBouteille:
		return "MessageInABottle"
	case db.ItemMontre:
		return "Watch"
	case db.ItemDent:
		return "SharkTooth"
	default:
		return string(item)
	}
}

func getNextDailyRewardTime(today time.Time) (time.Time, bool) {
	var nextTime time.Time
	hasGroup := false
	for dateStr := range dailyItems {
		t, err := time.ParseInLocation("2006-01-02", dateStr, today.Location())
		if err != nil {
			continue
		}
		if t.After(today) {
			if !hasGroup || t.Before(nextTime) {
				nextTime = t
				hasGroup = true
			}
		}
	}
	return nextTime, hasGroup
}

func init() {
	handlers.RegisterCommand(handlers.Command{
		Name:        "daily",
		Description: "Get your daily item",
		Create: discord.SlashCommandCreate{
			Name:        "daily",
			Description: "Get your daily item",
			DescriptionLocalizations: map[discord.Locale]string{
				discord.LocaleFrench: "Récupérer la récompense quotidienne",
			},
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			loc := core.GetLocalizer(event.Locale().String())
			nowRaw := time.Now().Local()
			today := time.Date(nowRaw.Year(), nowRaw.Month(), nowRaw.Day(), 0, 0, 0, 0, nowRaw.Location())
			dateStr := today.Format("2006-01-02")

			dailyTitle, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "DailyTitle"})

			reward, hasReward := dailyItems[dateStr]
			if !hasReward {
				noRewardMsg, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "DailyNoReward"})
				msgContent := noRewardMsg

				if nextTime, ok := getNextDailyRewardTime(today); ok {
					nextMsg, _ := loc.Localize(&i18n.LocalizeConfig{
						MessageID: "DailyNextReward",
						TemplateData: map[string]interface{}{
							"Time": nextTime.Unix(),
						},
					})
					msgContent = fmt.Sprintf("%s\n%s", noRewardMsg, nextMsg)
				}

				msg := discord.NewMessageCreateV2(
					discord.NewContainer(
						discord.NewSection(
							discord.NewTextDisplay(dailyTitle),
							discord.NewTextDisplay(msgContent),
						).WithAccessory(discord.NewThumbnail("https://raw.githubusercontent.com/googlefonts/noto-emoji/main/png/512/emoji_u1f3dd.png")),
					).WithAccentColor(0xE74C3C),
				)
				_ = event.CreateMessage(msg)
				return
			}

			var user db.User
			if err := b.DB.GormDB.FirstOrCreate(&user, db.User{ID: event.User().ID.String()}).Error; err != nil {
				fmt.Println("DB error initializing user:", err)
			}

			var dailyRecord db.UserDaily
			err := b.DB.GormDB.Where("user_id = ? AND date = ?", event.User().ID.String(), dateStr).First(&dailyRecord).Error
			if err == nil {
				alreadyClaimedMsg, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "DailyAlreadyClaimed"})
				msg := discord.NewMessageCreateV2(
					discord.NewContainer(
						discord.NewSection(
							discord.NewTextDisplay(dailyTitle),
							discord.NewTextDisplay(alreadyClaimedMsg),
						).WithAccessory(discord.NewThumbnail("https://raw.githubusercontent.com/googlefonts/noto-emoji/main/png/512/emoji_u1f3dd.png")),
					).WithAccentColor(0xE74C3C),
				)
				_ = event.CreateMessage(msg)
				return
			}

			newRecord := db.UserDaily{
				UserID: event.User().ID.String(),
				Date:   dateStr,
			}
			if err := b.DB.GormDB.Create(&newRecord).Error; err != nil {
				fmt.Println("DB error inserting user_daily:", err)
				return
			}

			total := db.IncrementItem(b.DB.GormDB, event.User().ID.String(), reward.Item, reward.Quantity)

			itemKey := getItemTranslationKey(reward.Item)
			itemLoc, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: itemKey})
			if itemLoc == "" {
				itemLoc = string(reward.Item)
			}

			successMsg, _ := loc.Localize(&i18n.LocalizeConfig{
				MessageID: "DailySuccess",
				TemplateData: map[string]interface{}{
					"Quantity": reward.Quantity,
					"Item":     itemLoc,
					"Total":    total,
				},
			})

			msg := discord.NewMessageCreateV2(
				discord.NewContainer(
					discord.NewSection(
						discord.NewTextDisplay(dailyTitle),
						discord.NewTextDisplay(successMsg),
					).WithAccessory(discord.NewThumbnail("https://raw.githubusercontent.com/googlefonts/noto-emoji/main/png/512/emoji_u1f3dd.png")),
				).WithAccentColor(0x2ECC71),
			)

			if err := event.CreateMessage(msg); err != nil {
				fmt.Println("Error sending daily success message:", err)
			}
		},
	})
}
