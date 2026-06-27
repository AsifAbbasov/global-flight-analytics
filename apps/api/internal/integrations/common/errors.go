package integrations

import "errors"

var (
	ErrEmptyBaseURL   = errors.New("integration base url is empty")
	ErrInvalidRequest = errors.New("integration request is invalid")
	ErrUnauthorized   = errors.New("integration request is unauthorized")
	ErrRateLimited    = errors.New("integration request is rate limited")
	ErrNotFound       = errors.New("integration resource was not found")
	ErrBadResponse    = errors.New("integration returned invalid response")
)
