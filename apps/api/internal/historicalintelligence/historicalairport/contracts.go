package historicalairport

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalread"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalwindow"
)

const Version = "historical-airport-intelligence-v1"

type Request struct {
	Snapshot        historicalread.Snapshot
	Plan            historicalwindow.Plan
	AirportICAOCode string
	MetricName      historicalcontract.MetricName
	GeneratedAt     time.Time
}
