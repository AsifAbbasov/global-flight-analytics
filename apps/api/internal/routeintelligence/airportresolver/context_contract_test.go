package airportresolver

import (
	"errors"
	"testing"
)

func TestResolveRejectsNilContext(
	t *testing.T,
) {
	result, err := (&Resolver{}).Resolve(
		nil,
		Query{},
	)

	if !errors.Is(
		err,
		ErrAirportResolutionContextRequired,
	) {
		t.Fatalf(
			"error = %v, want airport resolution context required",
			err,
		)
	}
	if result.Version != "" ||
		len(result.Candidates) != 0 ||
		result.InputFingerprint != "" {
		t.Fatalf(
			"result = %#v, want empty result",
			result,
		)
	}
}
