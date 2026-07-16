package interactiongraph

import "errors"

var (
	ErrInvalidRequest = errors.New("interaction graph request is invalid")
	ErrInvalidGraph   = errors.New("interaction graph result is invalid")
)
