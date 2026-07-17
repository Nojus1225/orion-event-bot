package vars

import "orion-event-bot/db"

type ShopItem struct {
	Name        string
	Description string
	Cost        int
	ID          db.Item
	Thumbnail   string
}

type ShopCategory struct {
	Name        string
	Description string
	Items       []ShopItem
}

var ProtectionItems = []ShopItem{
	{
		Name:        "ShopSunscreenName",
		Description: "ShopSunscreenDesc",
		Cost:        2,
		ID:          db.ShopCremeSolaire,
		Thumbnail:   "https://raw.githubusercontent.com/googlefonts/noto-emoji/main/png/512/emoji_u1f9f4.png",
	},
	{
		Name:        "ShopBandageName",
		Description: "ShopBandageDesc",
		Cost:        4,
		ID:          db.ShopPansement,
		Thumbnail:   "https://raw.githubusercontent.com/googlefonts/noto-emoji/main/png/512/emoji_u1fa79.png",
	},
	{
		Name:        "ShopWaterBottleName",
		Description: "ShopWaterBottleDesc",
		Cost:        5,
		ID:          db.ShopBouteilleEau,
		Thumbnail:   "https://raw.githubusercontent.com/googlefonts/noto-emoji/main/png/512/emoji_u1f4a7.png",
	},
	{
		Name:        "ShopCremeSolaire1000Name",
		Description: "ShopCremeSolaire1000Desc",
		Cost:        5,
		ID:          db.ShopCremeSolaire1000,
		Thumbnail:   "https://raw.githubusercontent.com/googlefonts/noto-emoji/main/png/512/emoji_u2b50.png",
	},
	{
		Name:        "ShopDrapeauBaignadeName",
		Description: "ShopDrapeauBaignadeDesc",
		Cost:        10,
		ID:          db.ShopDrapeauBaignade,
		Thumbnail:   "https://raw.githubusercontent.com/googlefonts/noto-emoji/main/png/512/emoji_u1f6a9.png",
	},
}

var SabotageItems = []ShopItem{
	{
		Name:        "ShopVolDePelleName",
		Description: "ShopVolDePelleDesc",
		Cost:        8,
		ID:          db.ShopVolDePelle,
		Thumbnail:   "https://raw.githubusercontent.com/googlefonts/noto-emoji/main/png/512/emoji_u26cf.png",
	}, {
		Name:        "ShopCrabeDePocheName",
		Description: "ShopCrabeDePocheDesc",
		Cost:        15,
		ID:          db.ShopCrabeDePoche,
		Thumbnail:   "https://raw.githubusercontent.com/googlefonts/noto-emoji/main/png/512/emoji_u1f980.png",
	}, {
		Name:        "ShopAlerteMedusesName",
		Description: "ShopAlerteMedusesDesc",
		Cost:        45,
		ID:          db.ShopAlerteMeduses,
		Thumbnail:   "https://raw.githubusercontent.com/googlefonts/noto-emoji/main/png/512/emoji_u1f419.png",
	},
}

var Shop = []ShopCategory{
	{
		Name:        "ShopProtectionName",
		Description: "ShopProtectionDesc",
		Items:       ProtectionItems,
	},
	{
		Name:        "ShopSabotageName",
		Description: "ShopSabotageDesc",
		Items:       SabotageItems,
	},
}

func GetItem(id db.Item) *ShopItem {
	for _, cat := range Shop {
		for i := range cat.Items {
			if cat.Items[i].ID == id {
				return &cat.Items[i]
			}
		}
	}
	return nil
}
