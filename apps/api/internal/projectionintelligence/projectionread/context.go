package projectionread

import "context"

func validateReadContext(
	ctx context.Context,
) error {
	if ctx == nil {
		return ErrContextRequired
	}
	return ctx.Err()
}
