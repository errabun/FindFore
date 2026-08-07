// Package apperr holds small cross-domain application errors.
package apperr

// ValidationError is a user-facing validation failure (HTTP 400).
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
