package config

import (
	"os"

	"github.com/servasec/servasec/backend/debug"
	"github.com/servasec/servasec/backend/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() *gorm.DB {
	dsn := os.Getenv("DATABASE_URL")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		debug.Log("Failed to connect to database: %v", err)
		os.Exit(1)
	}

	err = db.AutoMigrate(
		&models.User{},
		&models.BlacklistedToken{},
		&models.Group{},
		&models.Application{},
		&models.Scan{},
		&models.Finding{},
		&models.Team{},
		&models.TeamMember{},
	)
	if err != nil {
		debug.Log("Failed to migrate database: %v", err)
		os.Exit(1)
	}

	DB = db

	if os.Getenv("SSC_SEED_DATABASE") == "true" {
		SeedDatabase()
	}
	return db
}
