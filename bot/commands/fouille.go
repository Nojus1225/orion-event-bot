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

func grantLoot(b *core.Bot, data *db.User, idx int, rewards []string, items []db.Item) (string, int) {
	itemName := rewards[idx]
	itemID := items[idx]
	total := db.IncrementItem(b.DB.GormDB, data.ID, itemID, 1)

	db.IncrementStat(b.DB.GormDB, data.ID, db.StatObjetsTrouve, 1)

	switch itemID {
	case db.ItemCoquillage:
		db.IncrementStat(b.DB.GormDB, data.ID, db.StatCoquillagesTrouves, 1)
	case db.ItemBouteille:
		db.IncrementStat(b.DB.GormDB, data.ID, db.StatBouteillesTrouvees, 1)
	case db.ItemCannette:
		db.IncrementStat(b.DB.GormDB, data.ID, db.StatCannettesTrouvees, 1)
	case db.ItemLunettes:
		db.IncrementStat(b.DB.GormDB, data.ID, db.StatLunettesTrouvees, 1)
	case db.ItemPiece:
		db.IncrementStat(b.DB.GormDB, data.ID, db.StatPiecesTrouvees, 1)
	case db.ItemPince:
		db.IncrementStat(b.DB.GormDB, data.ID, db.StatPincesTrouvees, 1)
	case db.ItemMsgBouteille:
		db.IncrementStat(b.DB.GormDB, data.ID, db.StatMsgBouteilleTrouves, 1)
	case db.ItemMontre:
		db.IncrementStat(b.DB.GormDB, data.ID, db.StatMontresTrouvees, 1)
	case db.ItemDent:
		db.IncrementStat(b.DB.GormDB, data.ID, db.StatDentsTrouvees, 1)
	}

	db.LogAction(b.DB.GormDB, data.ID, db.ActionFouille, fmt.Sprintf(`{"result":"item","item":"%s"}`, itemName))

	return itemName, total
}

func applyMalus(b *core.Bot, data *db.User, duration time.Duration, status db.StatusMalus) int {
	data.CDMalus = int(time.Now().Add(duration).Unix())
	data.StatusMalus = status

	updates := map[string]any{"cd_malus": data.CDMalus, "status_malus": data.StatusMalus}
	if err := b.DB.GormDB.Model(data).Updates(updates).Error; err != nil {
		fmt.Println("DB error:", err)
	}

	if status == db.MalusSunBurn {
		db.IncrementStat(b.DB.GormDB, data.ID, db.StatTotalSunburn, 1)
	} else if status == db.MalusCut {
		db.IncrementStat(b.DB.GormDB, data.ID, db.StatTotalCuts, 1)
	}

	if err := b.DB.GormDB.Model(data).Updates(updates).Error; err != nil {
		fmt.Println("DB error:", err)
	}
	db.LogAction(b.DB.GormDB, data.ID, db.ActionMalus, fmt.Sprintf(`{"status":"%s"}`, string(status)))
	return data.CDMalus
}

func init() {
	handlers.RegisterCommand(handlers.Command{
		Name:        "search",
		Description: "Search the sand",
		Create: discord.SlashCommandCreate{
			Name:        "search",
			Description: "Search the sand",

			DescriptionLocalizations: map[discord.Locale]string{
				discord.LocaleFrench: "Fouiller le sable",
			},
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			loc := core.GetLocalizer(event.Locale().String())

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
					if Items[i].ID == object {
						item = &Items[i]
						break
					}
				}

				hasItem := db.GetItemQuantity(b.DB.GormDB, data.ID, object) > 0

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
						discord.NewButton(discord.ButtonStyleSuccess, useBtnTxt, fmt.Sprintf("search-use-item-%s", object), "", snowflake.ID(0)),
					))
				}

				msg := discord.NewMessageCreateV2(
					discord.NewContainer(components...).WithAccentColor(0xE74C3C),
				)

				if err := event.CreateMessage(msg); err != nil {
					fmt.Println("Error sending cd message:", err)
				}
				db.IncrementStat(b.DB.GormDB, data.ID, db.StatEchecsFouilleMalus, 1)
				db.LogAction(b.DB.GormDB, data.ID, db.ActionFouille, `{"result":"failed_malus"}`)
				return
			}

			cdFouille := time.Unix(int64(data.CDFouille), 0)
			if now.Before(cdFouille) {
				object := db.ShopBouteilleEau
				var item *ShopItem
				for i := range Items {
					if Items[i].ID == object {
						item = &Items[i]
						break
					}
				}

				hasItem := db.GetItemQuantity(b.DB.GormDB, data.ID, object) > 0

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
						discord.NewButton(discord.ButtonStyleSuccess, useBtnTxt, fmt.Sprintf("search-use-item-%s", object), "", snowflake.ID(0)),
					))
				}

				msg := discord.NewMessageCreateV2(
					discord.NewContainer(components...).WithAccentColor(0xE74C3C),
				)

				if err := event.CreateMessage(msg); err != nil {
					fmt.Println("Error sending cd message:", err)
				}
				db.IncrementStat(b.DB.GormDB, data.ID, db.StatEchecsFouilleCooldown, 1)
				db.LogAction(b.DB.GormDB, data.ID, db.ActionFouille, `{"result":"failed_cooldown"}`)
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
					).WithAccessory(discord.NewThumbnail("https://raw.githubusercontent.com/googlefonts/noto-emoji/main/png/512/emoji_u1f3d6.png")),
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
				object := db.Item(customID[16:])
				hasItem := db.GetItemQuantity(b.DB.GormDB, data.ID, object) > 0
				if hasItem {
					db.RemoveItem(b.DB.GormDB, data.ID, object, 1)

					switch object {
					case db.ShopCremeSolaire, db.ShopPansement:
						data.CDMalus = 0
						data.StatusMalus = db.MalusNone
						if err := b.DB.GormDB.Model(&data).Select("CDMalus", "StatusMalus").Updates(data).Error; err != nil {
							fmt.Println("DB error:", err)
						}
					case db.ShopBouteilleEau:
						data.CDFouille = 0
						if err := b.DB.GormDB.Model(&data).Select("CDFouille").Updates(data).Error; err != nil {
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
								).WithAccessory(discord.NewThumbnail("https://raw.githubusercontent.com/googlefonts/noto-emoji/main/png/512/emoji_u1f3d6.png")),
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

			switch customID {
			case "search-transats":
				rewards := []string{"Shell", "Bottle", "Can"}
				items := []db.Item{db.ItemCoquillage, db.ItemBouteille, db.ItemCannette}
				random, err := rand.Int(rand.Reader, big.NewInt(100))
				if err != nil {
					fmt.Println("Error generating random reward:", err)
					return
				}
				roll := int(random.Int64()) + 1

				data.CDFouille = int(time.Now().Add(30 * time.Minute).Unix())
				if err := b.DB.GormDB.Model(&data).Updates(map[string]any{"cd_fouille": data.CDFouille}).Error; err != nil {
					fmt.Println("DB error:", err)
				}
				db.IncrementStat(b.DB.GormDB, data.ID, db.StatFouille, 1)
				db.IncrementStat(b.DB.GormDB, data.ID, db.StatFouilleTransats, 1)

				message, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "SearchDeckchairsEmpty"})
				if roll <= 85 {
					idx := roll % len(rewards)
					itemName, total := grantLoot(b, &data, idx, rewards, items)
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
					db.IncrementStat(b.DB.GormDB, data.ID, db.StatTotalRien, 1)
					db.LogAction(b.DB.GormDB, data.ID, db.ActionFouille, `{"result":"rien"}`)
				}

				searchTitle, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "SearchTitle"})
				msg := discord.NewMessageUpdateV2(
					discord.NewContainer(
						discord.NewSection(
							discord.NewTextDisplay(searchTitle),
							discord.NewTextDisplay(message),
						).WithAccessory(discord.NewThumbnail("https://raw.githubusercontent.com/googlefonts/noto-emoji/main/png/512/emoji_u1fa91.png")),
					).WithAccentColor(0x5865F2),
				)

				if err := event.UpdateMessage(msg); err != nil {
					fmt.Println("Error sending cd message:", err)
				}
			case "search-eau":
				rewards := []string{"Shell", "Bottle", "Can", "Glasses", "Coin", "Crab Claw"}
				items := []db.Item{db.ItemCoquillage, db.ItemBouteille, db.ItemCannette, db.ItemLunettes, db.ItemPiece, db.ItemPince}
				random, err := rand.Int(rand.Reader, big.NewInt(100))
				if err != nil {
					fmt.Println("Error generating random reward:", err)
					return
				}
				roll := int(random.Int64()) + 1

				data.CDFouille = int(time.Now().Add(30 * time.Minute).Unix())
				if err := b.DB.GormDB.Model(&data).Updates(map[string]any{"cd_fouille": data.CDFouille}).Error; err != nil {
					fmt.Println("DB error:", err)
				}
				db.IncrementStat(b.DB.GormDB, data.ID, db.StatFouille, 1)
				db.IncrementStat(b.DB.GormDB, data.ID, db.StatFouilleEau, 1)

				message, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "SearchWaterEmpty"})
				if roll <= 60 {
					idx := roll % len(rewards)
					itemName, total := grantLoot(b, &data, idx, rewards, items)
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
					db.IncrementStat(b.DB.GormDB, data.ID, db.StatTotalRien, 1)
					db.LogAction(b.DB.GormDB, data.ID, db.ActionFouille, `{"result":"rien"}`)
				}

				searchTitle, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "SearchTitle"})
				msg := discord.NewMessageUpdateV2(
					discord.NewContainer(
						discord.NewSection(
							discord.NewTextDisplay(searchTitle),
							discord.NewTextDisplay(message),
						).WithAccessory(discord.NewThumbnail("https://raw.githubusercontent.com/googlefonts/noto-emoji/main/png/512/emoji_u1f30a.png")),
					).WithAccentColor(0x5865F2),
				)

				if err := event.UpdateMessage(msg); err != nil {
					fmt.Println("Error sending cd message:", err)
				}

			case "search-rochers":
				rewards := []string{"Glasses", "Coin", "Crab Claw", "Message in a Bottle", "Watch", "Shark Tooth"}
				items := []db.Item{db.ItemLunettes, db.ItemPiece, db.ItemPince, db.ItemMsgBouteille, db.ItemMontre, db.ItemDent}
				random, err := rand.Int(rand.Reader, big.NewInt(100))
				if err != nil {
					fmt.Println("Error generating random reward:", err)
					return
				}
				roll := int(random.Int64()) + 1

				data.CDFouille = int(time.Now().Add(30 * time.Minute).Unix())
				if err := b.DB.GormDB.Model(&data).Updates(map[string]any{"cd_fouille": data.CDFouille}).Error; err != nil {
					fmt.Println("DB error:", err)
				}
				db.IncrementStat(b.DB.GormDB, data.ID, db.StatFouille, 1)
				db.IncrementStat(b.DB.GormDB, data.ID, db.StatFouilleRochers, 1)

				message, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "SearchRocksEmpty"})
				if roll <= 35 {
					idx := roll % len(rewards)
					itemName, total := grantLoot(b, &data, idx, rewards, items)
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
					db.IncrementStat(b.DB.GormDB, data.ID, db.StatTotalRien, 1)
					db.LogAction(b.DB.GormDB, data.ID, db.ActionFouille, `{"result":"rien"}`)
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
						).WithAccessory(discord.NewThumbnail("https://raw.githubusercontent.com/googlefonts/noto-emoji/main/png/512/emoji_u1faa8.png")),
					).WithAccentColor(0x5865F2),
				)

				if err := event.UpdateMessage(msg); err != nil {
					fmt.Println("Error sending cd message:", err)
				}
			}
		},
	})
}
