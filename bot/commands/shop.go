package commands

import (
	"fmt"
	"sort"
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

func getShopItems() []vars.ItemDef {
	items := make([]vars.ItemDef, 0, len(vars.Items))
	for _, item := range vars.Items {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Price == items[j].Price {
			return items[i].ID < items[j].ID
		}
		return items[i].Price < items[j].Price
	})
	return items
}

func getShopData(page, totalPages int, user db.User, loc *i18n.Localizer) discord.ContainerComponent {
	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}

	firstBtn := discord.NewButton(discord.ButtonStylePrimary, "", "shop-nav-1-first", "", snowflake.ID(0)).WithEmoji(discord.NewComponentEmoji("⏮️"))
	if page <= 1 {
		firstBtn = firstBtn.WithDisabled(true)
	}

	prevBtn := discord.NewButton(discord.ButtonStylePrimary, "", fmt.Sprintf("shop-nav-%d-prev", page-1), "", snowflake.ID(0)).WithEmoji(discord.NewComponentEmoji("◀️"))
	if page <= 1 {
		prevBtn = prevBtn.WithDisabled(true)
	}

	currBtn := discord.NewButton(discord.ButtonStyleSecondary, fmt.Sprintf("%d/%d", page, totalPages), fmt.Sprintf("shop-nav-%d-curr", page), "", snowflake.ID(0)).WithDisabled(true)

	nextBtn := discord.NewButton(discord.ButtonStylePrimary, "", fmt.Sprintf("shop-nav-%d-next", page+1), "", snowflake.ID(0)).WithEmoji(discord.NewComponentEmoji("▶️"))
	if page >= totalPages {
		nextBtn = nextBtn.WithDisabled(true)
	}

	lastBtn := discord.NewButton(discord.ButtonStylePrimary, "", fmt.Sprintf("shop-nav-%d-last", totalPages), "", snowflake.ID(0)).WithEmoji(discord.NewComponentEmoji("⏭️"))
	if page >= totalPages {
		lastBtn = lastBtn.WithDisabled(true)
	}

	actionRow := discord.NewActionRow(firstBtn, prevBtn, currBtn, nextBtn, lastBtn)

	items := getShopItems()
	item := items[page-1]

	priceTxt, _ := loc.Localize(&i18n.LocalizeConfig{
		MessageID:    "ShopPrice",
		TemplateData: map[string]interface{}{"Price": item.Price},
	})

	content := fmt.Sprintf("### %s %s\n-# %s\n%s", item.Emoji, item.Name, item.Description, priceTxt)

	canBuy := item.Price <= user.Snowflakes && len(user.Inventory) < 2

	shopTitle, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "ShopTitle"})
	buyItemTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "ShopBuyItem"})

	return discord.NewContainer(
		discord.NewTextDisplay(shopTitle),
		discord.NewTextDisplay(content),
		discord.NewActionRow(
			discord.NewButton(discord.ButtonStyleSuccess, buyItemTxt, fmt.Sprintf("shop-buy-%s", item.ID), "", snowflake.ID(0)).WithDisabled(!canBuy).WithEmoji(discord.NewComponentEmoji("🛒")),
		),
		actionRow,
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
			var user db.User
			if err := b.DB.GormDB.Where("id = ?", event.User().ID.String()).First(&user).Error; err != nil {
				fmt.Println("DB error:", err)
				return
			}

			items := getShopItems()
			container := getShopData(1, len(items), user, loc)

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

			var user db.User
			if err := b.DB.GormDB.Where("id = ?", event.User().ID.String()).First(&user).Error; err != nil {
				fmt.Println("DB error:", err)
				return
			}

			customID := event.Data.CustomID()
			parts := strings.Split(customID, "-")
			if len(parts) < 3 {
				return
			}

			items := getShopItems()
			totalPages := len(items)

			if parts[1] == "buy" {
				idStr := strings.Join(parts[2:], "-")
				var item *vars.ItemDef
				var page int
				for i := range items {
					if string(items[i].ID) == idStr {
						item = &items[i]
						page = i + 1
						break
					}
				}
				if item == nil {
					return
				}

				if user.Snowflakes < item.Price {
					errMsg, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "ShopNotEnoughSnowflakes"})
					_ = event.CreateMessage(discord.NewMessageCreate().WithEphemeral(true).WithContent(errMsg))
					return
				}

				if len(user.Inventory) >= 2 {
					errMsg, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "ShopInventoryFull"})
					_ = event.CreateMessage(discord.NewMessageCreate().WithEphemeral(true).WithContent(errMsg))
					return
				}

				user.Snowflakes -= item.Price
				user.Inventory = append(user.Inventory, string(item.ID))
				if err := b.DB.GormDB.Save(&user).Error; err != nil {
					fmt.Println("DB error:", err)
					return
				}

				container := getShopData(page, totalPages, user, loc)
				if err := event.UpdateMessage(discord.NewMessageUpdateV2(container)); err != nil {
					fmt.Println("Error updating shop message:", err)
				}

				successMsg, _ := loc.Localize(&i18n.LocalizeConfig{
					MessageID:    "ShopBuySuccess",
					TemplateData: map[string]interface{}{"Item": fmt.Sprintf("%s %s", item.Emoji, item.Name)},
				})
				followup := discord.NewMessageCreate().WithEphemeral(true).WithContent(successMsg)
				if _, err := event.Client().Rest.CreateFollowupMessage(event.ApplicationID(), event.Token(), followup); err != nil {
					fmt.Println("Error sending shop followup:", err)
				}
				return
			}

			if parts[1] == "nav" {
				page, err := strconv.Atoi(parts[2])
				if err != nil {
					return
				}

				container := getShopData(page, totalPages, user, loc)
				if err := event.UpdateMessage(discord.NewMessageUpdateV2(container)); err != nil {
					fmt.Println("Error updating shop message:", err)
				}
			}
		},
	})
}
