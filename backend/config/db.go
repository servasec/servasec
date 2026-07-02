package config

import (
	"os"
	"time"

	"github.com/servasec/servasec/backend/debug"
	"github.com/servasec/servasec/backend/migrations"
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

	sqlDB, err := db.DB()
	if err != nil {
		debug.Log("Failed to get underlying sql.DB: %v", err)
		os.Exit(1)
	}
	if err := migrations.Run(sqlDB); err != nil {
		debug.Log("Database migration failed: %v", err)
		os.Exit(1)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	DB = db

	if os.Getenv("SSC_SEED_DATABASE") == "true" {
		SeedDatabase()
	}
	return db
}
