package controllers

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/labstack/echo/v4"
	"gopkg.in/gomail.v2"
)

// EmailRequest contains complete order details
type ConfirmedEmail struct {
	Email        string `json:"email" form:"email"`
	CustomerName string `json:"customer_name" form:"customer_name"`
	OrderNumber  string `json:"order_number" form:"order_number"`
	OrderDate    string `json:"order_date" form:"order_date"`
	Quantity     int    `json:"quantity" form:"quantity"`
	TotalPrice   string `json:"total_price" form:"total_price"`
	QrData       string `json:"qr_data" form:"qr_data"`
}

// ConfirmedEmail contains basic order creation info
type Invoice struct {
	Email        string `json:"email" form:"email"`
	CustomerName string `json:"customer_name" form:"customer_name"`
	OrderNumber  string `json:"order_number" form:"order_number"`
	OrderDate    string `json:"order_date" form:"order_date"`
}

// SendInvoice sends a full order email
func SendInvoiceEmail(c echo.Context) error {
	var req Invoice
	if err := c.Bind(&req); err != nil {
		return logAndRespond(c, "Bind Invoice", err, http.StatusBadRequest)
	}
	if req.Email == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Email is required"})
	}
	return sendTemplatedEmail(c, req.Email, "Захиалгын нэхэмжлэх", "invoice.html", req)
}

// SendEmailOrderHandler sends a basic order creation email
func SendConfirmedEmail(c echo.Context) error {
	var req ConfirmedEmail
	if err := c.Bind(&req); err != nil {
		return logAndRespond(c, "Bind ConfirmedEmail", err, http.StatusBadRequest)
	}
	if req.Email == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Email is required"})
	}
	return sendTemplatedEmail(c, req.Email, "Захиалга баталгаажсан", "confirmed.html", req)
}

// sendTemplatedEmail renders template and sends the email
func sendTemplatedEmail(c echo.Context, to, subject, templateName string, data any) error {
	emailBody, err := renderTemplate(templateName, data)
	if err != nil {
		return logAndRespond(c, "Render Template: "+templateName, err, http.StatusInternalServerError)
	}

	if err := sendMail(to, subject, emailBody); err != nil {
		return logAndRespond(c, "Send Mail", err, http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Email sent successfully!"})
}

// renderTemplate loads and executes the given template
func renderTemplate(filename string, data any) (string, error) {
	tmplPath := filepath.Join("templates", filename)
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return "", fmt.Errorf("template parse error: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template execute error: %w", err)
	}
	return buf.String(), nil
}

// sendMail sends an email using SMTP
func sendMail(to, subject, body string) error {
	username, password, host, port := getSMTPConfig()
	if username == "" || password == "" || host == "" {
		return errors.New("missing SMTP credentials")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", username)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(host, port, username, password)
	d.SSL = true
	return d.DialAndSend(m)
}

// getSMTPConfig reads SMTP credentials from environment
func getSMTPConfig() (username, password, host string, port int) {
	return os.Getenv("MAIL_USERNAME"), os.Getenv("MAIL_PASSWORD"), os.Getenv("SMTP_SERVER"), 465
}

// logAndRespond logs the error and sends JSON response
func logAndRespond(c echo.Context, context string, err error, status int) error {
	log.Printf("[%s] %v\n", context, err)
	return c.JSON(status, map[string]string{"error": fmt.Sprintf("%s: %v", context, err)})
}
