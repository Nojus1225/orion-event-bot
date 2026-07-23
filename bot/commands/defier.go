package commands

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
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

type RPCGame struct {
	SessionID        string
	ChallengerID     string
	TargetID         string
	Wager            int
	Status           string
	ChallengerChoice string
	TargetChoice     string
	MessageID        string
	ChannelID        string
	Mu               sync.Mutex
	AcceptTimer      *time.Timer
	GameTimer        *time.Timer
	WarningTimer     *time.Timer
}

var activeRPCGames sync.Map

func init() {
	handlers.RegisterCommand(handlers.Command{
		Name:        "challenge",
		Description: "Challenge a player",
		Create: discord.SlashCommandCreate{
			Name:        "challenge",
			Description: "Challenge a player",
			DescriptionLocalizations: map[discord.Locale]string{
				discord.LocaleFrench: "Défier un joueur",
			},
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionSubCommand{
					Name:        "rpc",
					Description: "Challenge a player to Rock Paper Scissors",
					DescriptionLocalizations: map[discord.Locale]string{
						discord.LocaleFrench: "Défier un joueur au Pierre Papier Ciseaux",
					},
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionUser{
							Name:        "user",
							Description: "The targeted user",
							Required:    true,
						},
						discord.ApplicationCommandOptionInt{
							Name:        "mise",
							Description: "The wager in coins",
							Required:    true,
						},
					},
				},
			},
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			loc := core.GetLocalizer(event.Locale().String())

			if event.SlashCommandInteractionData().SubCommandName == nil || *event.SlashCommandInteractionData().SubCommandName != "rpc" {
				return
			}

			attackerID := strconv.FormatUint(uint64(event.User().ID), 10)
			targetUser := event.SlashCommandInteractionData().User("user")
			targetID := targetUser.ID.String()
			wager := event.SlashCommandInteractionData().Int("mise")

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
			if attacker.Coins < wager {
				msg, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "DefierNotEnoughCoinsAttacker"})
				if msg == "" {
					msg = "Vous n'avez pas assez de pièces pour cette mise."
				}
				_ = event.CreateMessage(discord.NewMessageCreate().WithContent(msg).WithEphemeral(true))
				return
			}

			var target db.User
			if err := b.DB.GormDB.FirstOrCreate(&target, db.User{ID: targetID}).Error; err != nil {
				return
			}
			if target.Coins < wager {
				msg, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "DefierNotEnoughCoinsTarget"})
				if msg == "" {
					msg = "L'utilisateur ciblé n'a pas assez de pièces pour cette mise."
				}
				_ = event.CreateMessage(discord.NewMessageCreate().WithContent(msg).WithEphemeral(true))
				return
			}

			bBytes := make([]byte, 16)
			_, _ = rand.Read(bBytes)
			sessionID := hex.EncodeToString(bBytes)

			game := &RPCGame{
				SessionID:    sessionID,
				ChallengerID: attackerID,
				TargetID:     targetID,
				Wager:        wager,
				Status:       "PENDING",
				ChannelID:    event.Channel().ID().String(),
			}

			activeRPCGames.Store(sessionID, game)

			btnAcceptTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "DefierAccept"})
			if btnAcceptTxt == "" {
				btnAcceptTxt = "Accepter"
			}
			btnDeclineTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "DefierDecline"})
			if btnDeclineTxt == "" {
				btnDeclineTxt = "Refuser"
			}

			btnAccept := discord.NewButton(discord.ButtonStyleSuccess, btnAcceptTxt, fmt.Sprintf("challenge-%s-accept", sessionID), "", snowflake.ID(0))
			btnDecline := discord.NewButton(discord.ButtonStyleDanger, btnDeclineTxt, fmt.Sprintf("challenge-%s-decline", sessionID), "", snowflake.ID(0))

			msgStr, _ := loc.Localize(&i18n.LocalizeConfig{
				MessageID: "DefierProposal",
				TemplateData: map[string]interface{}{
					"ChallengerID": attackerID,
					"TargetID":     targetID,
					"Wager":        wager,
				},
			})
			if msgStr == "" {
				msgStr = fmt.Sprintf("<@%s> vous défie au Pierre-Papier-Ciseaux pour **%d pièces** ! <@%s>, acceptez-vous le défi ?", attackerID, wager, targetID)
			}

			defierTitle, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "DefierTitle"})
			if defierTitle == "" {
				defierTitle = "# ⚔️ Défi Pierre-Papier-Ciseaux"
			}

			msgCreate := discord.NewMessageCreateV2(
				discord.NewContainer(
					discord.NewSection(
						discord.NewTextDisplay(defierTitle),
						discord.NewTextDisplay(msgStr),
					).WithAccessory(discord.NewThumbnail("https://raw.githubusercontent.com/googlefonts/noto-emoji/main/png/512/emoji_u270a.png")),
					discord.NewActionRow(btnAccept, btnDecline),
				).WithAccentColor(0x5865F2),
			).WithAllowedMentions(&discord.AllowedMentions{
				Users: []snowflake.ID{snowflake.MustParse(attackerID), snowflake.MustParse(targetID)},
			})

			if err := event.CreateMessage(msgCreate); err != nil {
				fmt.Println("Error sending rpc proposal message:", err)
				return
			}

			appID := event.ApplicationID()
			token := event.Token()

			game.AcceptTimer = time.AfterFunc(1*time.Minute, func() {
				game.Mu.Lock()
				defer game.Mu.Unlock()
				if game.Status != "PENDING" {
					return
				}
				activeRPCGames.Delete(sessionID)

				expireMsg, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "DefierExpiredProposal"})
				if expireMsg == "" {
					expireMsg = "⏳ Le défi a expiré."
				}

				update := discord.NewMessageUpdateV2(
					discord.NewContainer(
						discord.NewSection(
							discord.NewTextDisplay(defierTitle),
							discord.NewTextDisplay(expireMsg),
						).WithAccessory(discord.NewThumbnail("https://raw.githubusercontent.com/googlefonts/noto-emoji/main/png/512/emoji_u23f3.png")),
					).WithAccentColor(0xE74C3C),
				)
				_, _ = b.Client.Rest.UpdateInteractionResponse(appID, token, update)
			})
		},
		ExecuteButton: func(b *core.Bot, event *events.ComponentInteractionCreate) {
			loc := core.GetLocalizer(event.Locale().String())
			customID := event.ButtonInteractionData().CustomID()
			if !strings.HasPrefix(customID, "challenge-") {
				return
			}

			parts := strings.Split(customID, "-")
			if len(parts) < 3 {
				return
			}
			sessionID := parts[1]
			action := parts[2]

			val, ok := activeRPCGames.Load(sessionID)
			if !ok {
				_ = event.UpdateMessage(discord.NewMessageUpdateV2(
					discord.NewContainer(
						discord.NewTextDisplay("⏳ Action expirée."),
					).WithAccentColor(0xE74C3C),
				).WithComponents())
				return
			}

			game := val.(*RPCGame)
			game.Mu.Lock()
			defer game.Mu.Unlock()

			userID := strconv.FormatUint(uint64(event.User().ID), 10)
			appID := event.ApplicationID()
			token := event.Token()

			defierTitle, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "DefierTitle"})
			if defierTitle == "" {
				defierTitle = "# ⚔️ Défi Pierre-Papier-Ciseaux"
			}

			if game.Status == "PENDING" {
				if userID != game.TargetID {
					msg, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "DefierNotYourChallenge"})
					if msg == "" {
						msg = "Ce n'est pas votre défi !"
					}
					_ = event.CreateMessage(discord.NewMessageCreate().WithContent(msg).WithEphemeral(true))
					return
				}

				if game.AcceptTimer != nil {
					game.AcceptTimer.Stop()
				}

				if action == "decline" {
					activeRPCGames.Delete(sessionID)
					declineMsg, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "DefierDeclined"})
					if declineMsg == "" {
						declineMsg = "Le défi a été refusé."
					}

					_ = event.UpdateMessage(discord.NewMessageUpdateV2(
						discord.NewContainer(
							discord.NewSection(
								discord.NewTextDisplay(defierTitle),
								discord.NewTextDisplay(declineMsg),
							).WithAccessory(discord.NewThumbnail("https://raw.githubusercontent.com/googlefonts/noto-emoji/main/png/512/emoji_u274c.png")),
						).WithAccentColor(0xE74C3C),
					))
					return
				} else if action == "accept" {
					var attacker db.User
					b.DB.GormDB.First(&attacker, "id = ?", game.ChallengerID)
					var target db.User
					b.DB.GormDB.First(&target, "id = ?", game.TargetID)

					if attacker.Coins < game.Wager || target.Coins < game.Wager {
						activeRPCGames.Delete(sessionID)
						errMsg, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "DefierNotEnoughCoinsAny"})
						if errMsg == "" {
							errMsg = "L'un des joueurs n'a plus assez de pièces. Défi annulé."
						}
						_ = event.UpdateMessage(discord.NewMessageUpdateV2(
							discord.NewContainer(
								discord.NewSection(
									discord.NewTextDisplay(defierTitle),
									discord.NewTextDisplay(errMsg),
								).WithAccessory(discord.NewThumbnail("https://raw.githubusercontent.com/googlefonts/noto-emoji/main/png/512/emoji_u274c.png")),
							).WithAccentColor(0xE74C3C),
						))
						return
					}

					attacker.Coins -= game.Wager
					target.Coins -= game.Wager
					b.DB.GormDB.Model(&attacker).Update("coins", attacker.Coins)
					b.DB.GormDB.Model(&target).Update("coins", target.Coins)

					game.Status = "PLAYING"

					btnRockTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "DefierRock"})
					if btnRockTxt == "" {
						btnRockTxt = "Pierre"
					}
					btnPaperTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "DefierPaper"})
					if btnPaperTxt == "" {
						btnPaperTxt = "Papier"
					}
					btnScissorsTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "DefierScissors"})
					if btnScissorsTxt == "" {
						btnScissorsTxt = "Ciseaux"
					}

					btnRock := discord.NewButton(discord.ButtonStylePrimary, btnRockTxt, fmt.Sprintf("challenge-%s-play-rock", sessionID), "", snowflake.ID(0)).WithEmoji(discord.NewComponentEmoji("🪨"))
					btnPaper := discord.NewButton(discord.ButtonStylePrimary, btnPaperTxt, fmt.Sprintf("challenge-%s-play-paper", sessionID), "", snowflake.ID(0)).WithEmoji(discord.NewComponentEmoji("📄"))
					btnScissors := discord.NewButton(discord.ButtonStylePrimary, btnScissorsTxt, fmt.Sprintf("challenge-%s-play-scissors", sessionID), "", snowflake.ID(0)).WithEmoji(discord.NewComponentEmoji("✂️"))

					startMsg, _ := loc.Localize(&i18n.LocalizeConfig{
						MessageID:    "DefierGameStarted",
						TemplateData: map[string]interface{}{"ChallengerID": game.ChallengerID, "TargetID": game.TargetID},
					})
					if startMsg == "" {
						startMsg = "Le défi est accepté ! Choisissez votre coup."
					}

					update := discord.NewMessageUpdateV2(
						discord.NewContainer(
							discord.NewSection(
								discord.NewTextDisplay(defierTitle),
								discord.NewTextDisplay(startMsg),
							).WithAccessory(discord.NewThumbnail("https://raw.githubusercontent.com/googlefonts/noto-emoji/main/png/512/emoji_u270a.png")),
							discord.NewActionRow(btnRock, btnPaper, btnScissors),
						).WithAccentColor(0x5865F2),
					).WithAllowedMentions(&discord.AllowedMentions{
						Users: []snowflake.ID{snowflake.MustParse(game.ChallengerID), snowflake.MustParse(game.TargetID)},
					})
					_ = event.UpdateMessage(update)

					game.WarningTimer = time.AfterFunc(30*time.Second, func() {
						game.Mu.Lock()
						defer game.Mu.Unlock()
						if game.Status != "PLAYING" {
							return
						}

						var missing []string
						if game.ChallengerChoice == "" {
							missing = append(missing, fmt.Sprintf("<@%s>", game.ChallengerID))
						}
						if game.TargetChoice == "" {
							missing = append(missing, fmt.Sprintf("<@%s>", game.TargetID))
						}
						if len(missing) > 0 {
							warnMsg, _ := loc.Localize(&i18n.LocalizeConfig{
								MessageID:    "DefierWarning",
								TemplateData: map[string]interface{}{"Missing": strings.Join(missing, " et ")},
							})
							if warnMsg == "" {
								warnMsg = fmt.Sprintf("⏳ %s, il vous reste 30 secondes pour jouer, sinon vous serez déclaré forfait !", strings.Join(missing, " et "))
							}
							warnIDs := []snowflake.ID{}
							if game.ChallengerChoice == "" {
								warnIDs = append(warnIDs, snowflake.MustParse(game.ChallengerID))
							}
							if game.TargetChoice == "" {
								warnIDs = append(warnIDs, snowflake.MustParse(game.TargetID))
							}
							_, _ = b.Client.Rest.CreateMessage(snowflake.MustParse(game.ChannelID), discord.NewMessageCreate().WithContent(warnMsg).WithAllowedMentions(&discord.AllowedMentions{Users: warnIDs}))
						}
					})

					game.GameTimer = time.AfterFunc(1*time.Minute, func() {
						game.Mu.Lock()
						defer game.Mu.Unlock()
						if game.Status != "PLAYING" {
							return
						}

						activeRPCGames.Delete(sessionID)

						if game.ChallengerChoice == "" && game.TargetChoice == "" {
							b.DB.GormDB.Model(&db.User{}).Where("id = ?", game.ChallengerID).Update("coins", attacker.Coins+game.Wager)
							b.DB.GormDB.Model(&db.User{}).Where("id = ?", game.TargetID).Update("coins", target.Coins+game.Wager)
							cancelMsg, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "DefierGameCancelled"})
							if cancelMsg == "" {
								cancelMsg = "La partie a été annulée car aucun joueur n'a joué. Les mises ont été remboursées."
							}
							upd := discord.NewMessageUpdateV2(
								discord.NewContainer(
									discord.NewSection(discord.NewTextDisplay(defierTitle), discord.NewTextDisplay(cancelMsg)).WithAccessory(discord.NewThumbnail("https://raw.githubusercontent.com/googlefonts/noto-emoji/main/png/512/emoji_u274c.png")),
								).WithAccentColor(0xE74C3C),
							)
							_, _ = b.Client.Rest.UpdateInteractionResponse(appID, token, upd)
						} else {
							var winner string
							var loser string
							if game.ChallengerChoice != "" {
								winner = game.ChallengerID
								loser = game.TargetID
							} else {
								winner = game.TargetID
								loser = game.ChallengerID
							}

							var wUser db.User
							b.DB.GormDB.First(&wUser, "id = ?", winner)
							b.DB.GormDB.Model(&wUser).Update("coins", wUser.Coins+(game.Wager*2))

							forfeitMsg, _ := loc.Localize(&i18n.LocalizeConfig{
								MessageID:    "DefierForfeit",
								TemplateData: map[string]interface{}{"Loser": loser, "Winner": winner, "Wager": game.Wager},
							})
							if forfeitMsg == "" {
								forfeitMsg = fmt.Sprintf("⏱️ **Forfait !** <@%s> n'a pas joué à temps. <@%s> remporte la mise de **%d pièces** !", loser, winner, game.Wager)
							}
							upd := discord.NewMessageUpdateV2(
								discord.NewContainer(
									discord.NewSection(discord.NewTextDisplay(defierTitle), discord.NewTextDisplay(forfeitMsg)).WithAccessory(discord.NewThumbnail("https://raw.githubusercontent.com/googlefonts/noto-emoji/main/png/512/emoji_u23f3.png")),
								).WithAccentColor(0xE74C3C),
							).WithAllowedMentions(&discord.AllowedMentions{
								Users: []snowflake.ID{snowflake.MustParse(winner), snowflake.MustParse(loser)},
							})
							_, _ = b.Client.Rest.UpdateInteractionResponse(appID, token, upd)
						}
					})
				}
			} else if game.Status == "PLAYING" && action == "play" {
				if userID != game.ChallengerID && userID != game.TargetID {
					msg, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "DefierNotYourGame"})
					if msg == "" {
						msg = "Ce n'est pas votre partie !"
					}
					_ = event.CreateMessage(discord.NewMessageCreate().WithContent(msg).WithEphemeral(true))
					return
				}

				if len(parts) < 4 {
					return
				}
				choice := parts[3]

				if userID == game.ChallengerID {
					if game.ChallengerChoice != "" {
						msg, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "DefierAlreadyPlayed"})
						if msg == "" {
							msg = "Vous avez déjà joué !"
						}
						_ = event.CreateMessage(discord.NewMessageCreate().WithContent(msg).WithEphemeral(true))
						return
					}
					game.ChallengerChoice = choice
				} else {
					if game.TargetChoice != "" {
						msg, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "DefierAlreadyPlayed"})
						if msg == "" {
							msg = "Vous avez déjà joué !"
						}
						_ = event.CreateMessage(discord.NewMessageCreate().WithContent(msg).WithEphemeral(true))
						return
					}
					game.TargetChoice = choice
				}

				if game.ChallengerChoice != "" && game.TargetChoice != "" {
					activeRPCGames.Delete(sessionID)
					if game.WarningTimer != nil {
						game.WarningTimer.Stop()
					}
					if game.GameTimer != nil {
						game.GameTimer.Stop()
					}

					p1 := game.ChallengerChoice
					p2 := game.TargetChoice
					var winner string

					if p1 == p2 {
						winner = "draw"
					} else if (p1 == "rock" && p2 == "scissors") || (p1 == "paper" && p2 == "rock") || (p1 == "scissors" && p2 == "paper") {
						winner = game.ChallengerID
					} else {
						winner = game.TargetID
					}

					var resultMsg string
					var color int

					if winner == "draw" {
						var c db.User
						b.DB.GormDB.First(&c, "id = ?", game.ChallengerID)
						b.DB.GormDB.Model(&c).Update("coins", c.Coins+game.Wager)

						var t db.User
						b.DB.GormDB.First(&t, "id = ?", game.TargetID)
						b.DB.GormDB.Model(&t).Update("coins", t.Coins+game.Wager)

						p1Loc, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "DefierChoice_" + p1})
						if p1Loc == "" {
							p1Loc = p1
						}

						resultMsg, _ = loc.Localize(&i18n.LocalizeConfig{
							MessageID:    "DefierResultDraw",
							TemplateData: map[string]interface{}{"Choice": p1Loc},
						})
						if resultMsg == "" {
							resultMsg = fmt.Sprintf("🤝 **Égalité !** Vous avez tous les deux choisi **%s**. Les mises sont remboursées.", p1Loc)
						}
						color = 0xF1C40F
					} else {
						var wUser db.User
						b.DB.GormDB.First(&wUser, "id = ?", winner)
						b.DB.GormDB.Model(&wUser).Update("coins", wUser.Coins+(game.Wager*2))

						var loser string
						var winnerChoice string
						var loserChoice string
						if winner == game.ChallengerID {
							loser = game.TargetID
							winnerChoice = p1
							loserChoice = p2
						} else {
							loser = game.ChallengerID
							winnerChoice = p2
							loserChoice = p1
						}

						winnerChoiceLoc, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "DefierChoice_" + winnerChoice})
						if winnerChoiceLoc == "" {
							winnerChoiceLoc = winnerChoice
						}
						loserChoiceLoc, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "DefierChoice_" + loserChoice})
						if loserChoiceLoc == "" {
							loserChoiceLoc = loserChoice
						}

						resultMsg, _ = loc.Localize(&i18n.LocalizeConfig{
							MessageID:    "DefierResultWin",
							TemplateData: map[string]interface{}{"Winner": winner, "WinnerChoice": winnerChoiceLoc, "Loser": loser, "LoserChoice": loserChoiceLoc, "Wager": game.Wager},
						})
						if resultMsg == "" {
							resultMsg = fmt.Sprintf("🎉 **Victoire !** <@%s> (%s) bat <@%s> (%s) et remporte la mise de **%d pièces** !", winner, winnerChoiceLoc, loser, loserChoiceLoc, game.Wager)
						}
						color = 0x5865F2
					}

					upd := discord.NewMessageUpdateV2(
						discord.NewContainer(
							discord.NewSection(
								discord.NewTextDisplay(defierTitle),
								discord.NewTextDisplay(resultMsg),
							).WithAccessory(discord.NewThumbnail("https://raw.githubusercontent.com/googlefonts/noto-emoji/main/png/512/emoji_u270a.png")),
						).WithAccentColor(color),
					).WithAllowedMentions(&discord.AllowedMentions{
						Users: []snowflake.ID{snowflake.MustParse(game.ChallengerID), snowflake.MustParse(game.TargetID)},
					})
					_ = event.UpdateMessage(upd)

				} else {
					waitMsg, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "DefierWaiting"})
					if waitMsg == "" {
						waitMsg = "Choix enregistré, en attente de l'adversaire..."
					}
					_ = event.CreateMessage(discord.NewMessageCreate().WithContent(waitMsg).WithEphemeral(true))
				}
			}
		},
	})
}
