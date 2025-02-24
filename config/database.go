package config

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// ConnectDatabase initializes the database connection
func ConnectDatabase() {
	if AppConfig == nil {
		log.Fatal("❌ AppConfig is nil. Did you call LoadConfig() before ConnectDatabase()?")
		return
	}

	config := AppConfig

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		config.Database.Host,
		config.Database.User,
		config.Database.Password,
		config.Database.DBName,
		config.Database.Port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}

	// Validate the database connection
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("❌ Failed to get database handle:", err)
	}

	if err = sqlDB.Ping(); err != nil {
		log.Fatal("❌ Failed to ping database:", err)
	}

	DB = db
	log.Println("🚀 Database connected successfully")
}
