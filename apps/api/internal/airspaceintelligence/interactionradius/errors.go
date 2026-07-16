package interactionradius

import "errors"

var (
	ErrInvalidPolicy   = errors.New("interaction radius policy is invalid")
	ErrInvalidRequest  = errors.New("interaction radius request is invalid")
	ErrInvalidDecision = errors.New("interaction radius decision is invalid")
)
