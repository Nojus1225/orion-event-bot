package vars

import (
	"fmt"
	"math/rand/v2"
	"sort"
	"strings"

	"orion-event-bot/db"

	"gorm.io/gorm"
)

type ItemID string

const (
	ItemHelmet    ItemID = "helmet"
	ItemMagnet    ItemID = "magnet"
	ItemWax       ItemID = "wax"
	ItemAvalanche ItemID = "avalanche"
)

type ItemContext struct {
	DB         *gorm.DB
	User       *db.User
	AllPlayers []*db.User
}

type ItemDef struct {
	ID          ItemID
	Name        string
	Emoji       string
	Description string
	Price       int
	DropWeight  int
	Use         func(ctx *ItemContext) (string, error)
}

var Items = map[ItemID]ItemDef{
	ItemHelmet: {
		ID:          ItemHelmet,
		Name:        "Full-Face Helmet",
		Emoji:       "🛡️",
		Description: "Blocks the next negative trap, snowball, or pickpocket.",
		Price:       30,
		DropWeight:  35,
		Use: func(ctx *ItemContext) (string, error) {
			if ctx.User.BuffHelmet {
				return "", fmt.Errorf("you already have a helmet equipped")
			}
			ctx.User.BuffHelmet = true
			return "🛡️ You equipped the **Full-Face Helmet**! Your next hit is negated.", nil
		},
	},
	ItemMagnet: {
		ID:          ItemMagnet,
		Name:        "Snowflake Magnet",
		Emoji:       "🧲",
		Description: "Steals 10 Snowflakes from up to 2 players closest to your tile.",
		Price:       35,
		DropWeight:  25,
		Use: func(ctx *ItemContext) (string, error) {
			stolenTotal, impacted := stealFromClosest(ctx.DB, ctx.User, ctx.AllPlayers, 10, 2)
			ctx.User.Snowflakes += stolenTotal
			if len(impacted) > 0 {
				return fmt.Sprintf("🧲 Your magnet pulled **%d Snowflakes** from %s!", stolenTotal, strings.Join(impacted, ", ")), nil
			}
			return fmt.Sprintf("🧲 Your magnet pulled **%d Snowflakes** from nearby skiers!", stolenTotal), nil
		},
	},
	ItemWax: {
		ID:          ItemWax,
		Name:        "Turbo Ski Wax",
		Emoji:       "🎿",
		Description: "Doubles the distance of your next /ride action.",
		Price:       40,
		DropWeight:  25,
		Use: func(ctx *ItemContext) (string, error) {
			if ctx.User.BuffWax {
				return "", fmt.Errorf("you already applied Turbo Ski Wax")
			}
			ctx.User.BuffWax = true
			return "🎿 You waxed your skis! Your next `/ride` distance will be **doubled**.", nil
		},
	},
	ItemAvalanche: {
		ID:          ItemAvalanche,
		Name:        "Avalanche Firecracker",
		Emoji:       "🧨",
		Description: "Pushes back all unfinished players ahead of you by 3 tiles.",
		Price:       55,
		DropWeight:  15,
		Use: func(ctx *ItemContext) (string, error) {
			hitCount, impacted := triggerAvalanche(ctx.DB, ctx.User, ctx.AllPlayers, 3)
			if len(impacted) > 0 {
				return fmt.Sprintf("🧨 BOOM! An avalanche pushed back **%d skier(s)** (%s) ahead of you by 3 tiles!", hitCount, strings.Join(impacted, ", ")), nil
			}
			return fmt.Sprintf("🧨 BOOM! An avalanche pushed back **%d skier(s)** ahead of you by 3 tiles!", hitCount), nil
		},
	},
}

func RollRandomItem() ItemDef {
	totalWeight := 0
	for _, item := range Items {
		totalWeight += item.DropWeight
	}

	roll := rand.IntN(totalWeight)
	current := 0
	for _, item := range Items {
		current += item.DropWeight
		if roll < current {
			return item
		}
	}
	return Items[ItemHelmet]
}

func triggerAvalanche(gormDB *gorm.DB, user *db.User, allPlayers []*db.User, pushBack int) (int, []string) {
	hitCount := 0
	var impacted []string
	for _, p := range allPlayers {
		if p.ID == user.ID || p.Position <= user.Position || p.Position >= 50 {
			continue
		}
		if p.BuffHelmet {
			p.BuffHelmet = false
			if gormDB != nil {
				gormDB.Save(p)
			}
			continue
		}
		p.Position -= pushBack
		if p.Position < 0 {
			p.Position = 0
		}
		if gormDB != nil {
			gormDB.Save(p)
		}
		impacted = append(impacted, fmt.Sprintf("<@%s>", p.ID))
		hitCount++
	}
	return hitCount, impacted
}

func stealFromClosest(gormDB *gorm.DB, user *db.User, allPlayers []*db.User, amount int, maxTargets int) (int, []string) {
	var candidates []*db.User
	for _, p := range allPlayers {
		if p.ID != user.ID {
			candidates = append(candidates, p)
		}
	}

	abs := func(x int) int {
		if x < 0 {
			return -x
		}
		return x
	}

	sort.Slice(candidates, func(i, j int) bool {
		return abs(candidates[i].Position-user.Position) < abs(candidates[j].Position-user.Position)
	})

	totalStolen := 0
	targets := 0
	var impacted []string
	for _, p := range candidates {
		if targets >= maxTargets {
			break
		}
		targets++
		if p.BuffHelmet {
			p.BuffHelmet = false
			if gormDB != nil {
				gormDB.Save(p)
			}
			continue
		}
		stolen := amount
		if p.Snowflakes < stolen {
			stolen = p.Snowflakes
		}
		p.Snowflakes -= stolen
		if gormDB != nil {
			gormDB.Save(p)
		}
		if stolen > 0 {
			impacted = append(impacted, fmt.Sprintf("<@%s>", p.ID))
		}
		totalStolen += stolen
	}
	return totalStolen, impacted
}
