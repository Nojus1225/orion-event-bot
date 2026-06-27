package handlers

import (
	"orion-event-bot/bot/core"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
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
	var cmds []discord.ApplicationCommandCreate
	for _, cmd := range Commands {
		cmds = append(cmds, cmd.Create)
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
