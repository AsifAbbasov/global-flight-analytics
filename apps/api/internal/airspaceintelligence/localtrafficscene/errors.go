package localtrafficscene

import "errors"

var (
	ErrInvalidPolicy  = errors.New("local traffic scene policy is invalid")
	ErrInvalidRequest = errors.New("local traffic scene request is invalid")
	ErrInvalidScene   = errors.New("local traffic scene is invalid")
)
