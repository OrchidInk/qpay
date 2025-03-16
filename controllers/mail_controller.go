package controllers

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/labstack/echo/v4"
	"gopkg.in/gomail.v2"
)

type EmailRequest struct {
	Email string `json:"email" form:"email"`
}

// SendEmailHandler handles incoming email requests
func SendEmailHandler(c echo.Context) error {
	// Bind email from either JSON or FormData
	var req EmailRequest
	if err := c.Bind(&req); err != nil {
		log.Println("Error binding request:", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
	}

	recipient := req.Email
	log.Println("Recipient:", recipient)

	if recipient == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Email is required"})
	}

	// Load Email Template
	templatePath := filepath.Join("templates", "email_template.html")
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		log.Println("Error loading template:", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to load email template"})
	}

	var body bytes.Buffer
	err = tmpl.Execute(&body, map[string]string{"Name": "Saruul"})
	if err != nil {
		log.Println("Error rendering template:", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to render template"})
	}

	// Send email
	err = sendMail(recipient, "Test Email via Golang SMTP", body.String())
	if err != nil {
		log.Println("Failed to send email:", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to send email: " + err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Email sent successfully!"})
}

// sendMail sends an email using SMTP
func sendMail(to, subject, body string) error {
	MAIL_USERNAME := os.Getenv("MAIL_USERNAME")
	MAIL_PASSWORD := os.Getenv("MAIL_PASSWORD")
	SMTP_SERVER := os.Getenv("SMTP_SERVER")
	SMTP_PORT := 587 // Change if needed

	// Check if environment variables are set
	if MAIL_USERNAME == "" || MAIL_PASSWORD == "" || SMTP_SERVER == "" {
		log.Println("Missing SMTP credentials")
		return echo.NewHTTPError(http.StatusInternalServerError, "Missing SMTP credentials")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", MAIL_USERNAME)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(SMTP_SERVER, SMTP_PORT, MAIL_USERNAME, MAIL_PASSWORD)

	if err := d.DialAndSend(m); err != nil {
		log.Println("Error sending email:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to send email: "+err.Error())
	}

	log.Println("Email sent successfully to:", to)
	return nil
}
