package routes

import (
	c "qpay/controllers"
	m "qpay/routes/middlewares"

	"github.com/labstack/echo/v4"
)

func MailRoute(e *echo.Echo) {
	e.POST("/api/v1/mail/invoice", c.SendInvoiceEmail, m.HeaderAuth)
	e.POST("/api/v1/mail/confirmed", c.SendConfirmedEmail, m.HeaderAuth)
}
