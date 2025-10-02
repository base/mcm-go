package proposal

import "errors"

var (
	// ErrInvalidValidUntil is returned when validUntil is 0 or invalid
	ErrInvalidValidUntil = errors.New("invalid validUntil value")

	// ErrNoInstructions is returned when a proposal has no instructions
	ErrNoInstructions = errors.New("proposal must have at least one instruction")

	// ErrInvalidOpCount is returned when preOpCount > postOpCount
	ErrInvalidOpCount = errors.New("preOpCount must be <= postOpCount")
)
