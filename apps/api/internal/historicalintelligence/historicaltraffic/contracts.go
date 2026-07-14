package historicaltraffic

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalread"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalwindow"
)

const Version = "historical-traffic-metrics-v1"

type Request struct {
	Snapshot    historicalread.Snapshot
	Plan        historicalwindow.Plan
	MetricName  historicalcontract.MetricName
	GeneratedAt time.Time
}
