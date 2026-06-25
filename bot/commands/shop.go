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
)

type shopItem struct {
	Name        string
	Description string
	Cost        int
	ID          db.Item
	Thumbnail   string
}

var items = []shopItem{
	{
		Name:        "Sunscreen",
		Description: "Cures the 'Sunburn' debuff",
		Cost:        2,
		ID:          db.ShopCremeSolaire,
		Thumbnail:   "https://raw.githubusercontent.com/googlefonts/noto-emoji/main/png/512/emoji_u1f9f4.png",
	},
	{
		Name:        "Bandage",
		Description: "Cures the 'Cut' debuff",
		Cost:        4,
		ID:          db.ShopPansement,
		Thumbnail:   "https://raw.githubusercontent.com/googlefonts/noto-emoji/main/png/512/emoji_u1fa79.png",
	},
	{
		Name:        "Water Bottle",
		Description: "Allows scavenging without cooldown (per unit)",
		Cost:        10,
		ID:          db.ShopBouteilleEau,
		Thumbnail:   "https://raw.githubusercontent.com/googlefonts/noto-emoji/main/png/512/emoji_u1f4a7.png",
	},
}

func getShopData(page, totalPages int, user db.User) discord.ContainerComponent {
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

	item := items[page-1]
	content := fmt.Sprintf("### %s\n-# %s\n> Coins : %d", item.Name, item.Description, item.Cost)

	canBuy := false
	if item.Cost <= user.Coins {
		canBuy = true
	}

	return discord.NewContainer(
		discord.NewTextDisplay("# Shop"),
		discord.NewSection(
			discord.NewTextDisplay(content),
		).WithAccessory(
			discord.NewThumbnail(item.Thumbnail),
		),
		discord.NewActionRow(
			discord.NewButton(discord.ButtonStyleSuccess, "Buy item", fmt.Sprintf("shop-buy-%d", item.ID), "", snowflake.ID(0)).WithDisabled(!canBuy).WithEmoji(discord.NewComponentEmoji("🛒")),
		),
		actionRow,
	).WithAccentColor(0x5865F2)
}

var totalPages = len(items)

func init() {
	handlers.RegisterCommand(handlers.Command{
		Name:        "shop",
		Description: "Open the shop",
		Create: discord.SlashCommandCreate{
			Name:        "shop",
			Description: "Open the shop",
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			var data db.User
			if err := b.DB.GormDB.FirstOrCreate(&data, db.User{ID: strconv.FormatUint(uint64(event.User().ID), 10)}).Error; err != nil {
				fmt.Println("DB error:", err)
			}

			container := getShopData(1, totalPages, data)

			msg := discord.NewMessageCreateV2(container)

			if err := event.CreateMessage(msg); err != nil {
				fmt.Println("Error sending shop message:", err)
			}
		},
		ExecuteButton: func(b *core.Bot, event *events.ComponentInteractionCreate) {
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
				id, err := strconv.Atoi(parts[2])
				if err != nil {
					return
				}
				var item *shopItem
				for i := range items {
					if int(items[i].ID) == id {
						item = &items[i]
						break
					}
				}
				if item == nil {
					return
				}

				if data.Coins < item.Cost {
					msg := discord.NewMessageCreate().WithEphemeral(true).WithContent(":x: You don't have enough coins to buy this item!")
					if err := event.CreateMessage(msg); err != nil {
						fmt.Println("Error sending shop error message:", err)
					}
					return
				}

				data.Coins -= item.Cost
				data.Inventory = append(data.Inventory, int(item.ID))
				if err := b.DB.GormDB.Model(&data).Select("Coins", "Inventory").Updates(data).Error; err != nil {
					fmt.Println("DB error:", err)
				}

				var page int
				for i := range items {
					if items[i].ID == item.ID {
						page = i + 1
						break
					}
				}

				container := getShopData(page, totalPages, data)
				msg := discord.NewMessageUpdateV2(container)

				if err := event.UpdateMessage(msg); err != nil {
					fmt.Println("Error updating shop message:", err)
				}

				followup := discord.NewMessageCreate().WithEphemeral(true).WithContent(":white_check_mark: You just bought this item!")
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

			container := getShopData(page, totalPages, data)

			msg := discord.NewMessageUpdateV2(container)

			if err := event.UpdateMessage(msg); err != nil {
				fmt.Println("Error sending shop message:", err)
			}
		},
	})
}
