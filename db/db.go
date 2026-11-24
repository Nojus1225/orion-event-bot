package db

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DB struct {
	GormDB *gorm.DB
}

func New() *DB {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Error,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	db, err := gorm.Open(sqlite.Open("bot.db"), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatalf("Echec de la connexion à la base de donnees: %v", err)
	}

	err = db.AutoMigrate(&User{}, &Event{}, &GamePosition{})
	if err != nil {
		log.Fatalf("Erreur lors de la migration de la base de donnees: %v", err)
	}

	db.FirstOrCreate(&Event{ID: 1})

	return &DB{
		GormDB: db,
	}
}
