package historicalseries

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalwindow"
)

const Version = "historical-series-builder-v2"

type CoverageState string

const (
	CoverageStateUnavailable CoverageState = "unavailable"
	CoverageStatePartial     CoverageState = "partial"
	CoverageStateComplete    CoverageState = "complete"
)

type DatasetReadState string

const (
	DatasetReadComplete   DatasetReadState = "complete"
	DatasetReadIncomplete DatasetReadState = "incomplete"
)

type DatasetCoverage struct {
	State        DatasetReadState
	MatchedCount int64
}

type CoverageEvidence struct {
	State        CoverageState
	LoadedCount  int64
	MatchedCount int64
	Ratio        float64
}

type BucketValue struct {
	Bucket      historicalwindow.Bucket
	Value       float64
	SampleCount int
	Coverage    CoverageEvidence
}

type BuildRequest struct {
	Metric historicalcontract.Metric
	Scope  historicalcontract.Scope
	Plan   historicalwindow.Plan

	Values []BucketValue

	BuilderVersion        string
	InputFingerprint      string
	SourceNames           []string
	LatestSourceUpdatedAt time.Time
	GeneratedAt           time.Time

	Limitations []historicalcontract.Limitation
}
