package historicalaggregate

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
)

type storedRow struct {
	id string

	schemaVersion string
	metricName    string

	scopeType           string
	scopeKey            string
	regionCode          string
	airportICAOCode     string
	originICAOCode      string
	destinationICAOCode string

	granularity      string
	inputFingerprint string
	seriesStatus     string
	confidenceLevel  string
	payload          []byte

	windowStartMirror   time.Time
	windowStartUnixNano int64
	windowEndMirror     time.Time
	windowEndUnixNano   int64
	asOfTimeMirror      time.Time
	asOfTimeUnixNano    int64
	storedAtMirror      time.Time
	storedAtUnixNano    int64
}

func scanStoredRow(
	scanner rowScanner,
) (storedRow, error) {
	var row storedRow
	err := scanner.Scan(
		&row.id,
		&row.schemaVersion,
		&row.metricName,
		&row.scopeType,
		&row.scopeKey,
		&row.regionCode,
		&row.airportICAOCode,
		&row.originICAOCode,
		&row.destinationICAOCode,
		&row.granularity,
		&row.inputFingerprint,
		&row.seriesStatus,
		&row.confidenceLevel,
		&row.payload,
		&row.windowStartMirror,
		&row.windowStartUnixNano,
		&row.windowEndMirror,
		&row.windowEndUnixNano,
		&row.asOfTimeMirror,
		&row.asOfTimeUnixNano,
		&row.storedAtMirror,
		&row.storedAtUnixNano,
	)
	if err != nil {
		return storedRow{}, err
	}
	return row, nil
}

func recordFromStoredRow(
	row storedRow,
) (Record, error) {
	exactWindowStart := time.Unix(
		0,
		row.windowStartUnixNano,
	).UTC()
	exactWindowEnd := time.Unix(
		0,
		row.windowEndUnixNano,
	).UTC()
	exactAsOfTime := time.Unix(
		0,
		row.asOfTimeUnixNano,
	).UTC()
	exactStoredAt := time.Unix(
		0,
		row.storedAtUnixNano,
	).UTC()

	for _, timestamp := range []struct {
		field  string
		mirror time.Time
		exact  time.Time
	}{
		{
			field:  "window_start",
			mirror: row.windowStartMirror,
			exact:  exactWindowStart,
		},
		{
			field:  "window_end",
			mirror: row.windowEndMirror,
			exact:  exactWindowEnd,
		},
		{
			field:  "as_of_time",
			mirror: row.asOfTimeMirror,
			exact:  exactAsOfTime,
		},
		{
			field:  "stored_at",
			mirror: row.storedAtMirror,
			exact:  exactStoredAt,
		},
	} {
		if err := validateTimestampMirror(
			timestamp.field,
			timestamp.mirror,
			timestamp.exact,
		); err != nil {
			return Record{}, err
		}
	}

	var result historicalcontract.Result
	if err := json.Unmarshal(
		row.payload,
		&result,
	); err != nil {
		return Record{},
			corruptResult("result_json")
	}
	if _, err := validateStorableResult(
		result,
	); err != nil {
		return Record{},
			corruptResult(
				"result_json.contract",
			)
	}

	if !result.Window.StartTime.UTC().
		Equal(exactWindowStart) {
		return Record{},
			corruptResult(
				"window_start_unix_nano",
			)
	}
	if !result.Window.EndTime.UTC().
		Equal(exactWindowEnd) {
		return Record{},
			corruptResult(
				"window_end_unix_nano",
			)
	}
	if !result.Window.AsOfTime.UTC().
		Equal(exactAsOfTime) {
		return Record{},
			corruptResult(
				"as_of_time_unix_nano",
			)
	}

	rowKey, err := normalizeResultKey(
		ResultKey{
			SchemaVersion: historicalcontract.
				SchemaVersion(row.schemaVersion),
			MetricName: historicalcontract.
				MetricName(row.metricName),
			Scope: historicalcontract.Scope{
				Type: historicalcontract.
					ScopeType(row.scopeType),
				RegionCode: row.regionCode,
				AirportICAOCode: row.
					airportICAOCode,
				OriginICAOCode: row.
					originICAOCode,
				DestinationICAOCode: row.
					destinationICAOCode,
			},
			Granularity: historicalcontract.
				Granularity(row.granularity),
			Window: historicalcontract.TimeWindow{
				StartTime: exactWindowStart,
				EndTime:   exactWindowEnd,
				AsOfTime:  exactAsOfTime,
			},
		},
	)
	if err != nil {
		return Record{},
			corruptResult("stored_key")
	}

	expectedScopeKey, err := scopeKey(
		rowKey.Scope,
	)
	if err != nil ||
		expectedScopeKey != row.scopeKey {
		return Record{},
			corruptResult("scope_key")
	}

	resultKeyValue, err := normalizeResultKey(
		resultKey(result),
	)
	if err != nil ||
		!resultKeysEqual(
			rowKey,
			resultKeyValue,
		) {
		return Record{},
			corruptResult(
				"result_json.key",
			)
	}
	if string(result.Status) !=
		row.seriesStatus {
		return Record{},
			corruptResult(
				"series_status",
			)
	}
	if string(result.Confidence.Level) !=
		row.confidenceLevel {
		return Record{},
			corruptResult(
				"confidence_level",
			)
	}
	if result.Provenance.InputFingerprint !=
		row.inputFingerprint {
		return Record{},
			corruptResult(
				"input_fingerprint",
			)
	}

	compositeKey, err := encodeResultKey(
		rowKey,
	)
	if err != nil {
		return Record{},
			corruptResult(
				"record_key_encoding",
			)
	}
	expectedID := makeRecordID(
		compositeKey,
		row.inputFingerprint,
	)
	if row.id != expectedID {
		return Record{},
			corruptResult("id")
	}
	if err := validateStoredAt(
		exactStoredAt,
		result.GeneratedAt,
	); err != nil {
		return Record{},
			corruptResult("stored_at")
	}

	return Record{
		ID:               row.id,
		Key:              rowKey,
		InputFingerprint: row.inputFingerprint,
		Result:           result,
		StoredAt:         exactStoredAt,
	}, nil
}

func resultKeysEqual(
	left ResultKey,
	right ResultKey,
) bool {
	return left.SchemaVersion ==
		right.SchemaVersion &&
		left.MetricName ==
			right.MetricName &&
		left.Scope == right.Scope &&
		left.Granularity ==
			right.Granularity &&
		left.Window.StartTime.UTC().
			Equal(
				right.Window.StartTime.UTC(),
			) &&
		left.Window.EndTime.UTC().
			Equal(
				right.Window.EndTime.UTC(),
			) &&
		left.Window.AsOfTime.UTC().
			Equal(
				right.Window.AsOfTime.UTC(),
			)
}

func canonicalResultPayload(
	result historicalcontract.Result,
) ([]byte, error) {
	canonical := normalizeResult(result)
	if _, err := validateStorableResult(
		canonical,
	); err != nil {
		return nil, err
	}
	return json.Marshal(canonical)
}

func resultsHaveSameCanonicalPayload(
	left historicalcontract.Result,
	right historicalcontract.Result,
) (bool, error) {
	leftPayload, err := canonicalResultPayload(
		left,
	)
	if err != nil {
		return false, err
	}
	rightPayload, err := canonicalResultPayload(
		right,
	)
	if err != nil {
		return false, err
	}
	return bytes.Equal(
		leftPayload,
		rightPayload,
	), nil
}

func validateStoredAt(
	storedAt time.Time,
	generatedAt time.Time,
) error {
	if storedAt.IsZero() ||
		generatedAt.IsZero() ||
		storedAt.UTC().Before(
			generatedAt.UTC(),
		) {
		return ErrStoredAtInvalid
	}
	return nil
}

func requireContext(
	ctx context.Context,
) error {
	if ctx == nil {
		return ErrContextRequired
	}
	return ctx.Err()
}

func corruptResult(
	field string,
) error {
	return &CorruptResultError{
		Field: field,
	}
}
