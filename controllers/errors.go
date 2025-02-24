package controllers

import "errors"

var (
	ErrBind errResponse = errResponse{Code: "9001", Message: "Invalid body"}

	ErrCreate errResponse = errResponse{Code: "8001", Message: "Create error"}
	ErrRead   errResponse = errResponse{Code: "8002", Message: "Read error"}
	ErrUpdate errResponse = errResponse{Code: "8003", Message: "Update error"}
	ErrDelete errResponse = errResponse{Code: "8004", Message: "Delete error"}
)

var ErrAssertion = errors.New("not asserted")
