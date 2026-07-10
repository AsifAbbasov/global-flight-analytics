package airport

import "errors"

var ErrNotFound = errors.New(
	"airport not found",
)
