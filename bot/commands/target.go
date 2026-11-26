package commands

import (
	"fmt"
	"math/rand"

	"orion-event-bot/bot/core"
	"orion-event-bot/bot/handlers"
	"orion-event-bot/db"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

func sendTargetError(event *events.ApplicationCommandInteractionCreate, loc *i18n.Localizer, msgID string, data map[string]interface{}) {
	txt, _ := loc.Localize(&i18n.LocalizeConfig{
		MessageID:    msgID,
		TemplateData: data,
	})
	msg := discord.NewMessageCreateV2(
		discord.NewContainer(
			discord.NewSection(
				discord.NewTextDisplay(func() string { t, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "TargetTitle"}); return t }()),
				discord.NewTextDisplay(txt),
			).WithAccessory(discord.NewThumbnail(event.User().EffectiveAvatarURL())),
		).WithAccentColor(0xED4245),
	)
	_ = event.CreateMessage(msg)
}

func init() {
	handlers.RegisterCommand(handlers.Command{
		Name:        "target",
		Description: "Attack a user",
		Create: discord.SlashCommandCreate{
			Name:        "target",
			Description: "Attack a user",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionUser{
					Name:        "user",
					Description: "The user to attack",
					Required:    true,
				},
				discord.ApplicationCommandOptionString{
					Name:        "type",
					Description: "The type of the attack",
					Required:    true,
					Choices: []discord.ApplicationCommandOptionChoiceString{
						{
							Name:  "snowball",
							Value: "snowball",
						},
						{
							Name:  "pickpocket",
							Value: "pickpocket",
						},
					},
				},
			},
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			loc := core.GetLocalizer(event.Locale().Code())

			var user db.User
			if err := b.DB.GormDB.Where("id = ?", event.User().ID.String()).First(&user).Error; err != nil {
				fmt.Println("DB err:", err)
				return
			}

			if user.Energy <= 0 {
				sendTargetError(event, loc, "TargetNoEnergy", nil)
				return
			}

			targetId := event.SlashCommandInteractionData().User("user").ID.String()

			if targetId == event.User().ID.String() {
				sendTargetError(event, loc, "TargetSelf", nil)
				return
			}

			var target db.User
			if err := b.DB.GormDB.Where("id = ?", targetId).First(&target).Error; err != nil {
				sendTargetError(event, loc, "TargetNotParticipating", map[string]interface{}{"Target": targetId})
				return
			}

			if target.Position >= 50 {
				sendTargetError(event, loc, "TargetFinished", nil)
				return
			}

			var allPlayers []db.User
			if err := b.DB.GormDB.Order("position desc").Find(&allPlayers).Error; err != nil {
				fmt.Println("DB err:", err)
				return
			}

			userRank, targetRank := -1, -1
			for i, p := range allPlayers {
				if p.ID == user.ID {
					userRank = i
				}
				if p.ID == target.ID {
					targetRank = i
				}
			}
			if targetRank-userRank > 2 {
				sendTargetError(event, loc, "TargetTooFarBehind", nil)
				return
			}

			user.Energy -= 1

			if target.BuffHelmet {
				target.BuffHelmet = false
				user.Energy -= 1
				_ = b.DB.GormDB.Save(&user).Error
				_ = b.DB.GormDB.Save(&target).Error
				sendTargetError(event, loc, "TargetHelmet", map[string]interface{}{"Target": target.ID})
				return
			}

			attackType := event.SlashCommandInteractionData().String("type")
			if attackType == "pickpocket" && target.Snowflakes < 15 {
				sendTargetError(event, loc, "TargetNotEnoughSnowflakes", map[string]interface{}{"Target": target.ID})
				return
			}

			user.Energy -= 1

			var responseText string
			switch attackType {
			case "snowball":
				target.Position -= 2
				target.Position = max(target.Position, 1)
				responseText, _ = loc.Localize(&i18n.LocalizeConfig{
					MessageID:    "TargetSnowballSuccess",
					TemplateData: map[string]interface{}{"Target": target.ID},
				})
			case "pickpocket":
				if n := rand.Intn(10); n < 6 {
					stolen := min(15, target.Snowflakes)
					target.Snowflakes -= stolen
					user.Snowflakes += stolen
					responseText, _ = loc.Localize(&i18n.LocalizeConfig{
						MessageID:    "TargetPickpocketSuccess",
						TemplateData: map[string]interface{}{"Stolen": stolen, "Target": target.ID},
					})
				} else {
					responseText, _ = loc.Localize(&i18n.LocalizeConfig{
						MessageID:    "TargetPickpocketFail",
						TemplateData: map[string]interface{}{"Target": target.ID},
					})
				}
			}

			if err := b.DB.GormDB.Save(&user).Error; err != nil {
				fmt.Println("DB err:", err)
				return
			}
			if err := b.DB.GormDB.Save(&target).Error; err != nil {
				fmt.Println("DB err:", err)
				return
			}

			msg := discord.NewMessageCreateV2(
				discord.NewContainer(
					discord.NewSection(
						discord.NewTextDisplay(func() string { t, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "TargetTitle"}); return t }()),
						discord.NewTextDisplay(responseText),
					).WithAccessory(discord.NewThumbnail(event.User().EffectiveAvatarURL())),
				).WithAccentColor(0x5865F2),
			).WithAllowedMentions(&discord.AllowedMentions{
				Users: []snowflake.ID{snowflake.MustParse(target.ID)},
			})

			err := event.CreateMessage(msg)
			if err != nil {
				fmt.Println("Error sending target message:", err)
			}
		},
	})
}
