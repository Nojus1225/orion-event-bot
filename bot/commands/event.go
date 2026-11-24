package commands

import (
	"fmt"

	"orion-event-bot/bot/core"
	"orion-event-bot/bot/handlers"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

func getEventContainer(page string, loc *i18n.Localizer, avatarURL string) discord.ContainerComponent {
	title, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "EventTitle"})
	prompt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "EventMenuPrompt"})

	overviewLabel, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "EventCatOverview"})
	overviewDesc, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "EventCatOverviewDesc"})
	commandsLabel, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "EventCatCommands"})
	commandsDesc, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "EventCatCommandsDesc"})
	itemsLabel, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "EventCatItems"})
	itemsDesc, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "EventCatItemsDesc"})

	var contentID string
	switch page {
	case "commands":
		contentID = "EventCommandsContent"
	case "items":
		contentID = "EventItemsContent"
	default:
		page = "overview"
		contentID = "EventOverviewContent"
	}

	content, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: contentID})

	selectMenu := discord.NewStringSelectMenu(
		"event-page",
		prompt,
		discord.StringSelectMenuOption{
			Label:       overviewLabel,
			Value:       "overview",
			Description: overviewDesc,
			Default:     page == "overview",
		},
		discord.StringSelectMenuOption{
			Label:       commandsLabel,
			Value:       "commands",
			Description: commandsDesc,
			Default:     page == "commands",
		},
		discord.StringSelectMenuOption{
			Label:       itemsLabel,
			Value:       "items",
			Description: itemsDesc,
			Default:     page == "items",
		},
	)

	return discord.NewContainer(
		discord.NewSection(
			discord.NewTextDisplay(title),
			discord.NewTextDisplay(content),
		).WithAccessory(discord.NewThumbnail(avatarURL)),
		discord.NewActionRow(selectMenu),
	).WithAccentColor(0x5865F2)
}

func init() {
	handlers.RegisterCommand(handlers.Command{
		Name:        "event",
		Description: "Learn how the Winter Rush event works",
		Create: discord.SlashCommandCreate{
			Name:        "event",
			Description: "Learn how the Winter Rush event works",
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			loc := core.GetLocalizer(event.Locale().Code())
			container := getEventContainer("overview", loc, event.User().EffectiveAvatarURL())

			if err := event.CreateMessage(discord.NewMessageCreateV2(container)); err != nil {
				fmt.Println("Error sending event message:", err)
			}
		},
		ExecuteMenu: func(b *core.Bot, event *events.ComponentInteractionCreate) {
			loc := core.GetLocalizer(event.Locale().Code())
			notYourComponentTxt, _ := loc.Localize(&i18n.LocalizeConfig{MessageID: "NotYourComponent"})
			if (event.Message.InteractionMetadata != nil && event.Message.InteractionMetadata.User.ID != event.User().ID) ||
				(event.Message.Interaction != nil && event.Message.Interaction.User.ID != event.User().ID) {
				_ = event.CreateMessage(discord.NewMessageCreate().WithContent(notYourComponentTxt).WithEphemeral(true))
				return
			}

			if event.Data.CustomID() == "event-page" {
				page := event.StringSelectMenuInteractionData().Values[0]
				container := getEventContainer(page, loc, event.User().EffectiveAvatarURL())
				if err := event.UpdateMessage(discord.NewMessageUpdateV2(container)); err != nil {
					fmt.Println("Error updating event message:", err)
				}
			}
		},
	})
}
