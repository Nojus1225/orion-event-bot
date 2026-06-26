package commands

import (
	"fmt"
	"strconv"

	"orion-event-bot/bot/core"
	"orion-event-bot/bot/handlers"
	"orion-event-bot/db"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/nicksnyder/go-i18n/v2/i18n"
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
			loc := core.GetLocalizer(event.Locale().String())
			command := event.SlashCommandInteractionData()

			userOpt, exist := command.OptUser("user")

			var targetUser discord.User
			if exist {
				targetUser = userOpt
			} else {
				targetUser = event.User()
			}

			avatarURL := targetUser.EffectiveAvatarURL()

			var data db.User
			if err := b.DB.GormDB.FirstOrCreate(&data, db.User{ID: strconv.FormatUint(uint64(targetUser.ID), 10)}).Error; err != nil {
				fmt.Println("DB error:", err)
			}

			tierStr, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "None"})
			switch data.Tier {
			case db.TierSac:
				tierStr, _ = loc.Localize(&i18n.LocalizeConfig{MessageID: "Backpack"})
			case db.TierTamis:
				tierStr, _ = loc.Localize(&i18n.LocalizeConfig{MessageID: "Sieve"})
			case db.TierDetecteur:
				tierStr, _ = loc.Localize(&i18n.LocalizeConfig{MessageID: "Detector"})
			}

			malusStr, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "None"})
			switch data.StatusMalus {
			case db.MalusSunBurn:
				malusStr, _ = loc.Localize(&i18n.LocalizeConfig{MessageID: "Sunburn"})
			case db.MalusCut:
				malusStr, _ = loc.Localize(&i18n.LocalizeConfig{MessageID: "Cut"})
			}

			commonLines := []string{}
			if data.ItemCoquillage > 0 {
				itemTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "Shell"})
				commonLines = append(commonLines, fmt.Sprintf("> %s : %d", itemTxt, data.ItemCoquillage))
			}
			if data.ItemBouteille > 0 {
				itemTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "Bottle"})
				commonLines = append(commonLines, fmt.Sprintf("> %s : %d", itemTxt, data.ItemBouteille))
			}
			if data.ItemCannette > 0 {
				itemTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "Can"})
				commonLines = append(commonLines, fmt.Sprintf("> %s : %d", itemTxt, data.ItemCannette))
			}

			rareLines := []string{}
			if data.ItemLunettes > 0 {
				itemTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "Glasses"})
				rareLines = append(rareLines, fmt.Sprintf("> %s : %d", itemTxt, data.ItemLunettes))
			}
			if data.ItemPiece > 0 {
				itemTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "Coins"})
				rareLines = append(rareLines, fmt.Sprintf("> %s : %d", itemTxt, data.ItemPiece))
			}
			if data.ItemPince > 0 {
				itemTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "CrabClaw"})
				rareLines = append(rareLines, fmt.Sprintf("> %s : %d", itemTxt, data.ItemPince))
			}

			legendaryLines := []string{}
			if data.ItemMsgBouteille > 0 {
				itemTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "MessageInABottle"})
				legendaryLines = append(legendaryLines, fmt.Sprintf("> %s : %d", itemTxt, data.ItemMsgBouteille))
			}
			if data.ItemMontre > 0 {
				itemTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "Watch"})
				legendaryLines = append(legendaryLines, fmt.Sprintf("> %s : %d", itemTxt, data.ItemMontre))
			}
			if data.ItemDent > 0 {
				itemTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "SharkTooth"})
				legendaryLines = append(legendaryLines, fmt.Sprintf("> %s : %d", itemTxt, data.ItemDent))
			}

			noItems, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "NoItems"})
			commonItemsTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "CommonItems"})
			rareItemsTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "RareItems"})
			legItemsTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "LegendaryItems"})
			shopItemsTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "ShopItems"})

			objectStr := ""
			objectStr += "## " + commonItemsTxt + "\n"
			if len(commonLines) == 0 {
				objectStr += noItems + "\n"
			} else {
				for _, l := range commonLines {
					objectStr += l + "\n"
				}
			}

			objectStr += "## " + rareItemsTxt + "\n"
			if len(rareLines) == 0 {
				objectStr += noItems + "\n"
			} else {
				for _, l := range rareLines {
					objectStr += l + "\n"
				}
			}

			objectStr += "## " + legItemsTxt + "\n"
			if len(legendaryLines) == 0 {
				objectStr += noItems + "\n"
			} else {
				for _, l := range legendaryLines {
					objectStr += l + "\n"
				}
			}

			objectStr += "## " + shopItemsTxt + "\n"
			if len(data.Inventory) == 0 {
				objectStr += noItems + "\n"
			} else {
				counts := make(map[int]int)
				for _, id := range data.Inventory {
					counts[id]++
				}
				for _, j := range Items {
					if cnt, ok := counts[int(j.ID)]; ok {
						itemLoc, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: j.Name})
						if itemLoc == "" {
							itemLoc = j.Name
						}
						objectStr += fmt.Sprintf("(x%d) %s\n", cnt, itemLoc)
					}
				}
			}

			invTitle, _ := loc.Localize(&i18n.LocalizeConfig{
				MessageID:    "InventoryTitle",
				TemplateData: map[string]interface{}{"User": targetUser.Username},
			})

			invStats, _ := loc.Localize(&i18n.LocalizeConfig{
				MessageID:    "InventoryStats",
				TemplateData: map[string]interface{}{"Coins": data.Coins, "Penalty": malusStr, "Tier": tierStr},
			})

			msg := discord.NewMessageCreateV2(
				discord.NewContainer(
					discord.NewSection(
						discord.NewTextDisplay(invTitle),
						discord.NewTextDisplay(invStats),
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
