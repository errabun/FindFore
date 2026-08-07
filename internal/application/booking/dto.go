package booking

import "fmt"

// dto.go is reserved for booking request/response shapes when domain entities
// are not enough. Prefer entity types until a clear DTO boundary appears.

func errNotImplemented(op string) error {
	return fmt.Errorf("booking: %s not implemented", op)
}
