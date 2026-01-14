package response

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type SuccessResponse struct {
	Data interface{} `json:"data"`
}

// RespondSuccess sends a standardized success response
func RespondSuccess(c echo.Context, statusCode int, data interface{}) error {
	if statusCode == http.StatusNoContent {
		return c.JSON(statusCode, nil)
	}
	return c.JSON(statusCode, SuccessResponse{
		Data: data,
	})
}
