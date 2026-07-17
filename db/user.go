package db

type StatusMalus string

const (
	MalusNone        StatusMalus = "none"
	MalusSunBurn     StatusMalus = "sunburn"
	MalusCut         StatusMalus = "cut"
	MalusShovelTheft StatusMalus = "shovel_theft"
)

type Tier int

const (
	TierNone      Tier = 0
	TierSac       Tier = 1
	TierTamis     Tier = 2
	TierDetecteur Tier = 3
)

type Item string

const (
	ShopCremeSolaire Item = "creme_solaire"
	ShopPansement    Item = "pansement"
	ShopBouteilleEau Item = "bouteille_eau"

	ShopCremeSolaire1000 Item = "creme_solaire_1000"
	ShopDrapeauBaignade  Item = "drapeau_baignade"
	ShopVolDePelle       Item = "vol_de_pelle"
	ShopCrabeDePoche     Item = "crabe_de_poche"
	ShopAlerteMeduses    Item = "alerte_meduses"

	ItemCoquillage Item = "coquillage"
	ItemBouteille  Item = "bouteille_plastique"
	ItemCannette   Item = "cannette"

	ItemLunettes Item = "lunettes"
	ItemPiece    Item = "piece"
	ItemPince    Item = "pince"

	ItemMsgBouteille Item = "msg_bouteille"
	ItemMontre       Item = "montre"
	ItemDent         Item = "dent"
)

type StatName string

const (
	StatFouille           StatName = "fouille"
	StatObjetsTrouve      StatName = "objets_trouve"
	StatTotalCoins        StatName = "total_coins"
	StatTotalCoinsDepense StatName = "total_coins_depense"

	StatTotalRien    StatName = "total_rien"
	StatTotalSunburn StatName = "total_sunburn"
	StatTotalCuts    StatName = "total_cuts"

	StatFouilleTransats StatName = "fouille_transats"
	StatFouilleEau      StatName = "fouille_eau"
	StatFouilleRochers  StatName = "fouille_rochers"

	StatCoquillagesTrouves  StatName = "coquillages_trouves"
	StatBouteillesTrouvees  StatName = "bouteilles_trouvees"
	StatCannettesTrouvees   StatName = "cannettes_trouvees"
	StatLunettesTrouvees    StatName = "lunettes_trouvees"
	StatPiecesTrouvees      StatName = "pieces_trouvees"
	StatPincesTrouvees      StatName = "pinces_trouvees"
	StatMsgBouteilleTrouves StatName = "msg_bouteille_trouves"
	StatMontresTrouvees     StatName = "montres_trouvees"
	StatDentsTrouvees       StatName = "dents_trouvees"

	StatEchecsFouilleCooldown StatName = "echecs_fouille_cooldown"
	StatEchecsFouilleMalus    StatName = "echecs_fouille_malus"

	StatVentesTotal       StatName = "ventes_total"
	StatAchatCremeSolaire StatName = "achat_creme_solaire"
	StatAchatPansement    StatName = "achat_pansement"
	StatAchatEau          StatName = "achat_eau"
)

type UserItem struct {
	UserID   string `gorm:"primaryKey;type:varchar(255)"`
	ItemID   string `gorm:"primaryKey;type:varchar(50)"`
	Quantity int
}

type UserStat struct {
	UserID   string `gorm:"primaryKey;type:varchar(255)"`
	StatName string `gorm:"primaryKey;type:varchar(50)"`
	Value    int
}

type User struct {
	ID string `gorm:"primaryKey;type:varchar(255)"`

	Coins     int
	CDFouille int

	StatusMalus StatusMalus
	CDMalus     int

	Tier Tier

	SabotageImmunityExpiresAt int64

	Items []UserItem `gorm:"foreignKey:UserID;references:ID"`
	Stats []UserStat `gorm:"foreignKey:UserID;references:ID"`
}
