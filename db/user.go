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

type User struct {
	ID string `gorm:"primaryKey;type:varchar(255)"`

	Coins     int
	CDFouille int

	StatusMalus StatusMalus //"none", "sunburn", "cut"
	CDMalus     int

	Tier Tier // 0: aucun, 1: Sac, 2: Tamis, 3: Detecteur

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
