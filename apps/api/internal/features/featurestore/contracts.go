package featurestore

import (
	"context"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

const Version = "flight-feature-store-v1"

const (
	DefaultListLimit = 20
	MaximumListLimit = 100
)

type SnapshotKey struct {
	TrajectoryID  string
	SchemaVersion flightfeatures.SchemaVersion
	AsOfTime      time.Time
}

type Record struct {
	ID               string
	Key              SnapshotKey
	InputFingerprint string
	Features         flightfeatures.FlightFeatures
	StoredAt         time.Time
}

func (record Record) Clone() Record {
	cloned := record
	cloned.Features = record.Features.Clone()

	return cloned
}

type ListQuery struct {
	TrajectoryID   string
	SchemaVersion  flightfeatures.SchemaVersion
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
		features flightfeatures.FlightFeatures,
	) (Record, error)
	Get(
		ctx context.Context,
		key SnapshotKey,
	) (Record, error)
	GetLatest(
		ctx context.Context,
		trajectoryID string,
		schemaVersion flightfeatures.SchemaVersion,
	) (Record, error)
	List(
		ctx context.Context,
		query ListQuery,
	) (Page, error)
}

type MemoryConfig struct {
	Now func() time.Time
}
