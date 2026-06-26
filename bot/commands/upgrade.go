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
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

func getUpgradeContainer(user db.User, loc *i18n.Localizer) discord.ContainerComponent {
	if user.Tier == db.TierDetecteur {
		maxTierTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "UpgradeMaxTier"})
		return discord.NewContainer(
			discord.NewTextDisplay(maxTierTxt),
		).WithAccentColor(0x5865F2)
	}

	var nameKey, descKey string
	var cost int
	var nextTier db.Tier
	var thumbnail string

	switch user.Tier {
	case db.TierNone:
		nameKey = "UpgradeTier1Name"
		descKey = "UpgradeTier1Desc"
		cost = 60
		nextTier = db.TierSac
		thumbnail = "https://raw.githubusercontent.com/googlefonts/noto-emoji/main/png/512/emoji_u1f392.png"
	case db.TierSac:
		nameKey = "UpgradeTier2Name"
		descKey = "UpgradeTier2Desc"
		cost = 140
		nextTier = db.TierTamis
		thumbnail = "https://raw.githubusercontent.com/googlefonts/noto-emoji/main/png/512/emoji_u1faa3.png"
	case db.TierTamis:
		nameKey = "UpgradeTier3Name"
		descKey = "UpgradeTier3Desc"
		cost = 300
		nextTier = db.TierDetecteur
		thumbnail = "https://raw.githubusercontent.com/googlefonts/noto-emoji/main/png/512/emoji_u1f9f2.png"
	}

	locName, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: nameKey})
	locDesc, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: descKey})

	costTxt, _ := loc.Localize(&i18n.LocalizeConfig{
		MessageID:    "UpgradeCost",
		TemplateData: map[string]interface{}{"Coins": cost},
	})

	content := fmt.Sprintf("### %s\n-# %s\n%s", locName, locDesc, costTxt)

	canBuy := user.Coins >= cost

	buyItemTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "UpgradeBuy"})

	return discord.NewContainer(
		discord.NewSection(
			discord.NewTextDisplay(content),
		).WithAccessory(discord.NewThumbnail(thumbnail)),
		discord.NewActionRow(
			discord.NewButton(discord.ButtonStyleSuccess, buyItemTxt, fmt.Sprintf("upgrade-buy-%d", nextTier), "", snowflake.ID(0)).WithDisabled(!canBuy).WithEmoji(discord.NewComponentEmoji("⬆️")),
		),
	).WithAccentColor(0x5865F2)
}

func init() {
	handlers.RegisterCommand(handlers.Command{
		Name:        "upgrade",
		Description: "Upgrade your equipment",
		Create: discord.SlashCommandCreate{
			Name:        "upgrade",
			Description: "Upgrade your equipment",
			DescriptionLocalizations: map[discord.Locale]string{
				discord.LocaleFrench: "Améliorer votre équipement",
			},
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			loc := core.GetLocalizer(event.Locale().String())
			var data db.User
			if err := b.DB.GormDB.FirstOrCreate(&data, db.User{ID: strconv.FormatUint(uint64(event.User().ID), 10)}).Error; err != nil {
				fmt.Println("DB error:", err)
			}

			container := getUpgradeContainer(data, loc)

			msg := discord.NewMessageCreateV2(container)

			if err := event.CreateMessage(msg); err != nil {
				fmt.Println("Error sending upgrade message:", err)
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

			customID := event.Data.CustomID()
			if !strings.HasPrefix(customID, "upgrade-buy-") {
				return
			}

			parts := strings.SplitN(customID, "-", 3)
			if len(parts) != 3 {
				return
			}

			targetTierInt, err := strconv.Atoi(parts[2])
			if err != nil {
				return
			}
			targetTier := db.Tier(targetTierInt)

			var data db.User
			if err := b.DB.GormDB.FirstOrCreate(&data, db.User{ID: strconv.FormatUint(uint64(event.User().ID), 10)}).Error; err != nil {
				fmt.Println("DB error:", err)
			}

			var cost int
			if data.Tier == db.TierNone && targetTier == db.TierSac {
				cost = 60
			} else if data.Tier == db.TierSac && targetTier == db.TierTamis {
				cost = 140
			} else if data.Tier == db.TierTamis && targetTier == db.TierDetecteur {
				cost = 300
			} else {
				return
			}

			if data.Coins < cost {
				errMsg, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "UpgradeNotEnoughCoins"})
				msg := discord.NewMessageCreate().WithEphemeral(true).WithContent(errMsg)
				if err := event.CreateMessage(msg); err != nil {
					fmt.Println("Error sending upgrade error message:", err)
				}
				return
			}

			data.Coins -= cost
			data.StatsTotalCoinsDepense += cost
			data.Tier = targetTier

			if err := b.DB.GormDB.Model(&data).Select("Coins", "Tier", "StatsTotalCoinsDepense").Updates(data).Error; err != nil {
				fmt.Println("DB error:", err)
			}

			container := getUpgradeContainer(data, loc)
			msg := discord.NewMessageUpdateV2(container)

			if err := event.UpdateMessage(msg); err != nil {
				fmt.Println("Error updating upgrade message:", err)
			}

			successMsg, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "UpgradeBuySuccess"})
			followup := discord.NewMessageCreate().WithEphemeral(true).WithContent(successMsg)
			if _, err := event.Client().Rest.CreateFollowupMessage(event.ApplicationID(), event.Token(), followup); err != nil {
				fmt.Println("Error sending upgrade followup:", err)
			}
		},
	})
}
