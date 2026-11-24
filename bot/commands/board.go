package commands

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"orion-event-bot/bot/board"
	"orion-event-bot/bot/core"
	"orion-event-bot/bot/handlers"
	"orion-event-bot/bot/vars"
	"orion-event-bot/db"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"gorm.io/gorm"
)

func init() {
	handlers.RegisterCommand(handlers.Command{
		Name:        "board",
		Description: "Displays the current Winter Rush board and game status",
		Create: discord.SlashCommandCreate{
			Name:        "board",
			Description: "Displays the current Winter Rush board and game status",
			DescriptionLocalizations: map[discord.Locale]string{
				discord.LocaleFrench: "Affiche le plateau actuel et les informations de la partie Winter Rush",
			},
		},
		Execute: executeBoard,
	})
}

func executeBoard(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
	if err := event.DeferCreateMessage(false); err != nil {
		fmt.Println("Error deferring /board interaction:", err)
		return
	}

	loc := core.GetLocalizer(event.Locale().Code())
	callerID := event.User().ID.String()

	var evt db.Event
	if err := b.DB.GormDB.First(&evt).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		fmt.Println("Error reading event state:", err)
	}

	var activeUsers []db.User
	if err := b.DB.GormDB.
		Where("position >= ?", 1).
		Order("position DESC, finished_at ASC, id ASC").
		Find(&activeUsers).Error; err != nil {
		fmt.Println("Error reading active users:", err)
	}

	userIDs := make([]string, len(activeUsers))
	callerPos := 0
	for i, u := range activeUsers {
		userIDs[i] = u.ID
		if u.ID == callerID {
			callerPos = u.Position
		}
	}

	var guildID snowflake.ID
	if gid := event.GuildID(); gid != nil {
		guildID = *gid
	}
	avatars := board.FetchAvatars(context.Background(), b.Client, guildID, userIDs)

	pawns := make([]board.PlayerPawn, len(activeUsers))
	for i, u := range activeUsers {
		pawns[i] = board.PlayerPawn{
			UserID:   u.ID,
			Position: u.Position,
			Avatar:   avatars[u.ID],
		}
	}

	labels := buildBoardLabels(loc, event.Locale().Code())
	pngBytes, err := board.RenderBoardPNG(board.RenderData{
		LodgeTotal: evt.LodgeTotal,
		LodgeGoal:  vars.LodgeGoal,
		Players:    pawns,
		CallerID:   callerID,
		Labels:     labels,
	})
	if err != nil {
		fmt.Println("Error rendering board image:", err)
		errMsg := localizeText(loc, "BoardRenderError", nil, "Failed to generate the board image.")
		_, _ = b.Client.Rest.UpdateInteractionResponse(
			event.ApplicationID(),
			event.Token(),
			discord.NewMessageUpdateV2(
				discord.NewContainer(
					discord.NewTextDisplay(errMsg),
				).WithAccentColor(0xEF4444),
			),
		)
		return
	}

	lodgePercent := 0
	if vars.LodgeGoal > 0 {
		lodgePercent = int(math.Round(float64(evt.LodgeTotal) * 100 / float64(vars.LodgeGoal)))
		if lodgePercent > 100 {
			lodgePercent = 100
		} else if lodgePercent < 0 {
			lodgePercent = 0
		}
	}

	title := localizeText(loc, "BoardTitle", nil, "# ⛷️ Winter Rush — Board")
	lodgeLine := localizeText(loc, "BoardLodge", map[string]interface{}{
		"LodgeTotal": evt.LodgeTotal,
		"LodgeGoal":  vars.LodgeGoal,
		"Percent":    lodgePercent,
	}, fmt.Sprintf("🏡 **Winter Lodge:** %d / %d (%d%%)", evt.LodgeTotal, vars.LodgeGoal, lodgePercent))

	playersLine := localizeText(loc, "BoardActivePlayers", map[string]interface{}{
		"Count": len(activeUsers),
	}, fmt.Sprintf("🎿 **Active Skiers:** %d", len(activeUsers)))

	var posLine string
	if callerPos >= 1 {
		posLine = localizeText(loc, "BoardYourPos", map[string]interface{}{
			"Position":   callerPos,
			"TotalCases": len(vars.Plate),
		}, fmt.Sprintf("📍 **Your Position:** Tile **%d / %d**", callerPos, len(vars.Plate)))
	} else {
		posLine = localizeText(loc, "BoardNotStarted", nil, "📍 **Your Position:** Not on the slopes yet.")
	}

	statsBlock := strings.Join([]string{lodgeLine, playersLine, posLine}, "\n")
	leaderboardBlock := buildLeaderboardText(loc, activeUsers)

	containerItems := []discord.ContainerSubComponent{
		discord.NewSection(
			discord.NewTextDisplay(title),
			discord.NewTextDisplay(statsBlock),
		).WithAccessory(discord.NewThumbnail(event.User().EffectiveAvatarURL())),
		discord.NewSeparator(discord.SeparatorSpacingSizeSmall),
		discord.NewTextDisplay(leaderboardBlock),
		discord.NewMediaGallery(
			discord.MediaGalleryItem{
				Media: discord.UnfurledMediaItem{
					URL: "attachment://board.png",
				},
			},
		),
	}

	msgUpdate := discord.NewMessageUpdateV2(
		discord.NewContainer(containerItems...).WithAccentColor(0x38BDF8),
	).WithFiles(
		discord.NewFile("board.png", "Winter Rush Board", bytes.NewReader(pngBytes)),
	)

	if _, err := b.Client.Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), msgUpdate); err != nil {
		fmt.Println("Error sending /board response:", err)
	}
}

func buildLeaderboardText(loc *i18n.Localizer, activeUsers []db.User) string {
	header := localizeText(loc, "BoardLeaderboardTitle", nil, "🏆 **Top 3 Skiers:**")
	if len(activeUsers) == 0 {
		empty := localizeText(loc, "BoardLeaderboardEmpty", nil, "_No players on the slopes yet._")
		return header + "\n" + empty
	}

	medals := []string{"🥇", "🥈", "🥉"}
	limit := len(activeUsers)
	if limit > 3 {
		limit = 3
	}

	lines := make([]string, 0, limit+1)
	lines = append(lines, header)
	for i := 0; i < limit; i++ {
		u := activeUsers[i]
		entry := localizeText(loc, "BoardLeaderboardEntry", map[string]interface{}{
			"Medal":    medals[i],
			"UserID":   u.ID,
			"Position": u.Position,
		}, fmt.Sprintf("%s <@%s> — Tile **%d**", medals[i], u.ID, u.Position))
		lines = append(lines, entry)
	}
	return strings.Join(lines, "\n")
}

func buildBoardLabels(loc *i18n.Localizer, localeKey string) board.BoardLabels {
	return board.BoardLabels{
		LocaleKey:  localeKey,
		LodgeLabel: localizeText(loc, "ImageLodgeLabel", nil, "Mountain Refuge Progress"),
		StartLabel: localizeText(loc, "ImageStartLabel", nil, "START"),
	}
}

func localizeText(loc *i18n.Localizer, id string, data map[string]interface{}, fallback string) string {
	msg, err := loc.Localize(&i18n.LocalizeConfig{
		MessageID:    id,
		TemplateData: data,
	})
	if err != nil || msg == "" {
		return fallback
	}
	return msg
}
