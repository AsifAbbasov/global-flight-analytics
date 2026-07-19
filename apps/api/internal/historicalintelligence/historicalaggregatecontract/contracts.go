package historicalaggregatecontract

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
)

const Version = "historical-aggregate-store-v1"

const (
	DefaultListLimit = 20
	MaximumListLimit = 100
)

var (
	ErrUnsupportedSchemaVersion = errors.New(
		"historical aggregate schema version is unsupported",
	)
	ErrInputFingerprintRequired = errors.New(
		"historical aggregate input fingerprint is required",
	)
	ErrInvalidListLimit = errors.New(
		"historical aggregate list limit is invalid",
	)
	ErrResultNotFound = errors.New(
		"historical aggregate result was not found",
	)
	ErrResultConflict = errors.New(
		"historical aggregate result key already exists with a different input fingerprint",
	)
	ErrScopeInvalid = errors.New(
		"historical aggregate scope is invalid",
	)
	ErrWindowRequired = errors.New(
		"historical aggregate window is required",
	)
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
		Records: make(
			[]Record,
			0,
			len(page.Records),
		),
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

type Reader interface {
	GetLatest(
		ctx context.Context,
		query ListQuery,
	) (Record, error)
	List(
		ctx context.Context,
		query ListQuery,
	) (Page, error)
}

type Store interface {
	Reader
	Put(
		ctx context.Context,
		result historicalcontract.Result,
	) (Record, error)
	Get(
		ctx context.Context,
		key ResultKey,
	) (Record, error)
}

type ValidationError struct {
	Report historicalcontract.ValidationReport
}

func (err *ValidationError) Error() string {
	if err == nil {
		return "historical aggregate validation failed"
	}

	return fmt.Sprintf(
		"historical aggregate validation failed: errors=%d warnings=%d",
		err.Report.ErrorCount,
		err.Report.WarningCount,
	)
}
