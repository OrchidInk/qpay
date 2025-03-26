package middlewares

import (
	"context"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

func PopulateContext(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		timeout, err := strconv.Atoi(os.Getenv("TIMEOUT"))
		if err != nil {
			timeout = 10 // Default timeout
		}

		ctx, cancel := context.WithTimeout(c.Request().Context(), time.Duration(timeout)*time.Second)
		defer cancel()

		c.SetRequest(c.Request().WithContext(ctx))
		return next(c)
	}
}

// IPAuth restricts access to specific IPs, including local network IPs
func IPAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Allow specific callback path without IP validation
		if c.Path() == "/api/v1/invoices/:invoiceID" {
			return next(c)
		}

		ipAddress := c.RealIP()

		// Allow local network IPs (172.16.0.0/12, 10.0.0.0/8, 192.168.0.0/16) & localhost (::1)
		if isPrivateIP(ipAddress) || ipAddress == "::1" {
			return next(c)
		}

		// Fetch allowed IPs from environment variable
		allowedIPs := strings.Split(os.Getenv("ALLOWED_IPS"), ",")
		for i := range allowedIPs {
			allowedIPs[i] = strings.TrimSpace(allowedIPs[i])
		}

		// Check if the request IP is in the allowed list
		for _, allowedIP := range allowedIPs {
			if ipAddress == allowedIP {
				return next(c)
			}
		}

		c.Logger().Warnf("Unauthorized access attempt from IP: %s", ipAddress)
		return echo.ErrUnauthorized
	}
}

// HeaderAuth validates requests using the X-API-KEY header
func HeaderAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		key := c.Request().Header.Get("X-API-KEY")
		expectedKey := os.Getenv("API_KEY")

		if key != expectedKey {
			c.Logger().Warn("Invalid API Key provided")
			return echo.ErrUnauthorized
		}

		return next(c)
	}
}

// isPrivateIP checks if an IP address belongs to private ranges
func isPrivateIP(ip string) bool {
    // Define CIDR ranges for private networks.
    cidrRanges := []string{
        "172.16.0.0/12",
        "10.0.0.0/8",
        "192.168.0.0/16",
    }

    // Define single IP addresses.
    singleIPs := []string{
        "127.0.0.1",         // loopback
        "103.50.205.86",     // your server IP
    }

    // Check if the IP matches any of the single IPs.
    for _, single := range singleIPs {
        if ip == single {
            return true
        }
    }

    // Check if the IP is within any of the CIDR ranges.
    for _, cidr := range cidrRanges {
        _, ipnet, err := net.ParseCIDR(cidr)
        if err == nil && ipnet.Contains(net.ParseIP(ip)) {
            return true
        }
    }

    return false
}
