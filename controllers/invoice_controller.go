package controllers

import (
	"net/http"
	"os"
	"qpay/helpers"
	"qpay/models"
	q "qpay/qpay"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type RequestBody struct {
	Amount              float64 `json:"amount"`
	InvoiceNumber       string  `json:"invoiceNumber"`
	InvoiceReceiverCode string  `json:"invoiceReceiverCode"`
	CallbackURL         string  `json:"callbackURL"`
}

func CreateInvoice(c echo.Context) error {
	realIP := c.RealIP()

	// Bind request body
	var requestBody RequestBody
	if err := c.Bind(&requestBody); err != nil {
		log.Error().Err(err).Msg("Failed to bind request")
		return c.JSON(http.StatusBadRequest, errResponse{
			Code:    ErrBind.Code,
			Message: err.Error()})
	}

	// Check for an existing invoice with the same InvoiceNumber
	var existingInvoice models.Invoice
	err := models.DB.Where("invoice_number = ?", requestBody.InvoiceNumber).First(&existingInvoice).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Error().Err(err).Msg("Failed to find existing invoice")
		return c.JSON(http.StatusBadRequest, errResponse{
			Code:    ErrRead.Code,
			Message: err.Error()})
	}

	// Function to create and send a new invoice
	createSendInvoice := func() (res map[string]interface{}, err error) {
		expireSecondsEnv := os.Getenv("QPAY_INVOICE_EXPIRE_SECONDS")
		if expireSecondsEnv == "" {
			expireSecondsEnv = "600"
		}

		expireSeconds, err := strconv.Atoi(expireSecondsEnv)
		if err != nil {
			log.Error().Err(err).Msg("Invalid expiry seconds")
			return
		}

		expiryDate := time.Now().Add(time.Duration(expireSeconds) * time.Second)
		convertedExpiryDate, _ := helpers.ConvertDatetimeToTimezone(expiryDate)

		invoice := models.Invoice{
			ID:            uuid.New(),
			IpAddress:     realIP,
			CalledAt:      time.Now(),
			State:         models.Unpaid,
			CallbackURL:   requestBody.CallbackURL,
			InvoiceNumber: requestBody.InvoiceNumber,
			ExpireAt:      &expiryDate,
		}

		qpayClient, err := q.NewClient()
		if err != nil {
			log.Error().Err(err).Msg("Failed to create QPay client")
			return
		}

		// Create invoice request to QPay
		var req map[string]interface{}
		req, res, err = qpayClient.CreateInvoice(requestBody.Amount, requestBody.InvoiceNumber, requestBody.InvoiceReceiverCode, invoice.GenerateCallbackURL(), convertedExpiryDate)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create QPay invoice")
			return
		}

		invoiceID, ok := res["invoice_id"].(string)
		if !ok {
			log.Error().Msgf("Failed to retrieve invoice ID from QPay response: %v", res)
			err = ErrAssertion
			return
		}

		invoice.Request = req
		invoice.Response = res
		invoice.InvoiceID = invoiceID

		// Save the invoice to the database
		if err = models.DB.Create(&invoice).Error; err != nil {
			log.Error().Err(err).Msg("Failed to save invoice to database")
			return
		}

		return res, nil
	}

	now := time.Now()
	if existingInvoice.ID == uuid.Nil || existingInvoice.ExpireAt == nil || existingInvoice.ExpireAt.Before(now) {
		if existingInvoice.ID != uuid.Nil {
			if err = models.DB.Delete(&existingInvoice).Error; err != nil {
				log.Error().Err(err).Msg("Failed to delete expired invoice")
				return c.JSON(http.StatusBadRequest, errResponse{
					Code:    ErrDelete.Code,
					Message: err.Error()})
			}
		}

		res, err := createSendInvoice()
		if err != nil {
			return c.JSON(http.StatusBadRequest, errResponse{
				Code:    ErrCreate.Code,
				Message: err.Error()})
		}
		return c.JSON(http.StatusOK, res)
	}

	return c.JSON(http.StatusOK, existingInvoice.Response)
}

func Callback(c echo.Context) error {
	callbackIDParam := c.Param("callbackID")
	callbackID, err := uuid.Parse(callbackIDParam)
	if err != nil {
		log.Error().Err(err).Msg("Invalid callback ID")
		return c.String(http.StatusBadRequest, "Invalid callback ID")
	}

	var invoice models.Invoice
	err = models.DB.First(&invoice, "id = ?", callbackID).Error
	if err != nil {
		log.Error().Err(err).Msg("Invoice not found")
		return echo.ErrNotFound
	}

	invoice.State = models.Paid
	if err = models.DB.Save(&invoice).Error; err != nil {
		log.Error().Err(err).Msg("Failed to update invoice status")
		return echo.ErrInternalServerError
	}

	invoice.CallCallbackURL()
	return c.String(http.StatusOK, "SUCCESS")
}

func CheckInvoice(c echo.Context) error {
	invoiceIDParam := c.Param("invoiceID")
	var invoice models.Invoice
	err := models.DB.First(&invoice, "invoice_id = ?", invoiceIDParam).Error
	if err != nil {
		log.Error().Err(err).Msg("Invoice not found")
		return c.JSON(http.StatusBadRequest, errResponse{
			Code:    ErrRead.Code,
			Message: err.Error()})
	}

	if invoice.State == models.Paid {
		return c.JSON(http.StatusOK, &echo.Map{"isPaid": true})
	}

	qpayClient, err := q.NewClient()
	if err != nil {
		log.Error().Err(err).Msg("Failed to create QPay client")
		return echo.ErrInternalServerError
	}

	isPaid, err := qpayClient.CheckInvoice(invoiceIDParam)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check QPay invoice status")
		return echo.ErrInternalServerError
	}

	if isPaid {
		invoice.State = models.Paid
		if err = models.DB.Save(&invoice).Error; err != nil {
			log.Error().Err(err).Msg("Failed to update invoice status")
			return echo.ErrInternalServerError
		}
	}

	return c.JSON(http.StatusOK, &echo.Map{"isPaid": invoice.State == models.Paid})
}
