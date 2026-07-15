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
			DescriptionLocalizations: map[discord.Locale]string{
				discord.LocaleFrench: "Voir l'inventaire d'un utilisateur",
			},
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionUser{
					Name:        "user",
					Description: "Get this user inventory",
					DescriptionLocalizations: map[discord.Locale]string{
						discord.LocaleFrench: "Voir l'inventaire de cet utilisateur",
					},
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

			getQ := func(i db.Item) int {
				return db.GetItemQuantity(b.DB.GormDB, data.ID, i)
			}

			commonLines := []string{}
			if c := getQ(db.ItemCoquillage); c > 0 {
				itemTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "Shell"})
				commonLines = append(commonLines, fmt.Sprintf("> %s : %d", itemTxt, c))
			}
			if c := getQ(db.ItemBouteille); c > 0 {
				itemTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "Bottle"})
				commonLines = append(commonLines, fmt.Sprintf("> %s : %d", itemTxt, c))
			}
			if c := getQ(db.ItemCannette); c > 0 {
				itemTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "Can"})
				commonLines = append(commonLines, fmt.Sprintf("> %s : %d", itemTxt, c))
			}

			rareLines := []string{}
			if c := getQ(db.ItemLunettes); c > 0 {
				itemTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "Glasses"})
				rareLines = append(rareLines, fmt.Sprintf("> %s : %d", itemTxt, c))
			}
			if c := getQ(db.ItemPiece); c > 0 {
				itemTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "Coins"})
				rareLines = append(rareLines, fmt.Sprintf("> %s : %d", itemTxt, c))
			}
			if c := getQ(db.ItemPince); c > 0 {
				itemTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "CrabClaw"})
				rareLines = append(rareLines, fmt.Sprintf("> %s : %d", itemTxt, c))
			}

			legendaryLines := []string{}
			if c := getQ(db.ItemMsgBouteille); c > 0 {
				itemTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "MessageInABottle"})
				legendaryLines = append(legendaryLines, fmt.Sprintf("> %s : %d", itemTxt, c))
			}
			if c := getQ(db.ItemMontre); c > 0 {
				itemTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "Watch"})
				legendaryLines = append(legendaryLines, fmt.Sprintf("> %s : %d", itemTxt, c))
			}
			if c := getQ(db.ItemDent); c > 0 {
				itemTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "SharkTooth"})
				legendaryLines = append(legendaryLines, fmt.Sprintf("> %s : %d", itemTxt, c))
			}
			
			shopLines := []string{}
			for _, j := range Items {
				if c := getQ(j.ID); c > 0 {
					itemLoc, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: j.Name})
					if itemLoc == "" {
						itemLoc = j.Name
					}
					shopLines = append(shopLines, fmt.Sprintf("(x%d) %s", c, itemLoc))
				}
			}

			noItems, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "NoItems"})
			commonItemsTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "CommonItems"})
			rareItemsTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "RareItems"})
			legItemsTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "LegendaryItems"})
			shopItemsTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "ShopItems"})

			objectStr := ""
			objectStr += "**" + commonItemsTxt + "**\n"
			if len(commonLines) == 0 {
				objectStr += noItems + "\n"
			} else {
				for _, l := range commonLines {
					objectStr += l + "\n"
				}
			}
			objectStr += "\n"

			objectStr += "**" + rareItemsTxt + "**\n"
			if len(rareLines) == 0 {
				objectStr += noItems + "\n"
			} else {
				for _, l := range rareLines {
					objectStr += l + "\n"
				}
			}
			objectStr += "\n"

			objectStr += "**" + legItemsTxt + "**\n"
			if len(legendaryLines) == 0 {
				objectStr += noItems + "\n"
			} else {
				for _, l := range legendaryLines {
					objectStr += l + "\n"
				}
			}
			objectStr += "\n"

			objectStr += "**" + shopItemsTxt + "**\n"
			if len(shopLines) == 0 {
				objectStr += noItems + "\n"
			} else {
				for _, l := range shopLines {
					objectStr += l + "\n"
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
