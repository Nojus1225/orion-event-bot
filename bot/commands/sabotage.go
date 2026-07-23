package commands

import (
	"fmt"
	"strconv"
	"time"

	"orion-event-bot/bot/core"
	"orion-event-bot/bot/handlers"
	"orion-event-bot/bot/vars"
	"orion-event-bot/db"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

func init() {
	handlers.RegisterCommand(handlers.Command{
		Name:        "sabotage",
		Description: "Sabotage a user/zone",
		Create: discord.SlashCommandCreate{
			Name:        "sabotage",
			Description: "Sabotage a user/zone",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionSubCommand{
					Name:        "user",
					Description: "Sabotage a user",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionUser{
							Name:        "user",
							Description: "The targeted user",
							Required:    true,
						},
						discord.ApplicationCommandOptionString{
							Name:         "sabotage",
							Description:  "The sabotage to be inflicted",
							Required:     true,
							Autocomplete: true,
						},
					},
				},
				discord.ApplicationCommandOptionSubCommand{
					Name:        "zone",
					Description: "Sabotage a zone",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionString{
							Name:         "zone",
							Description:  "The targeted zone",
							Required:     true,
							Autocomplete: true,
						},
						discord.ApplicationCommandOptionString{
							Name:         "sabotage",
							Description:  "The sabotage to be inflicted",
							Required:     true,
							Autocomplete: true,
						},
					},
				},
			},
		},
		ExecuteAutocomplete: func(b *core.Bot, event *events.AutocompleteInteractionCreate) {
			loc := core.GetLocalizer(event.Locale().String())
			var choices []discord.AutocompleteChoice

			subCommand := event.Data.SubCommandName
			if subCommand == nil {
				return
			}

			var data db.User
			if err := b.DB.GormDB.FirstOrCreate(&data, db.User{ID: strconv.FormatUint(uint64(event.User().ID), 10)}).Error; err != nil {
				fmt.Println("DB error:", err)
			}

			switch *subCommand {
			case "user":
				for _, option := range event.Data.Options {
					if option.Name == "sabotage" && option.Focused {
						items := []db.Item{db.ShopVolDePelle, db.ShopCrabeDePoche}
						for _, itemID := range items {
							qty := db.GetItemQuantity(b.DB.GormDB, data.ID, itemID)
							if qty > 0 {
								item := vars.GetItem(itemID)
								if item != nil {
									itemLoc, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: item.Name})
									if itemLoc == "" {
										itemLoc = item.Name
									}
									label := fmt.Sprintf("%s (%d)", itemLoc, qty)
									choices = append(choices, discord.AutocompleteChoiceString{
										Name:  label,
										Value: string(itemID),
									})
								}
							}
						}
					}
				}
			case "zone":
				for _, option := range event.Data.Options {
					if option.Name == "zone" && option.Focused {
						zonesN := []string{"Deckchairs", "WatersEdge", "TheRocks"}
						var zones []db.Zone
						for _, n := range zonesN {
							zones = append(zones, db.GetZone(b.DB.GormDB, n))
						}

						now := time.Now().Unix()
						for _, zone := range zones {
							zoneLoc, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: zone.Name})
							if zoneLoc == "" {
								zoneLoc = zone.Name
							}

							if zone.SabotageExpiresAt > now || zone.ImmunityExpiresAt > now {

							} else {
								choices = append(choices, discord.AutocompleteChoiceString{
									Name:  zoneLoc,
									Value: zone.Name,
								})
							}
						}
					} else if option.Name == "sabotage" && option.Focused {
						items := []db.Item{db.ShopAlerteMeduses, db.ShopVoleurDeFilet}
						for _, itemID := range items {
							qty := db.GetItemQuantity(b.DB.GormDB, data.ID, itemID)
							if qty > 0 {
								item := vars.GetItem(itemID)
								if item != nil {
									itemLoc, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: item.Name})
									if itemLoc == "" {
										itemLoc = item.Name
									}
									label := fmt.Sprintf("%s (%d)", itemLoc, qty)
									choices = append(choices, discord.AutocompleteChoiceString{
										Name:  label,
										Value: string(itemID),
									})
								}
							}
						}
					}
				}
			}

			if err := event.AutocompleteResult(choices); err != nil {
				fmt.Println("Error sending autocomplete message:", err)
			}
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			loc := core.GetLocalizer(event.Locale().String())
			cmdData := event.SlashCommandInteractionData()
			if cmdData.SubCommandName == nil {
				return
			}

			subCommand := *cmdData.SubCommandName
			attackerID := strconv.FormatUint(uint64(event.User().ID), 10)

			var attacker db.User
			if err := b.DB.GormDB.FirstOrCreate(&attacker, db.User{ID: attackerID}).Error; err != nil {
				return
			}

			var sabotageType db.Item
			var responseMessageID string
			var publicAlertID string
			var targetID string
			now := time.Now().Unix()

			switch subCommand {
			case "user":
				targetID = cmdData.User("user").ID.String()
				sabotageType = db.Item(cmdData.String("sabotage"))

				if db.GetItemQuantity(b.DB.GormDB, attackerID, sabotageType) <= 0 {
					responseMessageID = "SabotageNoItem"
				} else {
					var target db.User
					b.DB.GormDB.FirstOrCreate(&target, db.User{ID: targetID})

					if target.SabotageImmunityExpiresAt > now {
						responseMessageID = "SabotageImmune"
						db.LogAction(b.DB.GormDB, attackerID, db.ActionSabotage, fmt.Sprintf(`{"target":"%s", "sabotage":"%s", "result":"immune"}`, targetID, sabotageType))
					} else if db.GetItemQuantity(b.DB.GormDB, targetID, db.ShopCremeSolaire1000) > 0 {
						db.RemoveItem(b.DB.GormDB, targetID, db.ShopCremeSolaire1000, 1)
						db.RemoveItem(b.DB.GormDB, attackerID, sabotageType, 1)
						responseMessageID = "SabotageBlockedByCreme"
						publicAlertID = "SabotageAlertBlocked"
						db.LogAction(b.DB.GormDB, attackerID, db.ActionSabotage, fmt.Sprintf(`{"target":"%s", "sabotage":"%s", "result":"blocked_by_creme"}`, targetID, sabotageType))
						db.LogAction(b.DB.GormDB, targetID, db.ActionProtect, fmt.Sprintf(`{"attacker":"%s", "sabotage":"%s", "result":"creme_consumed_passive"}`, attackerID, sabotageType))
					} else {
						switch sabotageType {
						case db.ShopVolDePelle:
							target.StatusMalus = db.MalusShovelTheft
							if int64(target.CDMalus) > now {
								target.CDMalus = int(int64(target.CDMalus) + 15*60)
							} else {
								target.CDMalus = int(now + 15*60)
							}
							target.CDFouille = 0
							publicAlertID = "SabotageAlertShovel"
						case db.ShopCrabeDePoche:
							target.StatusMalus = db.MalusCut
							target.CDMalus = int(now + 90*60)
							target.CDFouille = 0
							publicAlertID = "SabotageAlertCrab"
						}

						target.SabotageImmunityExpiresAt = now + 3600
						b.DB.GormDB.Save(&target)

						db.RemoveItem(b.DB.GormDB, attackerID, sabotageType, 1)
						responseMessageID = "SabotageSuccessUser"
						db.LogAction(b.DB.GormDB, attackerID, db.ActionSabotage, fmt.Sprintf(`{"target":"%s", "sabotage":"%s", "result":"success"}`, targetID, sabotageType))
					}
				}
			case "zone":
				zoneName := cmdData.String("zone")
				sabotageType = db.Item(cmdData.String("sabotage"))

				if db.GetItemQuantity(b.DB.GormDB, attackerID, sabotageType) <= 0 {
					responseMessageID = "SabotageNoItem"
				} else {
					zone := db.GetZone(b.DB.GormDB, zoneName)

					if zone.SabotageExpiresAt > now || zone.ImmunityExpiresAt > now {
						responseMessageID = "SabotageZoneImmune"
						db.LogAction(b.DB.GormDB, attackerID, db.ActionSabotage, fmt.Sprintf(`{"zone":"%s", "sabotage":"%s", "result":"immune"}`, zoneName, sabotageType))
					} else {
						var activeSabotages int64
						b.DB.GormDB.Model(&db.Zone{}).Where("sabotage_expires_at > ?", now).Count(&activeSabotages)

						if activeSabotages > 0 {
							responseMessageID = "SabotageZoneImmune"
							db.LogAction(b.DB.GormDB, attackerID, db.ActionSabotage, fmt.Sprintf(`{"zone":"%s", "sabotage":"%s", "result":"immune_other_active"}`, zoneName, sabotageType))
						} else {
							if sabotageType == db.ShopAlerteMeduses {
								sabotageDuration := int64(3 * 3600)
								immunityDuration := int64(4 * 3600)
								db.SetZoneSabotage(b.DB.GormDB, zoneName, string(sabotageType), now+sabotageDuration, now+sabotageDuration+immunityDuration, attackerID)
							} else if sabotageType == db.ShopVoleurDeFilet {
								sabotageDuration := int64(5 * 3600)
								db.SetZoneSabotage(b.DB.GormDB, zoneName, string(sabotageType), now+sabotageDuration, 0, attackerID)
							}

							db.RemoveItem(b.DB.GormDB, attackerID, sabotageType, 1)
							responseMessageID = "SabotageSuccessZone"
							db.LogAction(b.DB.GormDB, attackerID, db.ActionSabotage, fmt.Sprintf(`{"zone":"%s", "sabotage":"%s", "result":"success"}`, zoneName, sabotageType))
						}
					}
				}
			}

			msgStr, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: responseMessageID})
			if msgStr == "" {
				msgStr = responseMessageID
			}

			msg := discord.NewMessageCreateV2(
				discord.NewContainer(
					discord.NewTextDisplay(msgStr),
				).WithAccentColor(0x5865F2),
			).WithEphemeral(true)

			err := event.CreateMessage(msg)
			if err != nil {
				fmt.Println("Error sending sabotage message:", err)
			}

			if publicAlertID != "" {
				alertMsg, _ := loc.Localize(&i18n.LocalizeConfig{
					MessageID: publicAlertID,
					TemplateData: map[string]interface{}{
						"TargetID":   targetID,
						"AttackerID": attackerID,
					},
				})
				if alertMsg == "" {
					alertMsg = publicAlertID
				}
				_, errAlert := event.Client().Rest.CreateMessage(event.Channel().ID(), discord.NewMessageCreate().WithContent(alertMsg))
				if errAlert != nil {
					fmt.Println("Error sending sabotage alert message:", errAlert)
				}
			}
		},
	})
}
