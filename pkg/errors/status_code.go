package errors

import "net/http"

// GetStatusCode returns the HTTP status code for the given error
func GetStatusCode(err error) int {
	switch err.(type) {
	case *ValidationError:
		return http.StatusBadRequest
	case *NotFoundError:
		return http.StatusNotFound
	case *ConflictError:
		return http.StatusConflict
	case *UnauthorizedError:
		return http.StatusUnauthorized
	case *ForbiddenError:
		return http.StatusForbidden
	case *ExternalAPIError:
		return http.StatusServiceUnavailable
	case *LogicError:
		return http.StatusInternalServerError
	case *InternalError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// GetErrorCode returns the error code string for the given error
func GetErrorCode(err error) string {
	switch err.(type) {
	case *ValidationError:
		return "VALIDATION_ERROR"
	case *NotFoundError:
		return "NOT_FOUND"
	case *ConflictError:
		return "CONFLICT"
	case *UnauthorizedError:
		return "UNAUTHORIZED"
	case *ForbiddenError:
		return "FORBIDDEN"
	case *ExternalAPIError:
		return "EXTERNAL_API_ERROR"
	case *LogicError:
		return "LOGIC_ERROR"
	case *InternalError:
		return "INTERNAL_ERROR"
	default:
		return "INTERNAL_ERROR"
	}
}
