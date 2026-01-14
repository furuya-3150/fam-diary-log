package errors

import (
	"github.com/furuya-3150/fam-diary-log/pkg/response"
	"github.com/labstack/echo/v4"
)

// RespondWithError sends an error response based on the error type
func RespondWithError(c echo.Context, err error) error {
	statusCode := GetStatusCode(err)
	errCode := GetErrorCode(err)
	return response.RespondError(c, statusCode, errCode, err.Error())
}
