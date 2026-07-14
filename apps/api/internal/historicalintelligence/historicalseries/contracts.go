package historicalseries

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalwindow"
)

const Version = "historical-series-builder-v1"

type BucketValue struct {
	Bucket      historicalwindow.Bucket
	Value       float64
	SampleCount int
}

type BuildRequest struct {
	Metric historicalcontract.Metric
	Scope  historicalcontract.Scope
	Plan   historicalwindow.Plan

	Values []BucketValue

	DataCoverageRatio float64

	BuilderVersion        string
	InputFingerprint      string
	SourceNames           []string
	LatestSourceUpdatedAt time.Time
	GeneratedAt           time.Time

	Limitations []historicalcontract.Limitation
}
