package db

type SpecialCaseType uint8

type GamePosition struct {
	PositionId int `gorm:"column:posid;primaryKey"`
	Piege      SpecialCaseType
	Snowflakes int
}
