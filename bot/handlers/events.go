package handlers

import (
	"reflect"

	"orion-event-bot/bot/core"

	"github.com/disgoorg/disgo/bot"
)

type Event struct {
	Name     string
	ExecOnce bool
	Execute  func(b *core.Bot, e bot.Event)
}

var Events []Event

func RegisterEvent(e Event) {
	Events = append(Events, e)
}

func SetupEvents(b *core.Bot) {
	// Keep track of events executed for ExecOnce
	executedEvents := make(map[string]bool)

	b.Client.EventManager.AddEventListeners(bot.NewListenerFunc(func(e bot.Event) {
		// Use reflection to get the struct name (e.g. "ApplicationCommandInteractionCreate")
		eventName := reflect.TypeOf(e).Elem().Name()

		for _, ev := range Events {
			if ev.Name == eventName {
				// Handle ExecOnce
				if ev.ExecOnce {
					if executedEvents[ev.Name] {
						continue
					}
					executedEvents[ev.Name] = true
				}

				// Execute the event
				go ev.Execute(b, e)
			}
		}
	}))
}
