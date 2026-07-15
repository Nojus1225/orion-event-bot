package commands

import (
	"fmt"
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

type ShopItem struct {
	Name        string
	Description string
	Cost        int
	ID          db.Item
	Thumbnail   string
}

var Items = []ShopItem{
	{
		Name:        "ShopSunscreenName",
		Description: "ShopSunscreenDesc",
		Cost:        2,
		ID:          db.ShopCremeSolaire,
		Thumbnail:   "https://raw.githubusercontent.com/googlefonts/noto-emoji/main/png/512/emoji_u1f9f4.png",
	},
	{
		Name:        "ShopBandageName",
		Description: "ShopBandageDesc",
		Cost:        4,
		ID:          db.ShopPansement,
		Thumbnail:   "https://raw.githubusercontent.com/googlefonts/noto-emoji/main/png/512/emoji_u1fa79.png",
	},
	{
		Name:        "ShopWaterBottleName",
		Description: "ShopWaterBottleDesc",
		Cost:        10,
		ID:          db.ShopBouteilleEau,
		Thumbnail:   "https://raw.githubusercontent.com/googlefonts/noto-emoji/main/png/512/emoji_u1f4a7.png",
	},
}

func getShopData(page, totalPages int, user db.User, loc *i18n.Localizer) discord.ContainerComponent {
	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}

	firstBtn := discord.NewButton(discord.ButtonStylePrimary, "", "shop-first", "", snowflake.ID(0)).WithEmoji(discord.NewComponentEmoji("⏮️"))
	if page <= 1 {
		firstBtn = firstBtn.WithDisabled(true)
	}

	prevBtn := discord.NewButton(discord.ButtonStylePrimary, "", fmt.Sprintf("shop-%d", page-1), "", snowflake.ID(0)).WithEmoji(discord.NewComponentEmoji("◀️"))
	if page <= 1 {
		prevBtn = prevBtn.WithDisabled(true)
	}

	currBtn := discord.NewButton(discord.ButtonStyleSecondary, fmt.Sprintf("%d/%d", page, totalPages), "shop-current", "", snowflake.ID(0)).WithDisabled(true)

	nextBtn := discord.NewButton(discord.ButtonStylePrimary, "", fmt.Sprintf("shop-%d", page+1), "", snowflake.ID(0)).WithEmoji(discord.NewComponentEmoji("▶️"))
	if page >= totalPages {
		nextBtn = nextBtn.WithDisabled(true)
	}

	lastBtn := discord.NewButton(discord.ButtonStylePrimary, "", "shop-last", "", snowflake.ID(0)).WithEmoji(discord.NewComponentEmoji("⏭️"))
	if page >= totalPages {
		lastBtn = lastBtn.WithDisabled(true)
	}

	actionRow := discord.NewActionRow(firstBtn, prevBtn, currBtn, nextBtn, lastBtn)

	item := Items[page-1]
	
	locName, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: item.Name})
	locDesc, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: item.Description})
	
	coinsTxt, _ := loc.Localize(&i18n.LocalizeConfig{
		MessageID: "ShopCoins",
		TemplateData: map[string]interface{}{"Coins": item.Cost},
	})

	content := fmt.Sprintf("### %s\n-# %s\n%s", locName, locDesc, coinsTxt)

	canBuy := false
	if item.Cost <= user.Coins {
		canBuy = true
	}

	shopTitle, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "ShopTitle"})
	buyItemTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "ShopBuyItem"})

	return discord.NewContainer(
		discord.NewTextDisplay(shopTitle),
		discord.NewSection(
			discord.NewTextDisplay(content),
		).WithAccessory(
			discord.NewThumbnail(item.Thumbnail),
		),
		discord.NewActionRow(
			discord.NewButton(discord.ButtonStyleSuccess, buyItemTxt, fmt.Sprintf("shop-buy-%s", item.ID), "", snowflake.ID(0)).WithDisabled(!canBuy).WithEmoji(discord.NewComponentEmoji("🛒")),
		),
		actionRow,
	).WithAccentColor(0x5865F2)
}

var totalPages = len(Items)

func init() {
	handlers.RegisterCommand(handlers.Command{
		Name:        "shop",
		Description: "Open the shop",
		Create: discord.SlashCommandCreate{
			Name:        "shop",
			Description: "Open the shop",
			NameLocalizations: map[discord.Locale]string{
				discord.LocaleFrench: "boutique",
			},
			DescriptionLocalizations: map[discord.Locale]string{
				discord.LocaleFrench: "Ouvrir la boutique",
			},
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			loc := core.GetLocalizer(event.Locale().String())
			var data db.User
			if err := b.DB.GormDB.FirstOrCreate(&data, db.User{ID: strconv.FormatUint(uint64(event.User().ID), 10)}).Error; err != nil {
				fmt.Println("DB error:", err)
			}

			container := getShopData(1, totalPages, data, loc)

			msg := discord.NewMessageCreateV2(container)

			if err := event.CreateMessage(msg); err != nil {
				fmt.Println("Error sending shop message:", err)
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
			var data db.User
			if err := b.DB.GormDB.FirstOrCreate(&data, db.User{ID: strconv.FormatUint(uint64(event.User().ID), 10)}).Error; err != nil {
				fmt.Println("DB error:", err)
			}

			customID := event.Data.CustomID()
			parts := strings.SplitN(customID, "-", 2)

			if strings.HasPrefix(customID, "shop-buy") {
				parts := strings.SplitN(customID, "-", 3)
				if len(parts) != 3 {
					return
				}
				
				idStr := parts[2]
				var item *ShopItem
				for i := range Items {
					if string(Items[i].ID) == idStr {
						item = &Items[i]
						break
					}
				}
				if item == nil {
					return
				}

				if data.Coins < item.Cost {
					errMsg, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "ShopNotEnoughCoins"})
					msg := discord.NewMessageCreate().WithEphemeral(true).WithContent(errMsg)
					if err := event.CreateMessage(msg); err != nil {
						fmt.Println("Error sending shop error message:", err)
					}
					return
				}

				data.Coins -= item.Cost
				if err := b.DB.GormDB.Save(&data).Error; err != nil {
					fmt.Println("DB error:", err)
				}
				
				db.IncrementStat(b.DB.GormDB, data.ID, db.StatTotalCoinsDepense, item.Cost)
				db.IncrementItem(b.DB.GormDB, data.ID, item.ID, 1)

				switch item.ID {
				case db.ShopCremeSolaire:
					db.IncrementStat(b.DB.GormDB, data.ID, db.StatAchatCremeSolaire, 1)
				case db.ShopPansement:
					db.IncrementStat(b.DB.GormDB, data.ID, db.StatAchatPansement, 1)
				case db.ShopBouteilleEau:
					db.IncrementStat(b.DB.GormDB, data.ID, db.StatAchatEau, 1)
				}
				
				db.LogAction(b.DB.GormDB, data.ID, db.ActionShop, fmt.Sprintf(`{"item_id":"%s", "cost":%d}`, item.ID, item.Cost))

				var page int
				for i := range Items {
					if Items[i].ID == item.ID {
						page = i + 1
						break
					}
				}

				container := getShopData(page, totalPages, data, loc)
				msg := discord.NewMessageUpdateV2(container)

				if err := event.UpdateMessage(msg); err != nil {
					fmt.Println("Error updating shop message:", err)
				}

				successMsg, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "ShopBuySuccess"})
				followup := discord.NewMessageCreate().WithEphemeral(true).WithContent(successMsg)
				if _, err := event.Client().Rest.CreateFollowupMessage(event.ApplicationID(), event.Token(), followup); err != nil {
					fmt.Println("Error sending shop followup:", err)
				}

				return
			}

			if len(parts) != 2 {
				return
			}

			if parts[1] == "current" {
				return
			}

			var page int
			if parts[1] == "first" {
				page = 1
			} else if parts[1] == "last" {
				page = totalPages
			} else if strings.HasPrefix(parts[1], "buy") {
				return
			} else {
				p, err := strconv.Atoi(parts[1])
				if err != nil {
					return
				}
				page = p
			}

			container := getShopData(page, totalPages, data, loc)

			msg := discord.NewMessageUpdateV2(container)

			if err := event.UpdateMessage(msg); err != nil {
				fmt.Println("Error sending shop message:", err)
			}
		},
	})
}
