package main

import (
	"fmt"
	"net/http"
	"qpay/config"
	"qpay/models"
	"qpay/routes"
	"qpay/routes/middlewares"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
)

func main() {
	// Initialize logger first
	config.SetLogger()

	// Load configuration
	config.LoadConfig()
	if config.AppConfig == nil {
		log.Fatal().Msg("❌ AppConfig is nil. Did you call LoadConfig() before ConnectDatabase()?")
		return
	}

	// Connect to the database
	config.ConnectDatabase()

	err := config.DB.AutoMigrate(&models.Invoice{})
	if err != nil {
		log.Fatal().Msg("❌ Migration failed:")
	}
	log.Info().Msg("🚀 Database migrated successfully")

	// Create Echo instance
	e := echo.New()

	// Middleware for request logging
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:      true,
		LogStatus:   true,
		LogRemoteIP: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			log.Info().Msg(fmt.Sprintf("%d %s %s", v.Status, v.URI, v.RemoteIP))
			return nil
		},
	}))

	// CORS Middleware
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization, // <-- Add this
			"X-API-KEY",
		},
	}))

	// Custom middlewares
	e.Use(middlewares.PopulateContext)
	// e.Use(middlewares.IPAuth) // Uncomment if you want IP restriction

	// Example API route
	e.GET("/api", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "Hello, World!",
		})
	})

	// Initialize Invoice routes
	routes.InvoiceRoute(e)
	routes.MailRoute(e)

	// Start the server
	log.Info().Msg("🚀 Server starting on :1323")
	e.Logger.Fatal(e.Start(":1323"))
}
