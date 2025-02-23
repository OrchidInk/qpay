package main

import (
	"fmt"
	"log"

	"qpay/config"
	"qpay/models"
)

func main() {
	config.ConnectDatabase()

	err := config.DB.AutoMigrate(&models.Invoice{})
	if err != nil {
		log.Fatal("❌ Migration failed:", err)
	}

	fmt.Println("✅ Database migrated successfully!")
}
