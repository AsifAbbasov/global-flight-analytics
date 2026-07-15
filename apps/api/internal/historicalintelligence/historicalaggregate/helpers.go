package historicalaggregate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
)

const recordIDPrefix = "historical-aggregate-record-"

func normalizeResult(
	result historicalcontract.Result,
) historicalcontract.Result {
	normalized := result.Clone()

	normalized.Metric.Unit =
		strings.TrimSpace(normalized.Metric.Unit)
	normalized.Scope.RegionCode = strings.ToUpper(
		strings.TrimSpace(normalized.Scope.RegionCode),
	)
	normalized.Scope.AirportICAOCode = strings.ToUpper(
		strings.TrimSpace(normalized.Scope.AirportICAOCode),
	)
	normalized.Scope.OriginICAOCode = strings.ToUpper(
		strings.TrimSpace(normalized.Scope.OriginICAOCode),
	)
	normalized.Scope.DestinationICAOCode = strings.ToUpper(
		strings.TrimSpace(
			normalized.Scope.DestinationICAOCode,
		),
	)

	normalized.Window.StartTime =
		normalized.Window.StartTime.UTC()
	normalized.Window.EndTime =
		normalized.Window.EndTime.UTC()
	normalized.Window.AsOfTime =
		normalized.Window.AsOfTime.UTC()
	normalized.GeneratedAt =
		normalized.GeneratedAt.UTC()
	normalized.Provenance.BuilderVersion =
		strings.TrimSpace(
			normalized.Provenance.BuilderVersion,
		)
	normalized.Provenance.InputFingerprint =
		strings.TrimSpace(
			normalized.Provenance.InputFingerprint,
		)
	normalized.Provenance.LatestSourceUpdatedAt =
		normalized.Provenance.
			LatestSourceUpdatedAt.UTC()

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

	for index := range normalized.Points {
		normalized.Points[index].StartTime =
			normalized.Points[index].StartTime.UTC()
		normalized.Points[index].EndTime =
			normalized.Points[index].EndTime.UTC()
	}
	if normalized.Comparison != nil {
		normalized.Comparison.PreviousWindow.StartTime =
			normalized.Comparison.
				PreviousWindow.StartTime.UTC()
		normalized.Comparison.PreviousWindow.EndTime =
			normalized.Comparison.
				PreviousWindow.EndTime.UTC()
		normalized.Comparison.PreviousWindow.AsOfTime =
			normalized.Comparison.
				PreviousWindow.AsOfTime.UTC()
	}

	return normalized
}

func validateStorableResult(
	result historicalcontract.Result,
) (historicalcontract.ValidationReport, error) {
	if result.SchemaVersion !=
		historicalcontract.SchemaVersionV1 {
		return historicalcontract.ValidationReport{},
			ErrUnsupportedSchemaVersion
	}
	if strings.TrimSpace(
		result.Provenance.InputFingerprint,
	) == "" {
		return historicalcontract.ValidationReport{},
			ErrInputFingerprintRequired
	}

	report := historicalcontract.Validate(result)
	if report.Status !=
		historicalcontract.ValidationStatusValid {
		return report.Clone(), &ValidationError{
			Report: report.Clone(),
		}
	}

	return report.Clone(), nil
}

func resultKey(
	result historicalcontract.Result,
) ResultKey {
	return ResultKey{
		SchemaVersion: result.SchemaVersion,
		MetricName:    result.Metric.Name,
		Scope:         result.Scope,
		Granularity:   result.Granularity,
		Window:        result.Window,
	}
}

func normalizeResultKey(
	key ResultKey,
) (ResultKey, error) {
	if key.SchemaVersion !=
		historicalcontract.SchemaVersionV1 {
		return ResultKey{},
			ErrUnsupportedSchemaVersion
	}
	if key.Window.StartTime.IsZero() ||
		key.Window.EndTime.IsZero() ||
		key.Window.AsOfTime.IsZero() {
		return ResultKey{}, ErrWindowRequired
	}

	scope, err := normalizeScope(key.Scope)
	if err != nil {
		return ResultKey{}, err
	}

	return ResultKey{
		SchemaVersion: key.SchemaVersion,
		MetricName:    key.MetricName,
		Scope:         scope,
		Granularity:   key.Granularity,
		Window: historicalcontract.TimeWindow{
			StartTime: key.Window.StartTime.UTC(),
			EndTime:   key.Window.EndTime.UTC(),
			AsOfTime:  key.Window.AsOfTime.UTC(),
		},
	}, nil
}

func normalizeListQuery(
	query ListQuery,
) (ListQuery, error) {
	if query.SchemaVersion !=
		historicalcontract.SchemaVersionV1 {
		return ListQuery{},
			ErrUnsupportedSchemaVersion
	}

	scope, err := normalizeScope(query.Scope)
	if err != nil {
		return ListQuery{}, err
	}

	limit := query.Limit
	if limit == 0 {
		limit = DefaultListLimit
	}
	if limit < 1 || limit > MaximumListLimit {
		return ListQuery{}, ErrInvalidListLimit
	}

	beforeWindowEnd := query.BeforeWindowEnd
	if !beforeWindowEnd.IsZero() {
		beforeWindowEnd = beforeWindowEnd.UTC()
	}

	return ListQuery{
		SchemaVersion:   query.SchemaVersion,
		MetricName:      query.MetricName,
		Scope:           scope,
		Granularity:     query.Granularity,
		BeforeWindowEnd: beforeWindowEnd,
		Limit:           limit,
	}, nil
}

func normalizeScope(
	scope historicalcontract.Scope,
) (historicalcontract.Scope, error) {
	normalized := historicalcontract.Scope{
		Type: scope.Type,
		RegionCode: strings.ToUpper(
			strings.TrimSpace(scope.RegionCode),
		),
		AirportICAOCode: strings.ToUpper(
			strings.TrimSpace(scope.AirportICAOCode),
		),
		OriginICAOCode: strings.ToUpper(
			strings.TrimSpace(scope.OriginICAOCode),
		),
		DestinationICAOCode: strings.ToUpper(
			strings.TrimSpace(
				scope.DestinationICAOCode,
			),
		),
	}

	switch normalized.Type {
	case historicalcontract.ScopeTypeGlobal:
		if normalized.RegionCode != "" ||
			normalized.AirportICAOCode != "" ||
			normalized.OriginICAOCode != "" ||
			normalized.DestinationICAOCode != "" {
			return historicalcontract.Scope{},
				ErrScopeInvalid
		}

	case historicalcontract.ScopeTypeRegion:
		if normalized.RegionCode == "" ||
			normalized.AirportICAOCode != "" ||
			normalized.OriginICAOCode != "" ||
			normalized.DestinationICAOCode != "" {
			return historicalcontract.Scope{},
				ErrScopeInvalid
		}

	case historicalcontract.ScopeTypeAirport:
		if len(normalized.AirportICAOCode) != 4 ||
			normalized.RegionCode != "" ||
			normalized.OriginICAOCode != "" ||
			normalized.DestinationICAOCode != "" {
			return historicalcontract.Scope{},
				ErrScopeInvalid
		}

	case historicalcontract.ScopeTypeRoute:
		if len(normalized.OriginICAOCode) != 4 ||
			len(normalized.DestinationICAOCode) != 4 ||
			normalized.RegionCode != "" ||
			normalized.AirportICAOCode != "" {
			return historicalcontract.Scope{},
				ErrScopeInvalid
		}

	default:
		return historicalcontract.Scope{},
			ErrScopeInvalid
	}

	return normalized, nil
}

func scopeKey(
	scope historicalcontract.Scope,
) (string, error) {
	normalized, err := normalizeScope(scope)
	if err != nil {
		return "", err
	}

	switch normalized.Type {
	case historicalcontract.ScopeTypeGlobal:
		return "global", nil
	case historicalcontract.ScopeTypeRegion:
		return "region:" +
			normalized.RegionCode, nil
	case historicalcontract.ScopeTypeAirport:
		return "airport:" +
			normalized.AirportICAOCode, nil
	case historicalcontract.ScopeTypeRoute:
		return "route:" +
			normalized.OriginICAOCode +
			":" +
			normalized.DestinationICAOCode, nil
	default:
		return "", ErrScopeInvalid
	}
}

func encodeResultKey(
	key ResultKey,
) (string, error) {
	normalized, err := normalizeResultKey(key)
	if err != nil {
		return "", err
	}
	encodedScope, err := scopeKey(normalized.Scope)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(
		"%s\x00%s\x00%s\x00%s\x00%s\x00%s\x00%s",
		normalized.SchemaVersion,
		normalized.MetricName,
		encodedScope,
		normalized.Granularity,
		normalized.Window.StartTime.
			Format(time.RFC3339Nano),
		normalized.Window.EndTime.
			Format(time.RFC3339Nano),
		normalized.Window.AsOfTime.
			Format(time.RFC3339Nano),
	), nil
}

func makeRecordID(
	compositeKey string,
	fingerprint string,
) string {
	sum := sha256.Sum256(
		[]byte(
			compositeKey +
				"\x00" +
				fingerprint,
		),
	)

	return recordIDPrefix +
		hex.EncodeToString(sum[:])
}

func nonNilContext(
	ctx context.Context,
) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
