package main

import (
	"net/http"
	"os"

	"qpay/routes/middlewares"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	// Global middleware (your custom middlewares)
	e.Use(middleware.Logger())           // Default Echo logger
	e.Use(middleware.Recover())          // Recover from panics
	e.Use(middlewares.PopulateContext)   // Add request timeout
	e.Use(middlewares.IPAuth)            // Restrict access by IP
	e.Use(middlewares.HeaderAuth)        // Validate API key from headers

	// Example route
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, Echo with Custom Middleware!")
	})

	// Example secured route
	e.GET("/api/v1/invoices/:invoiceID", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "Invoice details here"})
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
