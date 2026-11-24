package handlers

import (
	"strings"

	"orion-event-bot/bot/core"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type Command struct {
	Name                string
	Description         string
	Create              discord.ApplicationCommandCreate
	Execute             func(b *core.Bot, event *events.ApplicationCommandInteractionCreate)
	ExecuteButton       func(b *core.Bot, event *events.ComponentInteractionCreate)
	ExecuteMenu         func(b *core.Bot, event *events.ComponentInteractionCreate)
	ExecuteModal        func(b *core.Bot, event *events.ModalSubmitInteractionCreate)
	ExecuteAutocomplete func(b *core.Bot, event *events.AutocompleteInteractionCreate)
}

var Commands = make(map[string]Command)

func RegisterCommand(cmd Command) {
	Commands[cmd.Name] = cmd
}

func SyncCommands(b *core.Bot) error {
	frLoc := core.GetLocalizer("fr")
	enLoc := core.GetLocalizer("en-US")

	var cmds []discord.ApplicationCommandCreate
	for _, cmd := range Commands {
		create := cmd.Create
		if slash, ok := create.(discord.SlashCommandCreate); ok && len(cmd.Name) > 0 {
			prefix := strings.ToUpper(cmd.Name[:1]) + cmd.Name[1:]
			if desc, err := enLoc.Localize(&i18n.LocalizeConfig{MessageID: prefix + "Description"}); err == nil && desc != "" {
				slash.Description = desc
			}
			if descFr, err := frLoc.Localize(&i18n.LocalizeConfig{MessageID: prefix + "Description"}); err == nil && descFr != "" {
				if slash.DescriptionLocalizations == nil {
					slash.DescriptionLocalizations = make(map[discord.Locale]string)
				}
				slash.DescriptionLocalizations[discord.LocaleFrench] = descFr
			}
			if nameFr, err := frLoc.Localize(&i18n.LocalizeConfig{MessageID: prefix + "Name"}); err == nil && nameFr != "" {
				if slash.NameLocalizations == nil {
					slash.NameLocalizations = make(map[discord.Locale]string)
				}
				slash.NameLocalizations[discord.LocaleFrench] = nameFr
			}
			create = slash
		}
		cmds = append(cmds, create)
	}

	if b.GuildID != "" {
		guildID := snowflake.MustParse(b.GuildID)
		_, _ = b.Client.Rest.SetGlobalCommands(b.Client.ApplicationID, []discord.ApplicationCommandCreate{})

		_, err := b.Client.Rest.SetGuildCommands(b.Client.ApplicationID, guildID, cmds)
		return err
	}

	_, err := b.Client.Rest.SetGlobalCommands(b.Client.ApplicationID, cmds)
	return err
}
