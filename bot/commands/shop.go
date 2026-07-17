package commands

import (
	"fmt"
	"strconv"
	"strings"

	"orion-event-bot/bot/core"
	"orion-event-bot/bot/handlers"
	"orion-event-bot/bot/vars"
	"orion-event-bot/db"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

func getShopData(categoryIndex, page, totalPages int, user db.User, loc *i18n.Localizer) discord.ContainerComponent {
	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}

	firstBtn := discord.NewButton(discord.ButtonStylePrimary, "", fmt.Sprintf("shop-nav-%d-1-first", categoryIndex), "", snowflake.ID(0)).WithEmoji(discord.NewComponentEmoji("⏮️"))
	if page <= 1 {
		firstBtn = firstBtn.WithDisabled(true)
	}

	prevBtn := discord.NewButton(discord.ButtonStylePrimary, "", fmt.Sprintf("shop-nav-%d-%d-prev", categoryIndex, page-1), "", snowflake.ID(0)).WithEmoji(discord.NewComponentEmoji("◀️"))
	if page <= 1 {
		prevBtn = prevBtn.WithDisabled(true)
	}

	currBtn := discord.NewButton(discord.ButtonStyleSecondary, fmt.Sprintf("%d/%d", page, totalPages), fmt.Sprintf("shop-nav-%d-%d-curr", categoryIndex, page), "", snowflake.ID(0)).WithDisabled(true)

	nextBtn := discord.NewButton(discord.ButtonStylePrimary, "", fmt.Sprintf("shop-nav-%d-%d-next", categoryIndex, page+1), "", snowflake.ID(0)).WithEmoji(discord.NewComponentEmoji("▶️"))
	if page >= totalPages {
		nextBtn = nextBtn.WithDisabled(true)
	}

	lastBtn := discord.NewButton(discord.ButtonStylePrimary, "", fmt.Sprintf("shop-nav-%d-%d-last", categoryIndex, totalPages), "", snowflake.ID(0)).WithEmoji(discord.NewComponentEmoji("⏭️"))
	if page >= totalPages {
		lastBtn = lastBtn.WithDisabled(true)
	}

	actionRow := discord.NewActionRow(firstBtn, prevBtn, currBtn, nextBtn, lastBtn)

	item := vars.Shop[categoryIndex].Items[page-1]

	locName, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: item.Name})
	locDesc, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: item.Description})

	coinsTxt, _ := loc.Localize(&i18n.LocalizeConfig{
		MessageID:    "ShopCoins",
		TemplateData: map[string]interface{}{"Coins": item.Cost},
	})

	content := fmt.Sprintf("### %s\n-# %s\n%s", locName, locDesc, coinsTxt)

	canBuy := false
	if item.Cost <= user.Coins {
		canBuy = true
	}

	shopTitle, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "ShopTitle"})
	buyItemTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "ShopBuyItem"})

	var options []discord.StringSelectMenuOption
	for i, cat := range vars.Shop {
		catName, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: cat.Name})
		catDesc, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: cat.Description})
		options = append(options, discord.StringSelectMenuOption{
			Label:       catName,
			Value:       strconv.Itoa(i),
			Description: catDesc,
			Default:     i == categoryIndex,
		})
	}

	shopCategoryPrompt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "ShopCategoryPrompt"})
	if shopCategoryPrompt == "" {
		shopCategoryPrompt = "Quelle catégorie souhaitez-vous parcourir ?"
	}
	selectMenu := discord.NewStringSelectMenu("shop-category", shopCategoryPrompt, options...)

	return discord.NewContainer(
		discord.NewTextDisplay(shopTitle),
		discord.NewSection(
			discord.NewTextDisplay(content),
		).WithAccessory(
			discord.NewThumbnail(item.Thumbnail),
		),
		discord.NewActionRow(
			discord.NewButton(discord.ButtonStyleSuccess, buyItemTxt, fmt.Sprintf("shop-buy-%d-%s", categoryIndex, item.ID), "", snowflake.ID(0)).WithDisabled(!canBuy).WithEmoji(discord.NewComponentEmoji("🛒")),
		),
		actionRow,
		discord.NewActionRow(selectMenu),
	).WithAccentColor(0x5865F2)
}

func init() {
	handlers.RegisterCommand(handlers.Command{
		Name:        "shop",
		Description: "Open the shop",
		Create: discord.SlashCommandCreate{
			Name:        "shop",
			Description: "Open the shop",

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

			totalPages := len(vars.Shop[0].Items)
			container := getShopData(0, 1, totalPages, data, loc)

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

			parts := strings.Split(customID, "-")
			if len(parts) < 3 {
				return
			}

			if parts[1] == "buy" {
				if len(parts) < 4 {
					return
				}
				catIndex, err := strconv.Atoi(parts[2])
				if err != nil {
					return
				}
				idStr := strings.Join(parts[3:], "-")

				var item *vars.ShopItem
				for i := range vars.Shop[catIndex].Items {
					if string(vars.Shop[catIndex].Items[i].ID) == idStr {
						item = &vars.Shop[catIndex].Items[i]
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
				for i := range vars.Shop[catIndex].Items {
					if vars.Shop[catIndex].Items[i].ID == item.ID {
						page = i + 1
						break
					}
				}
				totalPages := len(vars.Shop[catIndex].Items)
				container := getShopData(catIndex, page, totalPages, data, loc)
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

			if parts[1] == "nav" {
				if len(parts) < 4 {
					return
				}
				catIndex, err := strconv.Atoi(parts[2])
				if err != nil {
					return
				}
				page, err := strconv.Atoi(parts[3])
				if err != nil {
					return
				}

				totalPages := len(vars.Shop[catIndex].Items)
				container := getShopData(catIndex, page, totalPages, data, loc)

				msg := discord.NewMessageUpdateV2(container)

				if err := event.UpdateMessage(msg); err != nil {
					fmt.Println("Error sending shop message:", err)
				}
			}
		},
		ExecuteMenu: func(b *core.Bot, event *events.ComponentInteractionCreate) {
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

			if customID == "shop-category" {
				catIndexStr := event.StringSelectMenuInteractionData().Values[0]
				catIndex, err := strconv.Atoi(catIndexStr)
				if err != nil {
					return
				}
				totalPages := len(vars.Shop[catIndex].Items)
				container := getShopData(catIndex, 1, totalPages, data, loc)
				msg := discord.NewMessageUpdateV2(container)
				if err := event.UpdateMessage(msg); err != nil {
					fmt.Println("Error updating shop message:", err)
				}
				return
			}
		},
	})
}
