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
)

func getLeaderboardData(b *core.Bot, page int, totalPages int) (string, discord.ActionRowComponent) {
	limit := 10
	offset := (page - 1) * limit

	var users []db.User
	b.DB.GormDB.Order("coins desc").Limit(limit).Offset(offset).Find(&users)

	content := "## Leaderboard\n"
	for i, u := range users {
		rank := offset + i + 1
		content += fmt.Sprintf("**#%d** <@%s> - **%d** Coins\n> Common: %d | Rare: %d | Epic: %d\n", rank, u.ID, u.Coins, u.ItemCoquillage+u.ItemBouteille+u.ItemCannette, u.ItemLunettes+u.ItemPiece+u.ItemPince, u.ItemMsgBouteille+u.ItemMontre+u.ItemDent)
	}

	if len(users) == 0 {
		content += "No users found."
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
			var totalCount int64
			b.DB.GormDB.Model(&db.User{}).Count(&totalCount)
			totalPages := int(math.Ceil(float64(totalCount) / 10.0))
			if totalPages == 0 {
				totalPages = 1
			}

			content, actionRow := getLeaderboardData(b, 1, totalPages)
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
			if (event.Message.InteractionMetadata != nil && event.Message.InteractionMetadata.User.ID != event.User().ID) ||
				(event.Message.Interaction != nil && event.Message.Interaction.User.ID != event.User().ID) {
				_ = event.CreateMessage(discord.NewMessageCreate().WithContent("You cannot interact with components on someone else's command.").WithEphemeral(true))
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
			totalPages := int(math.Ceil(float64(totalCount) / 10.0))
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

			content, actionRow := getLeaderboardData(b, page, totalPages)

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
