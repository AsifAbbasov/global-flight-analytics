package routeresolver

import (
	"errors"
	"testing"
)

func TestResolveRejectsNilContext(
	t *testing.T,
) {
	resolution, err := (&Resolver{}).Resolve(
		nil,
		Input{},
	)

	if !errors.Is(
		err,
		ErrRouteResolutionContextRequired,
	) {
		t.Fatalf(
			"error = %v, want route resolution context required",
			err,
		)
	}
	if resolution.Version != "" ||
		resolution.Result.SchemaVersion != "" ||
		resolution.Result.TrajectoryID != "" {
		t.Fatalf(
			"resolution = %#v, want empty resolution",
			resolution,
		)
	}
}
