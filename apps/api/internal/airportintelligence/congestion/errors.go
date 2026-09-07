package congestion

import "errors"

var (
	ErrInvalidInput        = errors.New("Airport Congestion Intelligence input is invalid")
	ErrInsufficientHistory = errors.New("Airport Congestion Intelligence history is insufficient")
)
