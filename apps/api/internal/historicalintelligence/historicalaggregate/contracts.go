package historicalaggregate

import (
	"context"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const Version = "historical-aggregate-store-v1"

const (
	DefaultListLimit = 20
	MaximumListLimit = 100
)

type ResultKey struct {
	SchemaVersion historicalcontract.SchemaVersion
	MetricName    historicalcontract.MetricName
	Scope         historicalcontract.Scope
	Granularity   historicalcontract.Granularity
	Window        historicalcontract.TimeWindow
}

type Record struct {
	ID               string
	Key              ResultKey
	InputFingerprint string
	Result           historicalcontract.Result
	StoredAt         time.Time
}

func (record Record) Clone() Record {
	cloned := record
	cloned.Result = record.Result.Clone()
	return cloned
}

type ListQuery struct {
	SchemaVersion historicalcontract.SchemaVersion
	MetricName    historicalcontract.MetricName
	Scope         historicalcontract.Scope
	Granularity   historicalcontract.Granularity

	BeforeWindowEnd time.Time
	Limit           int
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
		result historicalcontract.Result,
	) (Record, error)
	Get(
		ctx context.Context,
		key ResultKey,
	) (Record, error)
	GetLatest(
		ctx context.Context,
		query ListQuery,
	) (Record, error)
	List(
		ctx context.Context,
		query ListQuery,
	) (Page, error)
}

type PostgresConfig struct {
	Pool *pgxpool.Pool
	Now  func() time.Time
}

type Executor interface {
	QueryRow(
		ctx context.Context,
		query string,
		args ...any,
	) pgx.Row
	Query(
		ctx context.Context,
		query string,
		args ...any,
	) (pgx.Rows, error)
}
