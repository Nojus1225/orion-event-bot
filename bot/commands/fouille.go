package commands

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"orion-event-bot/bot/core"
	"orion-event-bot/bot/handlers"
	"orion-event-bot/db"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

func updateLoot(b *core.Bot, data *db.User, field string, value int) int {
	if err := b.DB.GormDB.Model(data).Update(field, value).Error; err != nil {
		fmt.Println("DB error:", err)
	}
	return value
}

func grantLoot(b *core.Bot, data *db.User, idx int, rewards []string, fields []string, values []int) (string, int) {
	itemName := rewards[idx]
	total := updateLoot(b, data, fields[idx], values[idx]+1)
	return itemName, total
}

func applyMalus(b *core.Bot, data *db.User, duration time.Duration, status db.StatusMalus) int {
	data.CDMalus = int(time.Now().Add(duration).Unix())
	data.StatusMalus = status
	if err := b.DB.GormDB.Model(data).Updates(map[string]any{"cd_malus": data.CDMalus, "status_malus": data.StatusMalus}).Error; err != nil {
		fmt.Println("DB error:", err)
	}
	return data.CDMalus
}

func init() {
	handlers.RegisterCommand(handlers.Command{
		Name:        "search",
		Description: "Search the sand",
		Create: discord.SlashCommandCreate{
			Name:        "search",
			Description: "Search the sand",
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			avatarURL := event.User().EffectiveAvatarURL()

			var data db.User
			if err := b.DB.GormDB.FirstOrCreate(&data, db.User{ID: strconv.FormatUint(uint64(event.User().ID), 10)}).Error; err != nil {
				fmt.Println("DB error:", err)
			}

			now := time.Now()

			cdMalus := time.Unix(int64(data.CDMalus), 0)
			if now.Before(cdMalus) {
				txt := "got a sunburn"
				if data.StatusMalus == db.MalusCut {
					txt = "got cut"
				}
				message := fmt.Sprintf("Oh no, you %s come back <t:%d:R>.", txt, cdMalus.Unix())
				msg := discord.NewMessageCreateV2(
					discord.NewContainer(
						discord.NewTextDisplay("# Search"),
						discord.NewTextDisplay(message),
					).WithAccentColor(0xE74C3C),
				)

				if err := event.CreateMessage(msg); err != nil {
					fmt.Println("Error sending cd message:", err)
				}
				return
			}

			cdFouille := time.Unix(int64(data.CDFouille), 0)
			if now.Before(cdFouille) {
				msg := discord.NewMessageCreateV2(
					discord.NewContainer(
						discord.NewTextDisplay("# Search"),
						discord.NewTextDisplay(fmt.Sprintf("You already dug, come back <t:%d:R>.", cdFouille.Unix())),
					).WithAccentColor(0xE74C3C),
				)

				if err := event.CreateMessage(msg); err != nil {
					fmt.Println("Error sending cd message:", err)
				}
				return
			}

			msg := discord.NewMessageCreateV2(
				discord.NewContainer(
					discord.NewSection(
						discord.NewTextDisplay("# Search"),
						discord.NewTextDisplay("Which area would you like to search?"),
					).WithAccessory(discord.NewThumbnail(avatarURL)),
					discord.NewActionRow(
						discord.NewButton(discord.ButtonStylePrimary, "Deckchairs", "search-transats", "", snowflake.ID(0)),
						discord.NewButton(discord.ButtonStylePrimary, "Water's edge", "search-eau", "", snowflake.ID(0)),
						discord.NewButton(discord.ButtonStyleDanger, "The rocks", "search-rochers", "", snowflake.ID(0)),
					),
				).WithAccentColor(0x5865F2),
			)

			if err := event.CreateMessage(msg); err != nil {
				fmt.Println("Error sending search message:", err)
			}

		},
		ExecuteButton: func(b *core.Bot, event *events.ComponentInteractionCreate) {
			button := event.ButtonInteractionData()
			var data db.User
			if err := b.DB.GormDB.FirstOrCreate(&data, db.User{ID: strconv.FormatUint(uint64(event.User().ID), 10)}).Error; err != nil {
				fmt.Println("DB error:", err)
			}

			switch button.CustomID() {
			case "search-transats":
				rewards := []string{"Shell", "Bottle", "Can"}
				fields := []string{"item_coquillage", "item_bouteille", "item_cannette"}
				values := []int{data.ItemCoquillage, data.ItemBouteille, data.ItemCannette}
				random, err := rand.Int(rand.Reader, big.NewInt(100))
				if err != nil {
					fmt.Println("Error generating random reward:", err)
					return
				}
				roll := int(random.Int64()) + 1

				data.CDFouille = int(time.Now().Add(30 * time.Minute).Unix())
				if err := b.DB.GormDB.Model(&data).Update("cd_fouille", data.CDFouille).Error; err != nil {
					fmt.Println("DB error:", err)
				}

				message := "You found nothing while searching the deckchairs."
				if roll <= 85 {
					idx := roll % len(rewards)
					itemName, total := grantLoot(b, &data, idx, rewards, fields, values)
					message = fmt.Sprintf("You found x1 %s (%d) while searching the deckchairs.", itemName, total)
				} else if roll > 98 {
					cd := applyMalus(b, &data, 1*time.Hour, db.MalusSunBurn)
					message = fmt.Sprintf("You got a sunburn. Come back <t:%d:R>.", cd)
				}

				msg := discord.NewMessageUpdateV2(
					discord.NewContainer(
						discord.NewSection(
							discord.NewTextDisplay("# Search"),
							discord.NewTextDisplay(message),
						).WithAccessory(discord.NewThumbnail(event.User().EffectiveAvatarURL())),
					).WithAccentColor(0x5865F2),
				)

				if err := event.UpdateMessage(msg); err != nil {
					fmt.Println("Error sending cd message:", err)
				}
			case "search-eau":
				rewards := []string{"Shell", "Bottle", "Can", "Glasses", "Coin", "Crab Claw"}
				fields := []string{"item_coquillage", "item_bouteille", "item_cannette", "item_lunette", "item_piece", "item_pince"}
				values := []int{data.ItemCoquillage, data.ItemBouteille, data.ItemCannette, data.ItemLunettes, data.ItemPiece, data.ItemPince}
				random, err := rand.Int(rand.Reader, big.NewInt(100))
				if err != nil {
					fmt.Println("Error generating random reward:", err)
					return
				}
				roll := int(random.Int64()) + 1

				data.CDFouille = int(time.Now().Add(30 * time.Minute).Unix())
				if err := b.DB.GormDB.Model(&data).Update("cd_fouille", data.CDFouille).Error; err != nil {
					fmt.Println("DB error:", err)
				}

				message := "You found nothing while searching the water's edge."
				if roll <= 60 {
					idx := roll % len(rewards)
					itemName, total := grantLoot(b, &data, idx, rewards, fields, values)
					message = fmt.Sprintf("You found x1 %s (%d) while searching the water's edge.", itemName, total)
				} else if roll > 80 {
					cd := applyMalus(b, &data, 1*time.Hour, db.MalusSunBurn)
					message = fmt.Sprintf("You got a sunburn. Come back <t:%d:R>.", cd)
				}

				msg := discord.NewMessageUpdateV2(
					discord.NewContainer(
						discord.NewSection(
							discord.NewTextDisplay("# Search"),
							discord.NewTextDisplay(message),
						).WithAccessory(discord.NewThumbnail(event.User().EffectiveAvatarURL())),
					).WithAccentColor(0x5865F2),
				)

				if err := event.UpdateMessage(msg); err != nil {
					fmt.Println("Error sending cd message:", err)
				}

			case "search-rochers":
				rewards := []string{"Glasses", "Coin", "Crab Claw", "Message in a Bottle", "Watch", "Shark Tooth"}
				fields := []string{"item_lunette", "item_piece", "item_pince", "item_msg_bouteille", "item_montre", "item_dent"}
				values := []int{data.ItemLunettes, data.ItemPiece, data.ItemPince, data.ItemMsgBouteille, data.ItemMontre, data.ItemDent}
				random, err := rand.Int(rand.Reader, big.NewInt(100))
				if err != nil {
					fmt.Println("Error generating random reward:", err)
					return
				}
				roll := int(random.Int64()) + 1

				data.CDFouille = int(time.Now().Add(30 * time.Minute).Unix())
				if err := b.DB.GormDB.Model(&data).Update("cd_fouille", data.CDFouille).Error; err != nil {
					fmt.Println("DB error:", err)
				}

				message := "You found nothing while searching the rocks."
				if roll <= 35 {
					idx := roll % len(rewards)
					itemName, total := grantLoot(b, &data, idx, rewards, fields, values)
					message = fmt.Sprintf("You found x1 %s (%d) while searching the rocks.", itemName, total)
				} else if roll <= 75 {
					message = "You found nothing while searching the rocks."
				} else if roll <= 90 {
					cd := applyMalus(b, &data, 1*time.Hour, db.MalusCut)
					message = fmt.Sprintf("You got a sunburn. Come back <t:%d:R>.", cd)
				} else {
					cd := applyMalus(b, &data, 90*time.Minute, db.MalusCut)
					message = fmt.Sprintf("You got cut. Come back <t:%d:R>.", cd)
				}

				msg := discord.NewMessageUpdateV2(
					discord.NewContainer(
						discord.NewSection(
							discord.NewTextDisplay("# Search"),
							discord.NewTextDisplay(message),
						).WithAccessory(discord.NewThumbnail(event.User().EffectiveAvatarURL())),
					).WithAccentColor(0x5865F2),
				)

				if err := event.UpdateMessage(msg); err != nil {
					fmt.Println("Error sending cd message:", err)
				}
			}
		},
	})
}
