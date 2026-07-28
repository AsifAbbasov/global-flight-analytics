package historicalread

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

func TestValidatedResultAtAcceptsCompleteContractAndMetadata(t *testing.T) {
	result := strictRouteTestResult(
		routecontract.RouteStatusComplete,
		"trajectory-valid",
	)
	record := strictRouteTestRecord(t, "record-valid", result)

	decoded, err := record.ValidatedResultAt(result.Window.AsOfTime)
	if err != nil {
		t.Fatalf("ValidatedResultAt() error = %v", err)
	}
	if decoded.TrajectoryID != result.TrajectoryID ||
		decoded.Status != routecontract.RouteStatusComplete {
		t.Fatalf("decoded result = %#v", decoded)
	}
}

func TestValidatedResultAtRejectsInvalidRouteContract(t *testing.T) {
	result := strictRouteTestResult(
		routecontract.RouteStatusPartial,
		"trajectory-invalid-contract",
	)
	record := strictRouteTestRecord(t, "record-invalid-contract", result)

	invalid := result.Clone()
	invalid.Destination = strictRouteTestEndpoint(
		routecontract.EndpointRoleDestination,
		"UGTB",
		41.6692,
		44.9547,
		invalid.Window.EndTime,
		0.5,
	)
	invalid.Confidence.EvidenceCount = 2
	record.Result = invalid
	record.ResultAvailable = true

	_, err := record.ValidatedResultAt(invalid.Window.AsOfTime)
	if !errors.Is(err, ErrRouteContractInvalid) {
		t.Fatalf("ValidatedResultAt() error = %v, want ErrRouteContractInvalid", err)
	}
}

func TestValidatedResultAtRejectsUnsupportedSchema(t *testing.T) {
	result := strictRouteTestResult(
		routecontract.RouteStatusComplete,
		"trajectory-schema",
	)
	result.SchemaVersion = routecontract.SchemaVersion("route-intelligence-v999")
	record := strictRouteTestRecordUnchecked(t, "record-schema", result)

	_, err := record.ValidatedResultAt(result.Window.AsOfTime)
	if !errors.Is(err, ErrRouteContractInvalid) {
		t.Fatalf("ValidatedResultAt() error = %v, want ErrRouteContractInvalid", err)
	}
}

func TestValidatedResultAtRejectsPersistenceMetadataMismatch(t *testing.T) {
	result := strictRouteTestResult(
		routecontract.RouteStatusComplete,
		"trajectory-metadata",
	)
	record := strictRouteTestRecord(t, "record-metadata", result)

	tests := []struct {
		name   string
		mutate func(*RouteRecord)
	}{
		{
			name: "trajectory",
			mutate: func(item *RouteRecord) {
				item.TrajectoryID = "trajectory-other"
			},
		},
		{
			name: "status",
			mutate: func(item *RouteRecord) {
				item.Status = string(routecontract.RouteStatusPartial)
			},
		},
		{
			name: "confidence level",
			mutate: func(item *RouteRecord) {
				item.ConfidenceLevel = string(routecontract.ConfidenceLevelLow)
			},
		},
		{
			name: "input fingerprint",
			mutate: func(item *RouteRecord) {
				item.InputFingerprint = "sha256:" + strings.Repeat("f", 64)
			},
		},
		{
			name: "as of time",
			mutate: func(item *RouteRecord) {
				item.AsOfTime = item.AsOfTime.Add(-time.Second)
			},
		},
		{
			name: "event start",
			mutate: func(item *RouteRecord) {
				item.EventStartTime = item.EventStartTime.Add(time.Second)
			},
		},
		{
			name: "warning count",
			mutate: func(item *RouteRecord) {
				item.ValidationWarningCount++
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			changed := record
			test.mutate(&changed)
			_, err := changed.ValidatedResultAt(result.Window.AsOfTime)
			if !errors.Is(err, ErrRouteMetadataMismatch) {
				t.Fatalf("ValidatedResultAt() error = %v, want ErrRouteMetadataMismatch", err)
			}
		})
	}
}

func TestValidatedResultAtRejectsFutureEvidence(t *testing.T) {
	result := strictRouteTestResult(
		routecontract.RouteStatusComplete,
		"trajectory-future",
	)
	record := strictRouteTestRecord(t, "record-future", result)
	cutoff := result.Window.AsOfTime.Add(-time.Second)

	_, err := record.ValidatedResultAt(cutoff)
	if !errors.Is(err, ErrRouteEvidenceAfterCutoff) {
		t.Fatalf("ValidatedResultAt() error = %v, want ErrRouteEvidenceAfterCutoff", err)
	}
}

func TestValidatedResultAtRejectsInvalidJSONAndPayloadFingerprint(t *testing.T) {
	result := strictRouteTestResult(
		routecontract.RouteStatusComplete,
		"trajectory-json",
	)
	record := strictRouteTestRecord(t, "record-json", result)
	record.Result = routecontract.Result{}
	record.ResultAvailable = false
	record.RouteJSON = []byte("{")

	_, err := record.ValidatedResultAt(result.Window.AsOfTime)
	if !errors.Is(err, ErrRoutePayloadDecode) {
		t.Fatalf("invalid JSON error = %v, want ErrRoutePayloadDecode", err)
	}

	record = strictRouteTestRecord(t, "record-digest", result)
	payload, marshalErr := json.Marshal(result)
	if marshalErr != nil {
		t.Fatalf("marshal route result: %v", marshalErr)
	}
	record.Result = routecontract.Result{}
	record.ResultAvailable = false
	record.RouteJSON = payload
	record.PayloadFingerprint = "sha256:" + strings.Repeat("0", 64)

	_, err = record.ValidatedResultAt(result.Window.AsOfTime)
	if !errors.Is(err, ErrRoutePayloadFingerprintMismatch) {
		t.Fatalf("payload fingerprint error = %v, want ErrRoutePayloadFingerprintMismatch", err)
	}
}

func strictRouteTestRecordUnchecked(
	t *testing.T,
	id string,
	result routecontract.Result,
) RouteRecord {
	t.Helper()
	report := routecontract.Validate(result)
	return strictRouteTestRecordWithWarningCount(t, id, result, report.WarningCount)
}

func strictRouteTestRecord(
	t *testing.T,
	id string,
	result routecontract.Result,
) RouteRecord {
	t.Helper()

	report := routecontract.Validate(result)
	if report.Status != routecontract.ValidationStatusValid {
		t.Fatalf("test route contract is invalid: %#v", report)
	}
	return strictRouteTestRecordWithWarningCount(t, id, result, report.WarningCount)
}

func strictRouteTestRecordWithWarningCount(
	t *testing.T,
	id string,
	result routecontract.Result,
	warningCount int,
) RouteRecord {
	t.Helper()
	payload, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal test route: %v", err)
	}
	sum := sha256.Sum256(payload)

	return RouteRecord{
		ID:                     id,
		TrajectoryID:           result.TrajectoryID,
		EventStartTime:         result.Window.StartTime,
		EventEndTime:           result.Window.EndTime,
		AsOfTime:               result.Window.AsOfTime,
		InputFingerprint:       result.Provenance.InputFingerprint,
		Status:                 string(result.Status),
		ConfidenceLevel:        string(result.Confidence.Level),
		ValidationWarningCount: warningCount,
		StoredAt:               result.GeneratedAt,
		PayloadBytes:           int64(len(payload)),
		PayloadFingerprint:     "sha256:" + hex.EncodeToString(sum[:]),
		Result:                 result.Clone(),
		ResultAvailable:        true,
	}
}

func strictRouteTestResult(
	status routecontract.RouteStatus,
	trajectoryID string,
) routecontract.Result {
	asOfTime := time.Date(2026, time.July, 28, 12, 0, 0, 0, time.UTC)
	startTime := asOfTime.Add(-time.Hour)
	endTime := asOfTime.Add(-time.Minute)
	fingerprint := "sha256:" + strings.Repeat("a", 64)

	result := routecontract.Result{
		SchemaVersion: routecontract.SchemaVersionV1,
		Status:        status,
		TrajectoryID:  trajectoryID,
		ICAO24:        "ABC123",
		Callsign:      "J2001",
		Window: routecontract.RouteWindow{
			StartTime: startTime,
			EndTime:   endTime,
			AsOfTime:  asOfTime,
		},
		Provenance: routecontract.Provenance{
			ResolverVersion:     "route-resolver-test-v1",
			InputFingerprint:    fingerprint,
			TrajectoryUpdatedAt: endTime,
			SourceNames:         []string{"opensky"},
		},
		GeneratedAt: asOfTime,
		Limitations: []routecontract.Limitation{
			{
				Code:    "probable_route_only",
				Message: "Route evidence is inferred rather than filed flight-plan data.",
				Scope:   "route",
			},
		},
	}

	switch status {
	case routecontract.RouteStatusComplete:
		result.Origin = strictRouteTestEndpoint(
			routecontract.EndpointRoleOrigin,
			"UBBB",
			40.4675,
			50.0467,
			startTime,
			0.9,
		)
		result.Destination = strictRouteTestEndpoint(
			routecontract.EndpointRoleDestination,
			"UGTB",
			41.6692,
			44.9547,
			endTime,
			0.9,
		)
		result.Summary.GreatCircleDistanceKM = 450
		result.Confidence = strictRouteTestConfidence(0.9, 2, "route_complete")
	case routecontract.RouteStatusPartial:
		result.Origin = strictRouteTestEndpoint(
			routecontract.EndpointRoleOrigin,
			"UBBB",
			40.4675,
			50.0467,
			startTime,
			0.5,
		)
		result.Confidence = strictRouteTestConfidence(0.5, 1, "route_partial")
	case routecontract.RouteStatusUnavailable:
		result.Confidence = routecontract.Confidence{
			Level: routecontract.ConfidenceLevelNone,
			Reasons: []routecontract.ConfidenceReason{
				{
					Code:    "no_endpoint_evidence",
					Message: "No route endpoint evidence is available.",
				},
			},
		}
	}
	return result
}

func strictRouteTestEndpoint(
	role routecontract.EndpointRole,
	icaoCode string,
	latitude float64,
	longitude float64,
	observedAt time.Time,
	score float64,
) *routecontract.EndpointInference {
	return &routecontract.EndpointInference{
		Role: role,
		Airport: routecontract.AirportReference{
			ICAOCode:           icaoCode,
			Name:               "Test Airport " + icaoCode,
			Latitude:           latitude,
			Longitude:          longitude,
			ElevationM:         10,
			ElevationAvailable: true,
		},
		DistanceKM: 2,
		Confidence: strictRouteTestConfidence(
			score,
			1,
			string(role)+"_confidence",
		),
		Evidence: []routecontract.Evidence{
			{
				Type:          routecontract.EvidenceTypeTrajectoryEndpointProximity,
				SourceName:    "opensky",
				SourceVersion: "route-test-v1",
				Score:         score,
				Weight:        1,
				ObservedAt:    observedAt,
				Summary:       "Validated endpoint evidence.",
				Attributes: []routecontract.EvidenceAttribute{
					{Key: "airport_icao", Value: icaoCode},
				},
			},
		},
	}
}

func strictRouteTestConfidence(
	score float64,
	evidenceCount int,
	code string,
) routecontract.Confidence {
	return routecontract.Confidence{
		Score:         score,
		Level:         routecontract.ConfidenceLevelForScore(score),
		EvidenceCount: evidenceCount,
		Reasons: []routecontract.ConfidenceReason{
			{
				Code:         code,
				Message:      "Validated route confidence.",
				Contribution: score,
			},
		},
	}
}
