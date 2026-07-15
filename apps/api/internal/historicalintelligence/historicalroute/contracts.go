package historicalroute

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalread"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalwindow"
)

const Version = "historical-route-intelligence-v1"

type Request struct {
	Snapshot historicalread.Snapshot
	Plan     historicalwindow.Plan

	OriginICAOCode      string
	DestinationICAOCode string

	MetricName  historicalcontract.MetricName
	GeneratedAt time.Time
}
