package historicalroute

import (
	"errors"
	"fmt"
)

var (
	ErrSnapshotVersionInvalid = errors.New(
		"historical route intelligence requires the current historical read snapshot version",
	)
	ErrSnapshotWindowMismatch = errors.New(
		"historical route intelligence snapshot query window does not match the canonical plan window",
	)
	ErrMetricUnsupported = errors.New(
		"historical route intelligence metric is unsupported",
	)
	ErrMetricScopeUnsupported = errors.New(
		"historical route intelligence metric is unsupported for the requested scope",
	)
	ErrRouteScopeIncomplete = errors.New(
		"historical route scope requires both origin and destination ICAO codes or neither",
	)
	ErrOriginICAOInvalid = errors.New(
		"historical route origin ICAO code must contain four alphanumeric characters",
	)
	ErrDestinationICAOInvalid = errors.New(
		"historical route destination ICAO code must contain four alphanumeric characters",
	)
	ErrRouteRecordIdentityRequired = errors.New(
		"historical route record requires a persisted identifier and trajectory identifier",
	)
	ErrRouteRecordTimeInvalid = errors.New(
		"historical route record as-of time is missing or exceeds the analytical cutoff",
	)
	ErrRouteEvidenceInvalid = errors.New(
		"historical route evidence is invalid",
	)
	ErrRouteScopeCoverageUnavailable = errors.New(
		"historical route pair scope cannot claim coverage from a bounded incomplete route dataset",
	)
	ErrRouteMatchedCountRequired = errors.New(
		"historical route incomplete coverage requires an exact matched-route count",
	)
	ErrRouteMatchedCountInvalid = errors.New(
		"historical route incomplete coverage requires more matched routes than represented routes",
	)
	ErrRouteSourceEvidenceUnavailable = errors.New(
		"historical route source evidence is unavailable for the requested scope",
	)
)

type EvidenceError struct {
	RecordID string
	Err      error
}

func (err *EvidenceError) Error() string {
	if err == nil {
		return "historical route evidence is invalid"
	}
	return fmt.Sprintf(
		"historical route record %q is invalid: %v",
		err.RecordID,
		err.Err,
	)
}

func (err *EvidenceError) Unwrap() error {
	if err == nil || err.Err == nil {
		return ErrRouteEvidenceInvalid
	}
	return errors.Join(ErrRouteEvidenceInvalid, err.Err)
}
