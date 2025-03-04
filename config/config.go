package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Config struct {

	Invoice struct {
		URL string
	}

	Database struct {
		Host     string
		Port     int
		User     string
		Password string
		DBName   string
	}

	QPay struct {
		Username      string
		Password      string
		InvoiceCode   string
		URL           string
		ExpireSeconds int
	}

	App struct {
		Timeout    int
		Timezone   string
		URL        string
		AllowedIPs string
		APIKey     string
	}
}

var AppConfig *Config

func LoadConfig() *Config {
	// Load environment variables from .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Info().Msg("⚠️  No .env file found, using system environment variables")
	}

	config := &Config{}
	// URL
	config.Invoice.URL = getEnv("URL", "")
	// Database Config
	config.Database.Host = getEnv("DB_HOST", "localhost")
	config.Database.Port = getEnvAsInt("DB_PORT", 5432)
	config.Database.User = getEnv("DB_USER", "root")
	config.Database.Password = getEnv("DB_PASSWORD", "root")
	config.Database.DBName = getEnv("DB_NAME", "qpay")

	// QPay Config
	config.QPay.Username = getEnv("QPAY_USERNAME", "")
	config.QPay.Password = getEnv("QPAY_PASSWORD", "")
	config.QPay.InvoiceCode = getEnv("QPAY_INVOICE_CODE", "INV-000")
	config.QPay.URL = getEnv("QPAY_URL", "https://merchant.qpay.mn/v2")
	config.QPay.ExpireSeconds = getEnvAsInt("QPAY_INVOICE_EXPIRE_SECONDS", 600)

	// Application Config
	config.App.Timeout = getEnvAsInt("TIMEOUT", 10)
	config.App.Timezone = getEnv("TIMEZONE", "Asia/Ulaanbaatar")
	config.App.URL = getEnv("URL", "http://localhost:1324")
	config.App.AllowedIPs = getEnv("ALLOWED_IPS", "127.0.0.1")
	config.App.APIKey = getEnv("API_KEY", "default-api-key")

	AppConfig = config
	log.Info().Msg("✅ Configuration loaded successfully")
	return config
}

// Helper functions to read environment variables
func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func SetLogger() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}).With().Caller().Logger()

	log.Info().Msg("✅ Logger initialized successfully")
}
