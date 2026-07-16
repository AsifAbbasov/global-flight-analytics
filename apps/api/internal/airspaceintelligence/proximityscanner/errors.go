package proximityscanner

import "errors"

var (
	ErrInvalidPolicy  = errors.New("proximity scanner policy is invalid")
	ErrInvalidRequest = errors.New("proximity scanner request is invalid")
	ErrInvalidResult  = errors.New("proximity scanner result is invalid")
	ErrGraphBuild     = errors.New("interaction graph build failed")
)
