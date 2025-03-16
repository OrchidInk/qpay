package qpay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog/log"
)

type QpayClient struct {
	username          string
	password          string
	invoiceCode       string
	AccessToken       string `json:"access_token"`
	AccessTokenExpire int64  `json:"expires_in"`
	httpClient        *http.Client
}

var Client *QpayClient

func NewClient() (c *QpayClient, err error) {
	if Client != nil {
		c = Client
		return
	}

	c = &QpayClient{
		username:    os.Getenv("QPAY_USERNAME"),
		password:    os.Getenv("QPAY_PASSWORD"),
		invoiceCode: os.Getenv("QPAY_INVOICE_CODE"),
		httpClient:  &http.Client{},
	}
	Client = c
	return
}

func (c *QpayClient) login() (err error) {
	url := os.Getenv("QPAY_URL") + "/auth/token"

	// Create a new POST request
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.username, c.password)

	// Send the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, c)
	return
}

func (c *QpayClient) checkAndUpdateAccessToken() (err error) {
	now := time.Now()
	if c.AccessToken == "" || now.After(time.Unix(c.AccessTokenExpire, 0)) {
		err = c.login()
	}
	return
}

func (c *QpayClient) CreateInvoice(amount float64,
	invoiceNumber, invoiceReceiverCode, callbackURL string, expiryDate time.Time) (req map[string]interface{}, res map[string]interface{}, err error) {

	// checking and updating access token
	err = c.checkAndUpdateAccessToken()
	if err != nil {
		return
	}

	// preparing body
	req = map[string]interface{}{
		"invoice_code":          c.invoiceCode,
		"sender_invoice_no":     invoiceNumber,
		"invoice_description":   invoiceNumber,
		"invoice_receiver_code": invoiceReceiverCode,
		"amount":                amount,
		"callback_url":          callbackURL,
	}

	req["expiry_date"] = expiryDate.Format("2006-01-02 15:04:05")

	byteSlice, err := json.Marshal(req)
	if err != nil {
		return
	}

	// Create a new POST request
	url := os.Getenv("QPAY_URL") + "/invoice"
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(byteSlice))
	if err != nil {
		return
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.AccessToken))

	// Send the request
	resp, err := c.httpClient.Do(request)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &res)
	return
}

func (c *QpayClient) CheckInvoice(InvoiceID string) (isPaid bool, paymentID string,  err error) {

	// checking and updating access token
	err = c.checkAndUpdateAccessToken()
	if err != nil {
		return
	}

	// preparing body
	req := map[string]interface{}{
		"object_type": "INVOICE",
		"object_id":   InvoiceID,
	}
	byteSlice, err := json.Marshal(req)
	if err != nil {
		return
	}

	// Create a new POST request
	url := os.Getenv("QPAY_URL") + "/payment/check"
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(byteSlice))
	if err != nil {
		return
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.AccessToken))

	// Send the request
	resp, err := c.httpClient.Do(request)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	res := map[string]interface{}{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return
	}


	// Retrieve and print the value of the "name" key
	isPaid = res["count"].(float64) > 0

	if rows, ok := res["rows"].([]interface{}); ok && len(rows) > 0 {
		if firstRow, ok := rows[0].(map[string]interface{}); ok {
			if id, ok := firstRow["payment_id"].(string); ok {
				paymentID = id
			}
		}
	}

	log.Info().Msgf("isPaid: %v, paymentID: %v,  Check qpay: %v", isPaid, paymentID, res)
	return
}
