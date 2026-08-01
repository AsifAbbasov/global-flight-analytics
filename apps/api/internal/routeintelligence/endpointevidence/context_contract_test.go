package endpointevidence

import (
	"errors"
	"testing"
)

func TestBuildRejectsNilContext(
	t *testing.T,
) {
	result, err := (&Builder{}).Build(
		nil,
		Input{},
	)

	if !errors.Is(
		err,
		ErrEndpointEvidenceContextRequired,
	) {
		t.Fatalf(
			"error = %v, want endpoint evidence context required",
			err,
		)
	}
	if result.Version != "" ||
		result.Endpoint != nil ||
		len(result.Limitations) != 0 ||
		result.InputFingerprint != "" {
		t.Fatalf(
			"result = %#v, want empty result",
			result,
		)
	}
}
