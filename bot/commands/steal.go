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
	"orion-event-bot/db"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var stealCooldowns sync.Map

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
			if attacker.Coins <= 0 {
				msg, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "StealNotEnoughCoinsAttacker"})
				if msg == "" {
					msg = "Vous devez avoir des pièces pour pouvoir voler quelqu'un !"
				}
				_ = event.CreateMessage(discord.NewMessageCreate().WithContent(msg).WithEphemeral(true))
				return
			}

			var target db.User
			if err := b.DB.GormDB.FirstOrCreate(&target, db.User{ID: targetID}).Error; err != nil {
				return
			}
			if target.Coins <= 0 {
				msg, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "StealNotEnoughCoinsTarget"})
				if msg == "" {
					msg = "Cette personne n'a pas de pièces à voler !"
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
				stolenCoins := (target.Coins * 10) / 100
				if stolenCoins <= 0 && target.Coins > 0 {
					stolenCoins = 1
				} else if stolenCoins < 0 {
					stolenCoins = 0
				}

				receivedCoins := stolenCoins
				if receivedCoins > 20 {
					receivedCoins -= 10
				}

				target.Coins -= stolenCoins
				attacker.Coins += receivedCoins

				b.DB.GormDB.Model(&target).Update("coins", target.Coins)
				b.DB.GormDB.Model(&attacker).Update("coins", attacker.Coins)

				resultMsg, _ = loc.Localize(&i18n.LocalizeConfig{
					MessageID:    "StealSuccess",
					TemplateData: map[string]interface{}{"Coins": stolenCoins, "TargetID": targetID},
				})
			} else {
				lostCoins := (attacker.Coins * 5) / 100
				if lostCoins <= 0 && attacker.Coins > 0 {
					lostCoins = 1
				} else if lostCoins < 0 {
					lostCoins = 0
				}

				if lostCoins > 30 {
					lostCoins = 30
				}

				receivedCoins := lostCoins
				if receivedCoins > 20 {
					receivedCoins -= 10
				}

				attacker.Coins -= lostCoins
				target.Coins += receivedCoins

				b.DB.GormDB.Model(&attacker).Update("coins", attacker.Coins)
				b.DB.GormDB.Model(&target).Update("coins", target.Coins)

				resultMsg, _ = loc.Localize(&i18n.LocalizeConfig{
					MessageID:    "StealFailure",
					TemplateData: map[string]interface{}{"Coins": lostCoins, "TargetID": targetID},
				})
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
