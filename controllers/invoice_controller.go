// controllers/invoice_controller.go
package controllers

import (
	"encoding/json"
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

    // invoiceIdParam := c.Param("invoiceNumber")
	// existingInvoice := models.Invoice{InvoiceID: invoiceIdParam}
    // log.Info().Msgf("Checking invoice: %v", existingInvoice)
    // invoice := models.Invoice{InvoiceID: invoiceIdParam}

	// err := invoice.ReadForInvoiceID(c.Request().Context())
    // if err != nil {
	// 	log.Error().Err(err).Msgf("Could not read invoice: %v", err.Error())
	// 	return c.JSON(http.StatusBadRequest, errResponse{
	// 		Code:    ErrRead.Code,
	// 		Message: err.Error()})
	// }

    // Always create a new invoice without checking existing records
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

        jsonReq, err := json.Marshal(req)
        if err != nil {
            log.Error().Err(err).Msg("Failed to marshal request JSON")
            return
        }

        jsonRes, err := json.Marshal(res)
        if err != nil {
            log.Error().Err(err).Msg("Failed to marshal response JSON")
            return
        }

        log.Info().Msgf("Invoice created: %v", invoiceID)
        invoice.Request = jsonReq
        invoice.Response = jsonRes
        invoice.InvoiceID = invoiceID
        log.Info().Msgf("Invoice: %v", invoice.InvoiceID)

        // Save the invoice to the database
        if err = invoice.Create(c.Request().Context()); err != nil {
            log.Error().Err(err).Msg("Failed to save invoice to database")
            return
        }

        return res, nil
    }

    // Directly call createSendInvoice to always create a new invoice
    res, err := createSendInvoice()
    if err != nil {
        return c.JSON(http.StatusBadRequest, errResponse{
            Code:    ErrCreate.Code,
            Message: err.Error()})
    }

    return c.JSON(http.StatusOK, res)
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
	// getting invoice
	invoiceIdParam := c.Param("invoiceID")
	invoice := models.Invoice{
		InvoiceID: invoiceIdParam,
	}

	err := invoice.ReadForInvoiceID(c.Request().Context())
	if err != nil {
		log.Error().Err(err).Msgf("Could not read invoice: %v", err.Error())
		return c.JSON(http.StatusBadRequest, errResponse{
			Code:    ErrRead.Code,
			Message: err.Error()})
	}
	// checking if invoice is already paid
	if invoice.State == models.Paid {
		return c.JSON(http.StatusOK, response{
			Message: "Success",
			Data:    &echo.Map{"isPaid": invoice.State == models.Paid}})
	}

	// sending check invoice request
	qpayClient, err := q.NewClient()
	if err != nil {
		log.Error().Err(err).Msgf("Could not get qpay client: %v", err.Error())
		return c.JSON(http.StatusBadRequest, errResponse{
			Code:    ErrBind.Code,
			Message: err.Error()})
	}
	isPaid, err := qpayClient.CheckInvoice(invoiceIdParam)
	if err != nil {
		log.Error().Err(err).Msgf("Could not check qpay invoice: %v", err.Error())
		return c.JSON(http.StatusBadRequest, errResponse{
			Code:    ErrRead.Code,
			Message: err.Error()})
	}

	// updating paid
	if isPaid {
		log.Info().Msgf("Invoice is paid: %v", invoiceIdParam)
		err = invoice.Update(c.Request().Context(), models.Invoice{State: models.Paid})
		if err != nil {
			log.Info().Err(err).Msg("Could not update invoice.")
			return c.JSON(http.StatusBadRequest, errResponse{
				Code:    ErrUpdate.Code,
				Message: "Could not update invoice"})
		}
	}
	return c.JSON(http.StatusOK, &echo.Map{"isPaid": invoice.State == models.Paid})
}
