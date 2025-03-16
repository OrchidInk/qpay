package routes

import (
	c "qpay/controllers"
	m "qpay/routes/middlewares"

	"github.com/labstack/echo/v4"
)

func MailRoute(e *echo.Echo) {
	e.POST("/api/v1/mail", c.SendEmailHandler, m.HeaderAuth)
}
