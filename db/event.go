package db

type Event struct {
	ID         uint `gorm:"primaryKey"`
	IsActive   bool
	LodgeTotal int
}
