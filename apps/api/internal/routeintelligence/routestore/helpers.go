package routestore

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

const recordIDPrefix = "route-record-"

func normalizeResult(
	result routecontract.Result,
) routecontract.Result {
	normalized := result.Clone()

	normalized.TrajectoryID =
		strings.TrimSpace(normalized.TrajectoryID)
	normalized.IdentityKey =
		strings.TrimSpace(normalized.IdentityKey)
	normalized.FlightID =
		strings.TrimSpace(normalized.FlightID)
	normalized.AircraftID =
		strings.TrimSpace(normalized.AircraftID)
	normalized.ICAO24 = strings.ToUpper(
		strings.TrimSpace(normalized.ICAO24),
	)
	normalized.Callsign =
		strings.TrimSpace(normalized.Callsign)

	normalized.Window.StartTime =
		normalized.Window.StartTime.UTC()
	normalized.Window.EndTime =
		normalized.Window.EndTime.UTC()
	normalized.Window.AsOfTime =
		normalized.Window.AsOfTime.UTC()
	normalized.Provenance.ResolverVersion =
		strings.TrimSpace(
			normalized.Provenance.ResolverVersion,
		)
	normalized.Provenance.InputFingerprint =
		strings.TrimSpace(
			normalized.Provenance.InputFingerprint,
		)
	normalized.Provenance.TrajectoryUpdatedAt =
		normalized.Provenance.TrajectoryUpdatedAt.UTC()
	normalized.GeneratedAt =
		normalized.GeneratedAt.UTC()

	sourceNames := make(
		[]string,
		0,
		len(normalized.Provenance.SourceNames),
	)
	seenSources := make(map[string]struct{})
	for _, sourceName := range normalized.Provenance.SourceNames {
		canonical := strings.TrimSpace(sourceName)
		if canonical == "" {
			continue
		}
		if _, exists := seenSources[canonical]; exists {
			continue
		}
		seenSources[canonical] = struct{}{}
		sourceNames = append(sourceNames, canonical)
	}
	sort.Strings(sourceNames)
	normalized.Provenance.SourceNames = sourceNames

	return normalized
}

func validateStorableResult(
	result routecontract.Result,
) (routecontract.ValidationReport, error) {
	if strings.TrimSpace(result.TrajectoryID) == "" {
		return routecontract.ValidationReport{},
			ErrTrajectoryIDRequired
	}
	if result.SchemaVersion !=
		routecontract.SchemaVersionV1 {
		return routecontract.ValidationReport{},
			ErrUnsupportedSchemaVersion
	}
	if result.Window.AsOfTime.IsZero() {
		return routecontract.ValidationReport{},
			ErrAsOfTimeRequired
	}
	if strings.TrimSpace(
		result.Provenance.InputFingerprint,
	) == "" {
		return routecontract.ValidationReport{},
			ErrInputFingerprintRequired
	}

	report := routecontract.Validate(result)
	if report.Status !=
		routecontract.ValidationStatusValid {
		return report.Clone(), &ValidationError{
			Report: report.Clone(),
		}
	}

	return report.Clone(), nil
}

func resultKey(
	result routecontract.Result,
) ResultKey {
	return ResultKey{
		TrajectoryID:  result.TrajectoryID,
		SchemaVersion: result.SchemaVersion,
		AsOfTime:      result.Window.AsOfTime,
	}
}

func normalizeResultKey(
	key ResultKey,
) (ResultKey, error) {
	trajectoryID, err := normalizeTrajectoryID(
		key.TrajectoryID,
	)
	if err != nil {
		return ResultKey{}, err
	}
	if key.SchemaVersion !=
		routecontract.SchemaVersionV1 {
		return ResultKey{},
			ErrUnsupportedSchemaVersion
	}
	if key.AsOfTime.IsZero() {
		return ResultKey{}, ErrAsOfTimeRequired
	}

	return ResultKey{
		TrajectoryID:  trajectoryID,
		SchemaVersion: key.SchemaVersion,
		AsOfTime:      key.AsOfTime.UTC(),
	}, nil
}

func normalizeListQuery(
	query ListQuery,
) (ListQuery, error) {
	trajectoryID, err := normalizeTrajectoryID(
		query.TrajectoryID,
	)
	if err != nil {
		return ListQuery{}, err
	}
	if query.SchemaVersion !=
		routecontract.SchemaVersionV1 {
		return ListQuery{},
			ErrUnsupportedSchemaVersion
	}

	limit := query.Limit
	if limit == 0 {
		limit = DefaultListLimit
	}
	if limit < 1 || limit > MaximumListLimit {
		return ListQuery{}, ErrInvalidListLimit
	}

	beforeAsOfTime := query.BeforeAsOfTime
	if !beforeAsOfTime.IsZero() {
		beforeAsOfTime = beforeAsOfTime.UTC()
	}

	return ListQuery{
		TrajectoryID:   trajectoryID,
		SchemaVersion:  query.SchemaVersion,
		BeforeAsOfTime: beforeAsOfTime,
		Limit:          limit,
	}, nil
}

func normalizeTrajectoryID(
	trajectoryID string,
) (string, error) {
	normalized := strings.TrimSpace(trajectoryID)
	if normalized == "" {
		return "", ErrTrajectoryIDRequired
	}

	return normalized, nil
}

func encodeResultKey(key ResultKey) string {
	return fmt.Sprintf(
		"%s\x00%s\x00%s",
		key.TrajectoryID,
		key.SchemaVersion,
		key.AsOfTime.UTC().Format(time.RFC3339Nano),
	)
}

func makeRecordID(
	compositeKey string,
	fingerprint string,
) string {
	sum := sha256.Sum256(
		[]byte(compositeKey + "\x00" + fingerprint),
	)

	return recordIDPrefix + hex.EncodeToString(sum[:])
}

func nonNilContext(
	ctx context.Context,
) context.Context {
	if ctx == nil {
		return context.Background()
	}

	return ctx
}
