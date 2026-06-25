package db

type StatusMalus string

const (
	MalusNone    StatusMalus = "none"
	MalusSunBurn StatusMalus = "sunburn"
	MalusCut     StatusMalus = "cut"
)

type Tier int

const (
	TierNone      Tier = 0
	TierSac       Tier = 1
	TierTamis     Tier = 2
	TierDetecteur Tier = 3
)

type Item int

const (
	ShopCremeSolaire Item = 0
	ShopPansement    Item = 1
	ShopBouteilleEau Item = 2
)

type User struct {
	ID string `gorm:"primaryKey;type:varchar(255)"`

	Coins     int
	CDFouille int

	StatusMalus StatusMalus
	CDMalus     int

	Inventory []int `gorm:"serializer:json"`

	Tier Tier

	ItemCoquillage int
	ItemBouteille  int
	ItemCannette   int

	ItemLunettes int
	ItemPiece    int
	ItemPince    int

	ItemMsgBouteille int
	ItemMontre       int
	ItemDent         int

	StatsFouille           int
	StatsObjetsTrouve      int
	StatsTotalCoins        int
	StatsTotalCoinsDepense int

	StatsTotalRien    int
	StatsTotalSunburn int
	StatsTotalCuts    int
}
