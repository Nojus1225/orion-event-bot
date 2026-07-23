package commands

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	"orion-event-bot/bot/core"
	"orion-event-bot/bot/handlers"
	"orion-event-bot/bot/vars"
	"orion-event-bot/db"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var stealCooldowns sync.Map

func getItemName(itemID string, loc *i18n.Localizer) string {
	itemKey := ""
	switch db.Item(itemID) {
	case db.ItemCoquillage: itemKey = "Shell"
	case db.ItemBouteille: itemKey = "Bottle"
	case db.ItemCannette: itemKey = "Can"
	case db.ItemLunettes: itemKey = "Glasses"
	case db.ItemPiece: itemKey = "Coins"
	case db.ItemPince: itemKey = "CrabClaw"
	case db.ItemMsgBouteille: itemKey = "MessageInABottle"
	case db.ItemMontre: itemKey = "Watch"
	case db.ItemDent: itemKey = "SharkTooth"
	default:
		for _, cat := range vars.Shop {
			for _, j := range cat.Items {
				if string(j.ID) == itemID {
					itemKey = j.Name
					break
				}
			}
		}
	}
	
	if itemKey != "" {
		itemLoc, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: itemKey})
		if itemLoc != "" {
			return itemLoc
		}
		return itemKey
	}
	return itemID
}

func init() {
	handlers.RegisterCommand(handlers.Command{
		Name:        "steal",
		Description: "Steal from a player",
		Create: discord.SlashCommandCreate{
			Name:        "steal",
			Description: "Steal from a player",
			DescriptionLocalizations: map[discord.Locale]string{
				discord.LocaleFrench: "Voler un joueur",
			},
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionUser{
					Name:        "user",
					Description: "The targeted user",
					Required:    true,
				},
			},
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			loc := core.GetLocalizer(event.Locale().String())

			attackerID := strconv.FormatUint(uint64(event.User().ID), 10)
			targetUser := event.SlashCommandInteractionData().User("user")
			targetID := targetUser.ID.String()

			if attackerID == targetID {
				msg, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "StealSelfError"})
				_ = event.CreateMessage(discord.NewMessageCreate().WithContent(msg).WithEphemeral(true))
				return
			}

			if targetUser.Bot {
				msg, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "StealBotError"})
				_ = event.CreateMessage(discord.NewMessageCreate().WithContent(msg).WithEphemeral(true))
				return
			}

			var attacker db.User
			if err := b.DB.GormDB.FirstOrCreate(&attacker, db.User{ID: attackerID}).Error; err != nil {
				return
			}
			var attackerItems []db.UserItem
			b.DB.GormDB.Where("user_id = ? AND quantity > 0", attackerID).Find(&attackerItems)
			if len(attackerItems) == 0 {
				msg, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "StealNotEnoughItemsAttacker"})
				if msg == "" {
					msg = "Vous devez avoir au moins un objet pour pouvoir voler quelqu'un !"
				}
				_ = event.CreateMessage(discord.NewMessageCreate().WithContent(msg).WithEphemeral(true))
				return
			}

			var targetItems []db.UserItem
			b.DB.GormDB.Where("user_id = ? AND quantity > 0", targetID).Find(&targetItems)
			if len(targetItems) == 0 {
				msg, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "StealNotEnoughItemsTarget"})
				if msg == "" {
					msg = "Cette personne n'a pas d'objets à voler !"
				}
				_ = event.CreateMessage(discord.NewMessageCreate().WithContent(msg).WithEphemeral(true))
				return
			}

			cooldownKey := attackerID
			if val, ok := stealCooldowns.Load(cooldownKey); ok {
				lastTime := val.(time.Time)
				if time.Since(lastTime) < 30*time.Minute {
					remaining := (30 * time.Minute) - time.Since(lastTime)
					msg, _ := loc.Localize(&i18n.LocalizeConfig{
						MessageID:    "StealCooldownError",
						TemplateData: map[string]interface{}{"Minutes": int(remaining.Minutes())},
					})
					if msg == "" {
						msg = fmt.Sprintf("Vous avez déjà attaqué cette personne récemment. Réessayez dans %d minutes.", int(remaining.Minutes()))
					}
					_ = event.CreateMessage(discord.NewMessageCreate().WithContent(msg).WithEphemeral(true))
					return
				}
			}

			randomNum, _ := rand.Int(rand.Reader, big.NewInt(3))
			correctIdx := int(randomNum.Int64())

			var components []discord.InteractiveComponent
			for i := 0; i < 5; i++ {
				btnTxt, _ := loc.Localize(&i18n.LocalizeConfig{
					MessageID:    "StealPocket",
					TemplateData: map[string]interface{}{"Pocket": i + 1},
				})
				customID := fmt.Sprintf("steal-%s-%s-%d-%d", attackerID, targetID, correctIdx, i)
				components = append(components, discord.NewButton(discord.ButtonStylePrimary, btnTxt, customID, "", snowflake.ID(0)))
			}

			componentsList := []discord.ContainerSubComponent{
				discord.NewSection(
					discord.NewTextDisplay(func() string { s, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "StealTitle"}); return s }()),
					discord.NewTextDisplay(func() string {
						s, _ := loc.Localize(&i18n.LocalizeConfig{
							MessageID:    "StealPrompt",
							TemplateData: map[string]interface{}{"TargetID": targetID},
						})
						return s
					}()),
				).WithAccessory(discord.NewThumbnail("https://raw.githubusercontent.com/googlefonts/noto-emoji/main/png/512/emoji_u1f575.png")),
				discord.NewActionRow(components...),
			}

			msg := discord.NewMessageCreateV2(
				discord.NewContainer(componentsList...).WithAccentColor(0x5865F2),
			)

			stealCooldowns.Store(attackerID, time.Now())

			if err := event.CreateMessage(msg); err != nil {
				fmt.Println("Error sending steal message:", err)
			}
		},
		ExecuteButton: func(b *core.Bot, event *events.ComponentInteractionCreate) {
			customID := event.ButtonInteractionData().CustomID()
			if !strings.HasPrefix(customID, "steal-") {
				return
			}

			parts := strings.Split(customID, "-")
			if len(parts) != 5 {
				return
			}
			attackerID := parts[1]
			targetID := parts[2]
			correctIdx, err := strconv.Atoi(parts[3])
			if err != nil {
				return
			}
			chosenIdx, err := strconv.Atoi(parts[4])
			if err != nil {
				return
			}

			loc := core.GetLocalizer(event.Locale().String())

			userID := strconv.FormatUint(uint64(event.User().ID), 10)
			if userID != attackerID {
				notYourActionTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "NotYourComponent"})
				if notYourActionTxt == "" {
					notYourActionTxt = "Ce n'est pas votre action !"
				}
				_ = event.CreateMessage(discord.NewMessageCreate().WithContent(notYourActionTxt).WithEphemeral(true))
				return
			}

			var attacker db.User
			if err := b.DB.GormDB.FirstOrCreate(&attacker, db.User{ID: attackerID}).Error; err != nil {
				return
			}
			var target db.User
			if err := b.DB.GormDB.FirstOrCreate(&target, db.User{ID: targetID}).Error; err != nil {
				return
			}

			var resultMsg string
			if chosenIdx == correctIdx {
				var targetItems []db.UserItem
				b.DB.GormDB.Where("user_id = ? AND quantity > 0", targetID).Find(&targetItems)
				if len(targetItems) > 0 {
					randomNum, _ := rand.Int(rand.Reader, big.NewInt(int64(len(targetItems))))
					chosenItem := targetItems[randomNum.Int64()].ItemID

					db.RemoveItem(b.DB.GormDB, targetID, db.Item(chosenItem), 1)
					db.IncrementItem(b.DB.GormDB, attackerID, db.Item(chosenItem), 1)

					itemLoc := getItemName(chosenItem, loc)

					resultMsg, _ = loc.Localize(&i18n.LocalizeConfig{
						MessageID:    "StealSuccess",
						TemplateData: map[string]interface{}{"ItemName": itemLoc, "TargetID": targetID},
					})
				} else {
					resultMsg = "Mince, il semblerait que cette personne n'ait plus d'objets !"
				}
			} else {
				var attackerItems []db.UserItem
				b.DB.GormDB.Where("user_id = ? AND quantity > 0", attackerID).Find(&attackerItems)
				if len(attackerItems) > 0 {
					randomNum, _ := rand.Int(rand.Reader, big.NewInt(int64(len(attackerItems))))
					chosenItem := attackerItems[randomNum.Int64()].ItemID

					db.RemoveItem(b.DB.GormDB, attackerID, db.Item(chosenItem), 1)
					db.IncrementItem(b.DB.GormDB, targetID, db.Item(chosenItem), 1)

					itemLoc := getItemName(chosenItem, loc)

					resultMsg, _ = loc.Localize(&i18n.LocalizeConfig{
						MessageID:    "StealFailure",
						TemplateData: map[string]interface{}{"ItemName": itemLoc, "TargetID": targetID},
					})
				} else {
					resultMsg = "Vous avez échoué, mais vous n'aviez plus d'objets à vous faire voler !"
				}
			}

			color := 0x5865F2
			if chosenIdx != correctIdx {
				color = 0xE74C3C
			}

			stealTitle, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "StealTitle"})
			if stealTitle == "" {
				stealTitle = "# 🕵️ Vol"
			}

			errUpdate := event.UpdateMessage(discord.NewMessageUpdateV2(
				discord.NewContainer(
					discord.NewSection(
						discord.NewTextDisplay(stealTitle),
						discord.NewTextDisplay(resultMsg),
					).WithAccessory(discord.NewThumbnail("https://raw.githubusercontent.com/googlefonts/noto-emoji/main/png/512/emoji_u1f575.png")),
				).WithAccentColor(color),
			))
			if errUpdate != nil {
				fmt.Println("Error updating message in steal:", errUpdate)
			}
		},
	})
}
