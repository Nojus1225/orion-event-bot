package utils

import (
	"math/rand"
	"orion-event-bot/bot/core"
	"orion-event-bot/bot/vars"
	"orion-event-bot/db"

	"github.com/disgoorg/disgo/discord"
	"gorm.io/gorm"
)

func UpdateUserWithSpecialTile(gormDB *gorm.DB, user db.User, tileType vars.SpecialCaseType, game db.GamePosition, event db.Event) (db.User, db.GamePosition, db.Event) {
	switch tileType {
	case vars.SpecialCaseTree:
		user.Position -= 2
		game.Snowflakes += 10
		user.Snowflakes -= 10
	case vars.SpecialCaseRavine:
		if user.BuffHelmet {
			user.Position -= 4
		} else {
			user.Position = 32
		}
	case vars.SpecialCaseYeti:
		toRemove := int(float64(user.Snowflakes) * 0.20)
		user.Snowflakes -= toRemove
	case vars.SpecialCaseSlush:
		break
	case vars.SpecialCaseJump:
		user.Position += rand.Intn(10) - 4
	case vars.SpecialCaseBlizzard:
		var closest db.User
		if err := gormDB.Where("id != ?", user.ID).Order(gorm.Expr("ABS(position - ?)", user.Position)).First(&closest).Error; err == nil {
			user.Position, closest.Position = closest.Position, user.Position
			gormDB.Save(&closest)
		}
	case vars.SpecialCaseChairlift:
		user.Snowflakes += 25
	case vars.SpecialCaseIce:
		user.Position += 3
	case vars.SpecialCaseGondola:
		user.Position = 14
	case vars.SpecialCaseSupply:
		if len(user.Inventory) >= 2 || rand.Intn(2) == 0 {
			user.Snowflakes += 25
		} else {
			rItem := vars.RollRandomItem()
			user.Inventory = append(user.Inventory, string(rItem.ID))
		}
	}
	return user, game, event
}

func GetMessageByTileType(tile vars.PlateCase, user discord.User, u db.User, locale string) discord.MessageUpdate {
	loc := core.GetLocalizer(locale)

	title := LocalizeText(loc, "RideTitle", nil, "# ⛷️ Ride")
	var key string
	switch tile.SpecialCase {
	case vars.SpecialCaseTree:
		key = "RideSpecialTree"
	case vars.SpecialCaseRavine:
		if u.BuffHelmet {
			key = "RideSpecialCaveWithHelmet"
		} else {
			key = "RideSpecialCaveNoHelmet"
		}
	case vars.SpecialCaseYeti:
		key = "RideSpecialYeti"
	case vars.SpecialCaseSlush:
		key = "RideSpecialSlush"
	case vars.SpecialCaseJump:
		key = "RideSpecialJump"
	case vars.SpecialCaseBlizzard:
		key = "RideSpecialBlizzard"
	case vars.SpecialCaseChairlift:
		key = "RideSpecialChairlift"
	case vars.SpecialCaseIce:
		key = "RideSpecialIce"
	case vars.SpecialCaseGondola:
		key = "RideSpecialGondola"
	case vars.SpecialCaseSupply:
		key = "RideSpecialSupply"
	default:
		key = "RideNormal"
	}

	body := LocalizeText(loc, key, map[string]interface{}{
		"newPos": u.Position,
	}, "")

	return discord.NewMessageUpdateV2(
		discord.NewContainer(
			discord.NewSection(
				discord.NewTextDisplay(title),
				discord.NewTextDisplay(body),
			).WithAccessory(discord.NewThumbnail(*user.AvatarURL())),
		),
	)
}
