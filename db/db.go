package db

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DB struct {
	GormDB *gorm.DB
}

func New() *DB {
	db, err := gorm.Open(sqlite.Open("bot.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Echec de la connexion à la base de donnees: %v", err)
	}

	err = db.AutoMigrate(&User{})
	if err != nil {
		log.Fatalf("Erreur lors de la migration de la base de donnees: %v", err)
	}

	return &DB{
		GormDB: db,
	}
}
