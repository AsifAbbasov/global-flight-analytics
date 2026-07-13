package history

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/statistics"
)

type Input struct {
	ICAOCode    string
	Entries     []statistics.Statistics
	GeneratedAt time.Time
}

type History struct {
	ICAOCode    string
	WindowStart time.Time
	WindowEnd   time.Time
	Entries     []statistics.Statistics
	GeneratedAt time.Time
}
