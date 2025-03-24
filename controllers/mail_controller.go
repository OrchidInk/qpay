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

// EmailRequest struct for email request
type EmailRequest struct {
	Email        string `json:"email" form:"email"`
	CustomerName string `json:"customer_name" form:"customer_name"`
	OrderNumber  string `json:"order_number" form:"order_number"`
	OrderDate    string `json:"order_date" form:"order_date"`
	ProductName  string `json:"product_name" form:"product_name"`
	Quantity     int    `json:"quantity" form:"quantity"`
	TotalPrice   string `json:"total_price" form:"total_price"`
	QrData       string `json:"qr_data" form:"qr_data"`
}

// SendEmailHandler handles email requests
func SendEmailHandler(c echo.Context) error {
	var req EmailRequest
	if err := c.Bind(&req); err != nil {
		log.Println("Error binding request:", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
	}

	if req.Email == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Email is required"})
	}

	// Load and render email template
	emailBody, err := renderEmailTemplate(req)
	if err != nil {
		log.Println("Error rendering email template:", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to render email template"})
	}

	// Send email
	err = sendMail(req.Email, "Order Confirmation", emailBody)
	if err != nil {
		log.Println("Failed to send email:", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to send email: " + err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Email sent successfully!"})
}

// renderEmailTemplate loads the email template and populates it with data
func renderEmailTemplate(req EmailRequest) (string, error) {
	templatePath := filepath.Join("templates", "email_template.html")
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return "", err
	}

	var body bytes.Buffer
	err = tmpl.Execute(&body, req)
	if err != nil {
		return "", err
	}

	return body.String(), nil
}

// sendMail sends an email using SMTP
func sendMail(to, subject, body string) error {
	MAIL_USERNAME := os.Getenv("MAIL_USERNAME")
	MAIL_PASSWORD := os.Getenv("MAIL_PASSWORD")
	SMTP_SERVER := os.Getenv("SMTP_SERVER")
	SMTP_PORT := 587 // Change if needed

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
