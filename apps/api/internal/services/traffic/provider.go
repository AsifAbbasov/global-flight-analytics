package traffic

import (
	"context"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

type Provider interface {
	LoadByCallsign(
		ctx context.Context,
		callsign string,
	) ([]flightstate.FlightState, error)
}
