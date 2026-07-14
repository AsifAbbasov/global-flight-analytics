package routestore

import (
	"context"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

const Version = "route-store-v1"

const (
	DefaultListLimit = 20
	MaximumListLimit = 100
)

type ResultKey struct {
	TrajectoryID  string
	SchemaVersion routecontract.SchemaVersion
	AsOfTime      time.Time
}

type Record struct {
	ID               string
	Key              ResultKey
	InputFingerprint string
	Result           routecontract.Result
	StoredAt         time.Time
}

func (record Record) Clone() Record {
	cloned := record
	cloned.Result = record.Result.Clone()

	return cloned
}

type ListQuery struct {
	TrajectoryID   string
	SchemaVersion  routecontract.SchemaVersion
	BeforeAsOfTime time.Time
	Limit          int
}

type Page struct {
	Records []Record
	HasMore bool
}

func (page Page) Clone() Page {
	cloned := Page{
		Records: make([]Record, 0, len(page.Records)),
		HasMore: page.HasMore,
	}
	for _, record := range page.Records {
		cloned.Records = append(
			cloned.Records,
			record.Clone(),
		)
	}

	return cloned
}

type Store interface {
	Put(
		ctx context.Context,
		result routecontract.Result,
	) (Record, error)
	Get(
		ctx context.Context,
		key ResultKey,
	) (Record, error)
	GetLatest(
		ctx context.Context,
		trajectoryID string,
		schemaVersion routecontract.SchemaVersion,
	) (Record, error)
	List(
		ctx context.Context,
		query ListQuery,
	) (Page, error)
}
