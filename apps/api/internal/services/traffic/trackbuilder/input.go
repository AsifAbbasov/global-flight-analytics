package trackbuilder

import (
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

type InputState struct {
	State        flightstate.FlightState
	QualityScore float64
}
