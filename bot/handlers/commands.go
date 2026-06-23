package handlers

import (
	"orion-event-bot/bot/core"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

type Command struct {
	Name        string
	Description string
	Create      discord.ApplicationCommandCreate
	Execute     func(b *core.Bot, event *events.ApplicationCommandInteractionCreate)
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

	// Wait, if guildID is defined, we could sync to guild, or to global. 
	// For simplicity, let's sync to global. Main can choose what to do.
	_, err := b.Client.Rest.SetGlobalCommands(b.Client.ApplicationID, cmds)
	return err
}
