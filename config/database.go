package config

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// ConnectDatabase initializes the database connection
func ConnectDatabase() {
	// Ensure AppConfig is not nil
	if AppConfig == nil {
		log.Fatal().Msg("❌ AppConfig is nil. Did you call LoadConfig() before ConnectDatabase()?")
		return
	}

	config := AppConfig

	// Format DSN for PostgreSQL
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		config.Database.Host,
		config.Database.User,
		config.Database.Password,
		config.Database.DBName,
		config.Database.Port, // Ensure this is an integer
	)

	// Open GORM database connection
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal().Msgf("❌ Failed to connect to database: %v", err)
	}

	// Validate database connection
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal().Msgf("❌ Failed to get database handle: %v", err)
	}

	if err = sqlDB.Ping(); err != nil {
		log.Fatal().Msgf("❌ Failed to ping database: %v", err)
	}

	// Set database connection pool settings
	sqlDB.SetMaxOpenConns(20)                     // Maximum open connections
	sqlDB.SetMaxIdleConns(10)                     // Maximum idle connections
	sqlDB.SetConnMaxLifetime(30 * time.Minute)    // Connection max lifetime

	// Assign to global DB variable
	DB = db

	log.Info().Msg("🚀 Database connected successfully!")
}
