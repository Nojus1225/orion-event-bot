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
	if core.Bundle == nil {
		core.InitI18n()
	}

	tags := core.Bundle.LanguageTags()
	localizers := make(map[string]*i18n.Localizer, len(tags))
	for _, tag := range tags {
		if tag.String() == "und" {
			continue
		}
		localizers[tag.String()] = core.GetLocalizer(tag.String())
	}

	var cmds []discord.ApplicationCommandCreate
	for _, cmd := range Commands {
		create := cmd.Create
		if slash, ok := create.(discord.SlashCommandCreate); ok && len(cmd.Name) > 0 {
			prefix := strings.ToUpper(cmd.Name[:1]) + cmd.Name[1:]
			for lang, loc := range localizers {
				if desc, err := loc.Localize(&i18n.LocalizeConfig{MessageID: prefix + "Description"}); err == nil && desc != "" {
					if lang == "en-US" {
						slash.Description = desc
					} else {
						if slash.DescriptionLocalizations == nil {
							slash.DescriptionLocalizations = make(map[discord.Locale]string)
						}
						slash.DescriptionLocalizations[discord.Locale(lang)] = desc
					}
				}
				if lang != "en-US" {
					if name, err := loc.Localize(&i18n.LocalizeConfig{MessageID: prefix + "Name"}); err == nil && name != "" {
						if slash.NameLocalizations == nil {
							slash.NameLocalizations = make(map[discord.Locale]string)
						}
						slash.NameLocalizations[discord.Locale(lang)] = name
					}
				}
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
