package routes

import (
	c "qpay/controllers"
	m "qpay/routes/middlewares"

	"github.com/labstack/echo/v4"
)

func InvoiceRoute(e *echo.Echo) {
	e.POST("/api/v1/invoices", c.CreateInvoice, m.HeaderAuth)
	e.GET("/api/v1/invoices/:invoiceID", c.CheckInvoice, m.HeaderAuth)
	e.GET("/api/v1/invoices/callback/:callbackID", c.Callback)
}
