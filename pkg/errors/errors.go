package errors

import "fmt"

// ValidationError represents a validation error
type ValidationError struct {
	Message string
}

// BadRequestError represents a bad request error
type BadRequestError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func (e *BadRequestError) Error() string {
	return e.Message
}

// NotFoundError represents a not found error
type NotFoundError struct {
	Message string
}

func (e *NotFoundError) Error() string {
	return e.Message
}

// ConflictError represents a conflict error (e.g., duplicate record)
type ConflictError struct {
	Message string
}

func (e *ConflictError) Error() string {
	return e.Message
}

// UnauthorizedError represents an unauthorized error
type UnauthorizedError struct {
	Message string
}

func (e *UnauthorizedError) Error() string {
	return e.Message
}

// ForbiddenError represents a forbidden error
type ForbiddenError struct {
	Message string
}

func (e *ForbiddenError) Error() string {
	return e.Message
}

// InternalError represents an internal server error
type InternalError struct {
	Message string
}

func (e *InternalError) Error() string {
	return e.Message
}

// LogicError represents a logic error in the application
// This typically indicates an error in the program flow that should never happen
// in normal operation (e.g., required reference is nil, invalid state transition)
type LogicError struct {
	Message string
}

func (e *LogicError) Error() string {
	return e.Message
}

// ExternalAPIError represents an error from external API calls
type ExternalAPIError struct {
	Message string
	Cause   error
}

// Error implements the error interface
func (e *ExternalAPIError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("external api error: %s (%v)", e.Message, e.Cause)
	}
	return fmt.Sprintf("external api error: %s", e.Message)
}
