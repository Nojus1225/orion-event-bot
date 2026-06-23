package commands

import (
	"fmt"
	"strconv"

	"orion-event-bot/bot/core"
	"orion-event-bot/bot/handlers"
	"orion-event-bot/db"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	json "github.com/disgoorg/json/v2"
)

func init() {
	handlers.RegisterCommand(handlers.Command{
		Name:        "inventory",
		Description: "Get user inventory",
		Create: discord.SlashCommandCreate{
			Name:        "inventory",
			Description: "Get user inventory",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionUser{
					Name:        "user",
					Description: "Get this user inventory",
					Required:    false,
				},
			},
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			avatarURL := event.User().EffectiveAvatarURL()
			command := event.SlashCommandInteractionData()

			userOpt, exist := command.Option("user")

			var targetUser discord.User
			if exist {
				if err := json.Unmarshal(userOpt.Value, &targetUser); err != nil {
					targetUser = event.User()
				} else {
				}
			} else {
				targetUser = event.User()
			}

			var data db.User
			if err := b.DB.GormDB.FirstOrCreate(&data, db.User{ID: strconv.FormatUint(uint64(targetUser.ID), 10)}).Error; err != nil {
				fmt.Println("DB error:", err)
			}

			tierStr := "None"
			switch data.Tier {
			case db.TierSac:
				tierStr = "Backpack"
			case db.TierTamis:
				tierStr = "Sieve"
			case db.TierDetecteur:
				tierStr = "Detector"
			}

			malusStr := "none"
			switch data.StatusMalus {
			case db.MalusSunBurn:
				malusStr = "Sunburn"
			case db.MalusCut:
				malusStr = "Cut"
			}

			commonLines := []string{}
			if data.ItemCoquillage > 0 {
				commonLines = append(commonLines, fmt.Sprintf("> Shell : %d", data.ItemCoquillage))
			}
			if data.ItemBouteille > 0 {
				commonLines = append(commonLines, fmt.Sprintf("> Bottle : %d", data.ItemBouteille))
			}
			if data.ItemCannette > 0 {
				commonLines = append(commonLines, fmt.Sprintf("> Can : %d", data.ItemCannette))
			}

			rareLines := []string{}
			if data.ItemLunettes > 0 {
				rareLines = append(rareLines, fmt.Sprintf("> Glasses : %d", data.ItemLunettes))
			}
			if data.ItemPiece > 0 {
				rareLines = append(rareLines, fmt.Sprintf("> Coins : %d", data.ItemPiece))
			}
			if data.ItemPince > 0 {
				rareLines = append(rareLines, fmt.Sprintf("> Crab Claw : %d", data.ItemPince))
			}

			legendaryLines := []string{}
			if data.ItemMsgBouteille > 0 {
				legendaryLines = append(legendaryLines, fmt.Sprintf("> Message in a Bottle : %d", data.ItemMsgBouteille))
			}
			if data.ItemMontre > 0 {
				legendaryLines = append(legendaryLines, fmt.Sprintf("> Watch : %d", data.ItemMontre))
			}
			if data.ItemDent > 0 {
				legendaryLines = append(legendaryLines, fmt.Sprintf("> Tooth : %d", data.ItemDent))
			}

			objectStr := ""
			objectStr += "## Common Items\n"
			if len(commonLines) == 0 {
				objectStr += "> There are no items\n"
			} else {
				for _, l := range commonLines {
					objectStr += l + "\n"
				}
			}

			objectStr += "## Rare Items\n"
			if len(rareLines) == 0 {
				objectStr += "> There are no items\n"
			} else {
				for _, l := range rareLines {
					objectStr += l + "\n"
				}
			}

			objectStr += "## Legendary Items\n"
			if len(legendaryLines) == 0 {
				objectStr += "> There are no items\n"
			} else {
				for _, l := range legendaryLines {
					objectStr += l + "\n"
				}
			}

			msg := discord.NewMessageCreateV2(
				discord.NewContainer(
					discord.NewSection(
						discord.NewTextDisplay(fmt.Sprintf("# %s Inventory", targetUser.Username)),
						discord.NewTextDisplay(fmt.Sprintf("Coins: %d\nPenalty: %s\nTier: %s", data.Coins, malusStr, tierStr)),
						discord.NewTextDisplay(objectStr),
					).WithAccessory(discord.NewThumbnail(avatarURL)),
				).WithAccentColor(0x5865F2),
			)

			if err := event.CreateMessage(msg); err != nil {
				fmt.Println("Error sending inventory message:", err)
			}
		},
	})
}
