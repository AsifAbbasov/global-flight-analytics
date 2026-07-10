package flight

import "errors"

var ErrNotFound = errors.New(
	"flight not found",
)
