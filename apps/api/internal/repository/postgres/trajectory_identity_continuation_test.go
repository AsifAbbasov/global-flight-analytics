package postgres

import (
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func TestValidatePersistedFlightIdentityAcceptsContinuationSplitReason(
	t *testing.T,
) {
	item := validPersistedIdentityTrajectory()
	item.SplitReason =
		trajectory.FlightSplitReasonContinuedFromPreviousBatch

	if err := validatePersistedFlightIdentity(item); err != nil {
		t.Fatalf(
			"expected continuation split reason to be accepted, got %v",
			err,
		)
	}
}
