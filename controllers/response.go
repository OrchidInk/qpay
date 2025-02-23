package controllers

import "github.com/labstack/echo/v4"

type response struct {
	Message string    `json:"message"`
	Data    *echo.Map `json:"data"`
}

type errResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
