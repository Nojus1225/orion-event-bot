package main

import (
	"fmt"
	"log"

	"orion-event-bot/bot/vars"
	"orion-event-bot/db"
)

func main() {
	database := db.New()

	event := db.Event{
		ID:         1,
		IsActive:   true,
		LodgeTotal: 0,
	}
	if err := database.GormDB.Where("id = ?", event.ID).FirstOrCreate(&event).Error; err != nil {
		log.Fatalf("err %v", err)
	}

	for _, tile := range vars.Plate {
		pos := db.GamePosition{
			PositionId: tile.Number,
			Piege:      db.SpecialCaseType(tile.SpecialCase),
			Snowflakes: 0,
		}
		if err := database.GormDB.Where("posid = ?", tile.Number).FirstOrCreate(&pos).Error; err != nil {
			log.Fatalf("err tile %d: %v", tile.Number, err)
		}
	}
	fmt.Printf("%d init tiles\n", len(vars.Plate))
}
