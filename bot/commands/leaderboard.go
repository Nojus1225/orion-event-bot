package commands

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"orion-event-bot/bot/core"
	"orion-event-bot/bot/handlers"
	"orion-event-bot/db"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

func getLeaderboardData(b *core.Bot, loc *i18n.Localizer, page int, totalPages int) (string, discord.ActionRowComponent) {
	limit := 10
	offset := (page - 1) * limit

	var users []db.User
	b.DB.GormDB.Preload("Items").Order("coins desc").Limit(limit).Offset(offset).Find(&users)

	titleTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "LeaderboardTitle"})
	content := titleTxt + "\n"
	for i, u := range users {
		rank := offset + i + 1

		counts := make(map[string]int)
		for _, it := range u.Items {
			counts[it.ItemID] = it.Quantity
		}

		common := counts[string(db.ItemCoquillage)] + counts[string(db.ItemBouteille)] + counts[string(db.ItemCannette)]
		rare := counts[string(db.ItemLunettes)] + counts[string(db.ItemPiece)] + counts[string(db.ItemPince)]
		epic := counts[string(db.ItemMsgBouteille)] + counts[string(db.ItemMontre)] + counts[string(db.ItemDent)]

		entry, _ := loc.Localize(&i18n.LocalizeConfig{
			MessageID: "LeaderboardEntry",
			TemplateData: map[string]interface{}{
				"Rank":   rank,
				"ID":     u.ID,
				"Coins":  u.Coins,
				"Common": common,
				"Rare":   rare,
				"Epic":   epic,
			},
		})
		content += entry + "\n"
	}

	if len(users) == 0 {
		noUsersTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "NoUsersFound"})
		content += noUsersTxt
	}

	firstBtn := discord.NewButton(discord.ButtonStylePrimary, "", "leaderboard-first", "", snowflake.ID(0)).WithEmoji(discord.NewComponentEmoji("⏮️"))
	if page <= 1 {
		firstBtn = firstBtn.WithDisabled(true)
	}

	prevBtn := discord.NewButton(discord.ButtonStylePrimary, "", fmt.Sprintf("leaderboard-%d", page-1), "", snowflake.ID(0)).WithEmoji(discord.NewComponentEmoji("◀️"))
	if page <= 1 {
		prevBtn = prevBtn.WithDisabled(true)
	}

	currBtn := discord.NewButton(discord.ButtonStyleSecondary, fmt.Sprintf("%d/%d", page, totalPages), "leaderboard-current", "", snowflake.ID(0)).WithDisabled(true)

	nextBtn := discord.NewButton(discord.ButtonStylePrimary, "", fmt.Sprintf("leaderboard-%d", page+1), "", snowflake.ID(0)).WithEmoji(discord.NewComponentEmoji("▶️"))
	if page >= totalPages {
		nextBtn = nextBtn.WithDisabled(true)
	}

	lastBtn := discord.NewButton(discord.ButtonStylePrimary, "", "leaderboard-last", "", snowflake.ID(0)).WithEmoji(discord.NewComponentEmoji("⏭️"))
	if page >= totalPages {
		lastBtn = lastBtn.WithDisabled(true)
	}

	actionRow := discord.NewActionRow(firstBtn, prevBtn, currBtn, nextBtn, lastBtn)
	return content, actionRow
}

func init() {
	handlers.RegisterCommand(handlers.Command{
		Name:        "leaderboard",
		Description: "View the leaderboard",
		Create: discord.SlashCommandCreate{
			Name:        "leaderboard",
			Description: "View the leaderboard",
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			loc := core.GetLocalizer(event.Locale().String())
			var totalCount int64
			b.DB.GormDB.Model(&db.User{}).Count(&totalCount)
			totalPages := int(math.Ceil(float64(totalCount) / 5.0))
			if totalPages == 0 {
				totalPages = 1
			}

			content, actionRow := getLeaderboardData(b, loc, 1, totalPages)
			msg := discord.NewMessageCreateV2(
				discord.NewContainer(
					discord.NewTextDisplay(content),
					actionRow,
				).WithAccentColor(0x5865F2),
			)

			if err := event.CreateMessage(msg); err != nil {
				fmt.Println("Error sending leaderboard message:", err)
			}
		},
		ExecuteButton: func(b *core.Bot, event *events.ComponentInteractionCreate) {
			loc := core.GetLocalizer(event.Locale().String())
			notYourComponentTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "NotYourComponent"})
			if (event.Message.InteractionMetadata != nil && event.Message.InteractionMetadata.User.ID != event.User().ID) ||
				(event.Message.Interaction != nil && event.Message.Interaction.User.ID != event.User().ID) {
				_ = event.CreateMessage(discord.NewMessageCreate().WithContent(notYourComponentTxt).WithEphemeral(true))
				return
			}
			customID := event.Data.CustomID()
			parts := strings.SplitN(customID, "-", 2)
			if len(parts) != 2 {
				return
			}

			if parts[1] == "current" {
				return
			}

			var totalCount int64
			b.DB.GormDB.Model(&db.User{}).Count(&totalCount)
			totalPages := int(math.Ceil(float64(totalCount) / 5.0))
			if totalPages == 0 {
				totalPages = 1
			}

			var page int
			if parts[1] == "first" {
				page = 1
			} else if parts[1] == "last" {
				page = totalPages
			} else {
				p, err := strconv.Atoi(parts[1])
				if err != nil {
					return
				}
				page = p
			}

			content, actionRow := getLeaderboardData(b, loc, page, totalPages)

			msg := discord.NewMessageUpdateV2(
				discord.NewContainer(
					discord.NewSection(
						discord.NewTextDisplay(content),
					),
					actionRow,
				).WithAccentColor(0x5865F2),
			)

			if err := event.UpdateMessage(msg); err != nil {
				fmt.Println("Error sending leaderboard message:", err)
			}
		},
	})
}
