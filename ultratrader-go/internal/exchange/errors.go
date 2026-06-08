package exchange

import (
	"errors"
	"fmt"
)

// Common exchange error types inspired by CCXT.
var (
	ErrExchange          = errors.New("exchange error")
	ErrAuthentication    = fmt.Errorf("%w: authentication failed", ErrExchange)
	ErrPermissionDenied  = fmt.Errorf("%w: permission denied", ErrAuthentication)
	ErrBadRequest        = fmt.Errorf("%w: bad request", ErrExchange)
	ErrBadSymbol         = fmt.Errorf("%w: bad symbol", ErrBadRequest)
	ErrInsufficientFunds = fmt.Errorf("%w: insufficient funds", ErrExchange)
	ErrInvalidOrder      = fmt.Errorf("%w: invalid order", ErrExchange)
	ErrOrderNotFound     = fmt.Errorf("%w: order not found", ErrInvalidOrder)
	ErrNetwork           = errors.New("network error")
	ErrRateLimit         = fmt.Errorf("%w: rate limit exceeded", ErrNetwork)
	ErrRequestTimeout    = fmt.Errorf("%w: request timeout", ErrNetwork)
)

// TypedError provides a unified way to handle exchange-specific errors.
type TypedError struct {
	Code    string
	Message string
	Wrapped error
}

func (e *TypedError) Error() string {
	return fmt.Sprintf("exchange error: %s (%s): %v", e.Code, e.Message, e.Wrapped)
}

func (e *TypedError) Unwrap() error {
	return e.Wrapped
}

// NewTypedError creates a new mapped error.
func NewTypedError(code string, msg string, base error) error {
	return &TypedError{
		Code:    code,
		Message: msg,
		Wrapped: base,
	}
}
