package response

import (
	"github.com/labstack/echo/v4"
)

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func RespondError(c echo.Context, statusCode int, code, message string) error {
	return c.JSON(statusCode, ErrorResponse{
		Code:    code,
		Message: message,
	})
}

const (
	CodeValidationError = "VALIDATION_ERROR"
	CodeNotFound        = "NOT_FOUND"
	CodeUnauthorized    = "UNAUTHORIZED"
	CodeForbidden       = "FORBIDDEN"
	CodeInternalError   = "INTERNAL_ERROR"
	CodeConflict        = "CONFLICT"
)
