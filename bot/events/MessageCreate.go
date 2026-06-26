package events

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"orion-event-bot/bot/core"
	"orion-event-bot/bot/handlers"
	"orion-event-bot/db"

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

			case "give", "remove":
				if len(args) < 3 {
					_, _ = event.Client().Rest.CreateMessage(event.ChannelID, discord.NewMessageCreate().WithContent(fmt.Sprintf("Usage: ,%s <@user> <item> <amount>", command)))
					return
				}

				targetIDStr := args[0]
				match := mentionRegex.FindStringSubmatch(targetIDStr)
				if len(match) > 1 {
					targetIDStr = match[1]
				}

				_, err := strconv.ParseUint(targetIDStr, 10, 64)
				if err != nil {
					_, _ = event.Client().Rest.CreateMessage(event.ChannelID, discord.NewMessageCreate().WithContent("Invalid user ID."))
					return
				}

				item := strings.ToLower(args[1])
				amount, err := strconv.Atoi(args[2])
				if err != nil {
					_, _ = event.Client().Rest.CreateMessage(event.ChannelID, discord.NewMessageCreate().WithContent("Invalid amount."))
					return
				}

				if command == "remove" {
					amount = -amount
				}

				var user db.User
				if err := b.DB.GormDB.FirstOrCreate(&user, db.User{ID: targetIDStr}).Error; err != nil {
					_, _ = event.Client().Rest.CreateMessage(event.ChannelID, discord.NewMessageCreate().WithContent("DB error."))
					return
				}

				fieldToUpdate := ""
				switch item {
				case "coins":
					user.Coins += amount
					fieldToUpdate = "Coins"
				case "shell":
					user.ItemCoquillage += amount
					fieldToUpdate = "ItemCoquillage"
				case "bottle":
					user.ItemBouteille += amount
					fieldToUpdate = "ItemBouteille"
				case "can":
					user.ItemCannette += amount
					fieldToUpdate = "ItemCannette"
				case "glasses":
					user.ItemLunettes += amount
					fieldToUpdate = "ItemLunettes"
				case "crab_claw":
					user.ItemPince += amount
					fieldToUpdate = "ItemPince"
				case "msg_bottle":
					user.ItemMsgBouteille += amount
					fieldToUpdate = "ItemMsgBouteille"
				case "watch":
					user.ItemMontre += amount
					fieldToUpdate = "ItemMontre"
				case "shark_tooth":
					user.ItemDent += amount
					fieldToUpdate = "ItemDent"
				default:
					_, _ = event.Client().Rest.CreateMessage(event.ChannelID, discord.NewMessageCreate().WithContent("Unknown item: "+item))
					return
				}

				if err := b.DB.GormDB.Model(&user).Select(fieldToUpdate).Updates(user).Error; err != nil {
					_, _ = event.Client().Rest.CreateMessage(event.ChannelID, discord.NewMessageCreate().WithContent("Error updating DB."))
					return
				}

				_, _ = event.Client().Rest.CreateMessage(event.ChannelID, discord.NewMessageCreate().WithContent(fmt.Sprintf("Success! Updated %s for user %s.", item, targetIDStr)))

			case "resetcd":
				if len(args) < 1 {
					_, _ = event.Client().Rest.CreateMessage(event.ChannelID, discord.NewMessageCreate().WithContent("Usage: ,resetcd <@user>"))
					return
				}

				targetIDStr := args[0]
				match := mentionRegex.FindStringSubmatch(targetIDStr)
				if len(match) > 1 {
					targetIDStr = match[1]
				}

				var user db.User
				if err := b.DB.GormDB.FirstOrCreate(&user, db.User{ID: targetIDStr}).Error; err != nil {
					_, _ = event.Client().Rest.CreateMessage(event.ChannelID, discord.NewMessageCreate().WithContent("DB error."))
					return
				}

				user.CDFouille = 0
				user.CDMalus = 0
				user.StatusMalus = db.MalusNone

				if err := b.DB.GormDB.Model(&user).Select("CDFouille", "CDMalus", "StatusMalus").Updates(user).Error; err != nil {
					_, _ = event.Client().Rest.CreateMessage(event.ChannelID, discord.NewMessageCreate().WithContent("Error updating DB."))
					return
				}

				_, _ = event.Client().Rest.CreateMessage(event.ChannelID, discord.NewMessageCreate().WithContent(fmt.Sprintf("Reset cooldowns for user %s.", targetIDStr)))

			case "maintenance":
				if len(args) < 1 {
					status := "OFF"
					if b.MaintenanceMode {
						status = "ON"
					}
					_, _ = event.Client().Rest.CreateMessage(event.ChannelID, discord.NewMessageCreate().WithContent(fmt.Sprintf("Maintenance mode is currently %s.", status)))
					return
				}

				state := strings.ToLower(args[0])
				if state == "on" || state == "true" || state == "1" {
					b.MaintenanceMode = true
					_, _ = event.Client().Rest.CreateMessage(event.ChannelID, discord.NewMessageCreate().WithContent("Maintenance mode ENABLED"))
				} else if state == "off" || state == "false" || state == "0" {
					b.MaintenanceMode = false
					_, _ = event.Client().Rest.CreateMessage(event.ChannelID, discord.NewMessageCreate().WithContent("Maintenance mode DISABLED"))
				} else {
					_, _ = event.Client().Rest.CreateMessage(event.ChannelID, discord.NewMessageCreate().WithContent("Invalid state. Use 'on' or 'off'."))
				}

			case "globalstats":
				var totalUsers int64
				b.DB.GormDB.Model(&db.User{}).Count(&totalUsers)

				var totalCoins int64
				b.DB.GormDB.Model(&db.User{}).Select("COALESCE(SUM(coins), 0)").Scan(&totalCoins)

				var totalFouilles int64
				b.DB.GormDB.Model(&db.User{}).Select("COALESCE(SUM(stats_fouille), 0)").Scan(&totalFouilles)

				var totalObjets int64
				b.DB.GormDB.Model(&db.User{}).Select("COALESCE(SUM(stats_objets_trouve), 0)").Scan(&totalObjets)

				msg := fmt.Sprintf("📊 **Global Event Stats**\n- **Total Users**: %d\n- **Total Coins (Circulating)**: %d\n- **Total Searches (Fouilles)**: %d\n- **Total Items Found**: %d", totalUsers, totalCoins, totalFouilles, totalObjets)
				_, _ = event.Client().Rest.CreateMessage(event.ChannelID, discord.NewMessageCreate().WithContent(msg))
			}
		},
	})
}
