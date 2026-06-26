package commands

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"orion-event-bot/bot/core"
	"orion-event-bot/bot/handlers"
	"orion-event-bot/db"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"github.com/nicksnyder/go-i18n/v2/i18n"
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

	data.StatsObjetsTrouve++
	if err := b.DB.GormDB.Model(data).Update("stats_objets_trouve", data.StatsObjetsTrouve).Error; err != nil {
		fmt.Println("DB error:", err)
	}

	return itemName, total
}

func applyMalus(b *core.Bot, data *db.User, duration time.Duration, status db.StatusMalus) int {
	data.CDMalus = int(time.Now().Add(duration).Unix())
	data.StatusMalus = status

	updates := map[string]any{"cd_malus": data.CDMalus, "status_malus": data.StatusMalus}
	if status == db.MalusSunBurn {
		data.StatsTotalSunburn++
		updates["stats_total_sunburn"] = data.StatsTotalSunburn
	} else if status == db.MalusCut {
		data.StatsTotalCuts++
		updates["stats_total_cuts"] = data.StatsTotalCuts
	}

	if err := b.DB.GormDB.Model(data).Updates(updates).Error; err != nil {
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
			loc := core.GetLocalizer(event.Locale().String())
			avatarURL := event.User().EffectiveAvatarURL()

			var data db.User
			if err := b.DB.GormDB.FirstOrCreate(&data, db.User{ID: strconv.FormatUint(uint64(event.User().ID), 10)}).Error; err != nil {
				fmt.Println("DB error:", err)
			}

			now := time.Now()

			cdMalus := time.Unix(int64(data.CDMalus), 0)
			if now.Before(cdMalus) {
				txt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "SearchSunburn"})
				object := db.ShopCremeSolaire
				if data.StatusMalus == db.MalusCut {
					txt, _ = loc.Localize(&i18n.LocalizeConfig{MessageID: "SearchCut"})
					object = db.ShopPansement
				}

				var item *ShopItem
				for i := range Items {
					if int(Items[i].ID) == int(object) {
						item = &Items[i]
						break
					}
				}

				hasItem := false
				for _, invItem := range data.Inventory {
					if invItem == int(object) {
						hasItem = true
						break
					}
				}

				itemLoc, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: item.Name})
				if itemLoc == "" {
					itemLoc = item.Name
				}

				message, _ := loc.Localize(&i18n.LocalizeConfig{
					MessageID: "SearchMalusPrompt",
					TemplateData: map[string]interface{}{
						"Status": txt,
						"Time":   cdMalus.Unix(),
						"Item":   itemLoc,
					},
				})

				searchTitle, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "SearchTitle"})
				useBtnTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "Use"})

				components := []discord.ContainerSubComponent{
					discord.NewTextDisplay(searchTitle),
					discord.NewTextDisplay(message),
				}
				if hasItem {
					components = append(components, discord.NewActionRow(
						discord.NewButton(discord.ButtonStyleSuccess, useBtnTxt, fmt.Sprintf("search-use-item-%d", object), "", snowflake.ID(0)),
					))
				}

				msg := discord.NewMessageCreateV2(
					discord.NewContainer(components...).WithAccentColor(0xE74C3C),
				)

				if err := event.CreateMessage(msg); err != nil {
					fmt.Println("Error sending cd message:", err)
				}
				return
			}

			cdFouille := time.Unix(int64(data.CDFouille), 0)
			if now.Before(cdFouille) {
				object := db.ShopBouteilleEau
				var item *ShopItem
				for i := range Items {
					if int(Items[i].ID) == int(object) {
						item = &Items[i]
						break
					}
				}

				hasItem := false
				for _, invItem := range data.Inventory {
					if invItem == int(object) {
						hasItem = true
						break
					}
				}

				itemLoc, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: item.Name})
				if itemLoc == "" {
					itemLoc = item.Name
				}

				message, _ := loc.Localize(&i18n.LocalizeConfig{
					MessageID: "SearchCooldownPrompt",
					TemplateData: map[string]interface{}{
						"Time": cdFouille.Unix(),
						"Item": itemLoc,
					},
				})

				searchTitle, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "SearchTitle"})
				useBtnTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "Use"})

				components := []discord.ContainerSubComponent{
					discord.NewTextDisplay(searchTitle),
					discord.NewTextDisplay(message),
				}
				if hasItem {
					components = append(components, discord.NewActionRow(
						discord.NewButton(discord.ButtonStyleSuccess, useBtnTxt, fmt.Sprintf("search-use-item-%d", object), "", snowflake.ID(0)),
					))
				}

				msg := discord.NewMessageCreateV2(
					discord.NewContainer(components...).WithAccentColor(0xE74C3C),
				)

				if err := event.CreateMessage(msg); err != nil {
					fmt.Println("Error sending cd message:", err)
				}
				return
			}

			searchTitle, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "SearchTitle"})
			areaPrompt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "SearchAreaPrompt"})
			deckchairsTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "Deckchairs"})
			watersEdgeTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "WatersEdge"})
			theRocksTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "TheRocks"})

			msg := discord.NewMessageCreateV2(
				discord.NewContainer(
					discord.NewSection(
						discord.NewTextDisplay(searchTitle),
						discord.NewTextDisplay(areaPrompt),
					).WithAccessory(discord.NewThumbnail(avatarURL)),
					discord.NewActionRow(
						discord.NewButton(discord.ButtonStylePrimary, deckchairsTxt, "search-transats", "", snowflake.ID(0)),
						discord.NewButton(discord.ButtonStylePrimary, watersEdgeTxt, "search-eau", "", snowflake.ID(0)),
						discord.NewButton(discord.ButtonStyleDanger, theRocksTxt, "search-rochers", "", snowflake.ID(0)),
					),
				).WithAccentColor(0x5865F2),
			)

			if err := event.CreateMessage(msg); err != nil {
				fmt.Println("Error sending search message:", err)
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
			button := event.ButtonInteractionData()
			var data db.User
			if err := b.DB.GormDB.FirstOrCreate(&data, db.User{ID: strconv.FormatUint(uint64(event.User().ID), 10)}).Error; err != nil {
				fmt.Println("DB error:", err)
			}

			customID := button.CustomID()

			if strings.HasPrefix(customID, "search-use-item-") {
				itemIdx, err := strconv.Atoi(customID[16:])
				if err == nil {
					object := db.Item(itemIdx)
					hasItem := false
					itemIndex := -1
					for i, invItem := range data.Inventory {
						if invItem == int(object) {
							hasItem = true
							itemIndex = i
							break
						}
					}
					if hasItem {
						data.Inventory = append(data.Inventory[:itemIndex], data.Inventory[itemIndex+1:]...)

						switch object {
						case db.ShopCremeSolaire, db.ShopPansement:
							data.CDMalus = 0
							data.StatusMalus = db.MalusNone
							if err := b.DB.GormDB.Model(&data).Select("Inventory", "CDMalus", "StatusMalus").Updates(data).Error; err != nil {
								fmt.Println("DB error:", err)
							}
						case db.ShopBouteilleEau:
							data.CDFouille = 0
							if err := b.DB.GormDB.Model(&data).Select("Inventory", "CDFouille").Updates(data).Error; err != nil {
								fmt.Println("DB error:", err)
							}
						default:
							if err := b.DB.GormDB.Model(&data).Select("Inventory").Updates(data).Error; err != nil {
								fmt.Println("DB error:", err)
							}
						}

						searchTitle, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "SearchTitle"})
						areaPrompt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "SearchAreaPrompt"})
						deckchairsTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "Deckchairs"})
						watersEdgeTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "WatersEdge"})
						theRocksTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "TheRocks"})

						msg := discord.NewMessageUpdateV2(
							discord.NewContainer(
								discord.NewSection(
									discord.NewTextDisplay(searchTitle),
									discord.NewTextDisplay(areaPrompt),
								).WithAccessory(discord.NewThumbnail(event.User().EffectiveAvatarURL())),
								discord.NewActionRow(
									discord.NewButton(discord.ButtonStylePrimary, deckchairsTxt, "search-transats", "", snowflake.ID(0)),
									discord.NewButton(discord.ButtonStylePrimary, watersEdgeTxt, "search-eau", "", snowflake.ID(0)),
									discord.NewButton(discord.ButtonStyleDanger, theRocksTxt, "search-rochers", "", snowflake.ID(0)),
								),
							).WithAccentColor(0x5865F2),
						)

						if err := event.UpdateMessage(msg); err != nil {
							fmt.Println("Error updating search message:", err)
						}
						return
					}
				}
			}

			switch customID {
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
				data.StatsFouille++
				if err := b.DB.GormDB.Model(&data).Updates(map[string]any{"cd_fouille": data.CDFouille, "stats_fouille": data.StatsFouille}).Error; err != nil {
					fmt.Println("DB error:", err)
				}

				message, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "SearchDeckchairsEmpty"})
				if roll <= 85 {
					idx := roll % len(rewards)
					itemName, total := grantLoot(b, &data, idx, rewards, fields, values)
					itemLoc, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: itemName})
					if itemLoc == "" {
						itemLoc = itemName
					}
					message, _ = loc.Localize(&i18n.LocalizeConfig{
						MessageID: "SearchDeckchairsFound",
						TemplateData: map[string]interface{}{"Item": itemLoc, "Total": total},
					})
				} else if roll > 98 {
					cd := applyMalus(b, &data, 1*time.Hour, db.MalusSunBurn)
					message, _ = loc.Localize(&i18n.LocalizeConfig{
						MessageID: "SearchDeckchairsSunburn",
						TemplateData: map[string]interface{}{"Time": cd},
					})
				} else {
					data.StatsTotalRien++
					if err := b.DB.GormDB.Model(&data).Update("stats_total_rien", data.StatsTotalRien).Error; err != nil {
						fmt.Println("DB error:", err)
					}
				}

				searchTitle, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "SearchTitle"})
				msg := discord.NewMessageUpdateV2(
					discord.NewContainer(
						discord.NewSection(
							discord.NewTextDisplay(searchTitle),
							discord.NewTextDisplay(message),
						).WithAccessory(discord.NewThumbnail(event.User().EffectiveAvatarURL())),
					).WithAccentColor(0x5865F2),
				)

				if err := event.UpdateMessage(msg); err != nil {
					fmt.Println("Error sending cd message:", err)
				}
			case "search-eau":
				rewards := []string{"Shell", "Bottle", "Can", "Glasses", "Coin", "Crab Claw"}
				fields := []string{"item_coquillage", "item_bouteille", "item_cannette", "item_lunettes", "item_piece", "item_pince"}
				values := []int{data.ItemCoquillage, data.ItemBouteille, data.ItemCannette, data.ItemLunettes, data.ItemPiece, data.ItemPince}
				random, err := rand.Int(rand.Reader, big.NewInt(100))
				if err != nil {
					fmt.Println("Error generating random reward:", err)
					return
				}
				roll := int(random.Int64()) + 1

				data.CDFouille = int(time.Now().Add(30 * time.Minute).Unix())
				data.StatsFouille++
				if err := b.DB.GormDB.Model(&data).Updates(map[string]any{"cd_fouille": data.CDFouille, "stats_fouille": data.StatsFouille}).Error; err != nil {
					fmt.Println("DB error:", err)
				}

				message, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "SearchWaterEmpty"})
				if roll <= 60 {
					idx := roll % len(rewards)
					itemName, total := grantLoot(b, &data, idx, rewards, fields, values)
					itemLoc, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: itemName})
					if itemLoc == "" {
						itemLoc = itemName
					}
					message, _ = loc.Localize(&i18n.LocalizeConfig{
						MessageID: "SearchWaterFound",
						TemplateData: map[string]interface{}{"Item": itemLoc, "Total": total},
					})
				} else if roll > 80 {
					cd := applyMalus(b, &data, 1*time.Hour, db.MalusSunBurn)
					message, _ = loc.Localize(&i18n.LocalizeConfig{
						MessageID: "SearchDeckchairsSunburn",
						TemplateData: map[string]interface{}{"Time": cd},
					})
				} else {
					data.StatsTotalRien++
					if err := b.DB.GormDB.Model(&data).Update("stats_total_rien", data.StatsTotalRien).Error; err != nil {
						fmt.Println("DB error:", err)
					}
				}

				searchTitle, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "SearchTitle"})
				msg := discord.NewMessageUpdateV2(
					discord.NewContainer(
						discord.NewSection(
							discord.NewTextDisplay(searchTitle),
							discord.NewTextDisplay(message),
						).WithAccessory(discord.NewThumbnail(event.User().EffectiveAvatarURL())),
					).WithAccentColor(0x5865F2),
				)

				if err := event.UpdateMessage(msg); err != nil {
					fmt.Println("Error sending cd message:", err)
				}

			case "search-rochers":
				rewards := []string{"Glasses", "Coin", "Crab Claw", "Message in a Bottle", "Watch", "Shark Tooth"}
				fields := []string{"item_lunettes", "item_piece", "item_pince", "item_msg_bouteille", "item_montre", "item_dent"}
				values := []int{data.ItemLunettes, data.ItemPiece, data.ItemPince, data.ItemMsgBouteille, data.ItemMontre, data.ItemDent}
				random, err := rand.Int(rand.Reader, big.NewInt(100))
				if err != nil {
					fmt.Println("Error generating random reward:", err)
					return
				}
				roll := int(random.Int64()) + 1

				data.CDFouille = int(time.Now().Add(30 * time.Minute).Unix())
				data.StatsFouille++
				if err := b.DB.GormDB.Model(&data).Updates(map[string]any{"cd_fouille": data.CDFouille, "stats_fouille": data.StatsFouille}).Error; err != nil {
					fmt.Println("DB error:", err)
				}

				message, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "SearchRocksEmpty"})
				if roll <= 35 {
					idx := roll % len(rewards)
					itemName, total := grantLoot(b, &data, idx, rewards, fields, values)
					itemLoc, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: itemName})
					if itemLoc == "" {
						itemLoc = itemName
					}
					message, _ = loc.Localize(&i18n.LocalizeConfig{
						MessageID: "SearchRocksFound",
						TemplateData: map[string]interface{}{"Item": itemLoc, "Total": total},
					})
				} else if roll <= 75 {
					message, _ = loc.Localize(&i18n.LocalizeConfig{MessageID: "SearchRocksEmpty"})
					data.StatsTotalRien++
					if err := b.DB.GormDB.Model(&data).Update("stats_total_rien", data.StatsTotalRien).Error; err != nil {
						fmt.Println("DB error:", err)
					}
				} else if roll <= 90 {
					cd := applyMalus(b, &data, 1*time.Hour, db.MalusCut)
					message, _ = loc.Localize(&i18n.LocalizeConfig{
						MessageID: "SearchDeckchairsSunburn",
						TemplateData: map[string]interface{}{"Time": cd},
					})
				} else {
					cd := applyMalus(b, &data, 90*time.Minute, db.MalusCut)
					message, _ = loc.Localize(&i18n.LocalizeConfig{
						MessageID: "SearchRocksCut",
						TemplateData: map[string]interface{}{"Time": cd},
					})
				}

				searchTitle, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "SearchTitle"})
				msg := discord.NewMessageUpdateV2(
					discord.NewContainer(
						discord.NewSection(
							discord.NewTextDisplay(searchTitle),
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
