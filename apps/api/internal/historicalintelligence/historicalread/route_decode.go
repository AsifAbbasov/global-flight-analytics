package historicalread

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

func (record RouteRecord) ResultAt(
	asOfTime time.Time,
) (routecontract.Result, bool) {
	result := record.Result.Clone()
	if !record.ResultAvailable {
		if len(record.RouteJSON) == 0 {
			return routecontract.Result{}, false
		}
		if err := json.Unmarshal(record.RouteJSON, &result); err != nil {
			return routecontract.Result{}, false
		}
	}

	if !validRouteResult(result, record.TrajectoryID, asOfTime) {
		return routecontract.Result{}, false
	}
	if !record.EventStartTime.IsZero() &&
		!result.Window.StartTime.Equal(record.EventStartTime) {
		return routecontract.Result{}, false
	}
	if !record.EventEndTime.IsZero() &&
		!result.Window.EndTime.Equal(record.EventEndTime) {
		return routecontract.Result{}, false
	}

	return normalizeRouteResult(result), true
}

// ValidatedResultAt is the strict downstream trust gate for persisted Route
// Intelligence evidence. It preserves ResultAt for legacy callers while adding
// the complete Route Contract and persistence-metadata reconciliation required
// by Historical Route analytics.
func (record RouteRecord) ValidatedResultAt(
	asOfTime time.Time,
) (routecontract.Result, error) {
	cutoff := asOfTime.UTC()
	if asOfTime.IsZero() {
		return routecontract.Result{}, routeResultError(
			record.ID,
			"analytical_cutoff_required",
			ErrRouteResultInvalid,
		)
	}

	result := record.Result.Clone()
	if !record.ResultAvailable {
		if len(record.RouteJSON) == 0 {
			return routecontract.Result{}, routeResultError(
				record.ID,
				"route_payload_unavailable",
				ErrRoutePayloadUnavailable,
			)
		}
		if err := json.Unmarshal(record.RouteJSON, &result); err != nil {
			return routecontract.Result{}, routeResultError(
				record.ID,
				"route_payload_json_invalid",
				fmt.Errorf("%w: %v", ErrRoutePayloadDecode, err),
			)
		}
	}

	report := routecontract.Validate(result)
	if report.Status != routecontract.ValidationStatusValid {
		return routecontract.Result{}, routeResultError(
			record.ID,
			"route_contract_invalid:"+routeValidationCodes(report),
			ErrRouteContractInvalid,
		)
	}
	if err := validateRoutePersistenceMetadata(record, result, report, cutoff); err != nil {
		return routecontract.Result{}, err
	}
	return normalizeRouteResult(result), nil
}

func validateRoutePersistenceMetadata(
	record RouteRecord,
	result routecontract.Result,
	report routecontract.ValidationReport,
	cutoff time.Time,
) error {
	mismatch := func(code string) error {
		return routeResultError(record.ID, code, ErrRouteMetadataMismatch)
	}

	if strings.TrimSpace(record.ID) == "" ||
		strings.TrimSpace(record.TrajectoryID) == "" {
		return mismatch("persisted_identity_required")
	}
	if record.EventStartTime.IsZero() ||
		record.EventEndTime.IsZero() ||
		record.AsOfTime.IsZero() ||
		record.StoredAt.IsZero() {
		return mismatch("persisted_timestamps_required")
	}
	if strings.TrimSpace(record.Status) == "" ||
		strings.TrimSpace(record.ConfidenceLevel) == "" ||
		strings.TrimSpace(record.InputFingerprint) == "" {
		return mismatch("persisted_contract_metadata_required")
	}
	if strings.TrimSpace(record.TrajectoryID) !=
		strings.TrimSpace(result.TrajectoryID) {
		return mismatch("trajectory_identity_mismatch")
	}
	if strings.TrimSpace(record.Status) != string(result.Status) {
		return mismatch("route_status_mismatch")
	}
	if strings.TrimSpace(record.ConfidenceLevel) !=
		string(result.Confidence.Level) {
		return mismatch("confidence_level_mismatch")
	}
	if strings.TrimSpace(record.InputFingerprint) !=
		result.Provenance.InputFingerprint {
		return mismatch("input_fingerprint_mismatch")
	}
	if !record.AsOfTime.Equal(result.Window.AsOfTime) {
		return mismatch("as_of_time_mismatch")
	}
	if !record.EventStartTime.Equal(result.Window.StartTime) {
		return mismatch("event_start_time_mismatch")
	}
	if !record.EventEndTime.Equal(result.Window.EndTime) {
		return mismatch("event_end_time_mismatch")
	}
	if record.ValidationWarningCount != report.WarningCount {
		return mismatch("validation_warning_count_mismatch")
	}

	if result.Window.AsOfTime.After(cutoff) ||
		result.GeneratedAt.After(cutoff) ||
		record.AsOfTime.After(cutoff) ||
		record.StoredAt.After(cutoff) {
		return routeResultError(
			record.ID,
			"route_evidence_after_cutoff",
			ErrRouteEvidenceAfterCutoff,
		)
	}
	if result.GeneratedAt.After(record.StoredAt) {
		return mismatch("generated_after_stored_at")
	}
	if record.PayloadFingerprint != "" && len(record.RouteJSON) > 0 {
		sum := sha256.Sum256(record.RouteJSON)
		actual := "sha256:" + hex.EncodeToString(sum[:])
		if actual != record.PayloadFingerprint {
			return routeResultError(
				record.ID,
				"payload_fingerprint_mismatch",
				ErrRoutePayloadFingerprintMismatch,
			)
		}
	}
	return nil
}

func routeValidationCodes(report routecontract.ValidationReport) string {
	codes := make([]string, 0, report.ErrorCount)
	for _, issue := range report.Issues {
		if issue.Severity == routecontract.ValidationSeverityError {
			codes = append(codes, issue.Code)
		}
	}
	if len(codes) == 0 {
		return "unknown"
	}
	sort.Strings(codes)
	return strings.Join(codes, ",")
}

func normalizeRouteResult(result routecontract.Result) routecontract.Result {
	result.Window.StartTime = result.Window.StartTime.UTC()
	result.Window.EndTime = result.Window.EndTime.UTC()
	result.Window.AsOfTime = result.Window.AsOfTime.UTC()
	result.GeneratedAt = result.GeneratedAt.UTC()
	result.Provenance.TrajectoryUpdatedAt =
		result.Provenance.TrajectoryUpdatedAt.UTC()
	return result.Clone()
}

func validRouteResult(
	result routecontract.Result,
	trajectoryID string,
	asOfTime time.Time,
) bool {
	if result.SchemaVersion != routecontract.SchemaVersionV1 ||
		!knownRouteStatus(result.Status) ||
		strings.TrimSpace(result.TrajectoryID) == "" ||
		(strings.TrimSpace(trajectoryID) != "" &&
			strings.TrimSpace(result.TrajectoryID) != strings.TrimSpace(trajectoryID)) ||
		result.Window.StartTime.IsZero() ||
		result.Window.EndTime.IsZero() ||
		result.Window.AsOfTime.IsZero() ||
		!result.Window.StartTime.Before(result.Window.EndTime) ||
		result.Window.EndTime.After(result.Window.AsOfTime) ||
		result.Window.AsOfTime.After(asOfTime.UTC()) ||
		math.IsNaN(result.Confidence.Score) ||
		math.IsInf(result.Confidence.Score, 0) ||
		result.Confidence.Score < 0 ||
		result.Confidence.Score > 1 {
		return false
	}
	return true
}

func knownRouteStatus(status routecontract.RouteStatus) bool {
	switch status {
	case routecontract.RouteStatusComplete,
		routecontract.RouteStatusPartial,
		routecontract.RouteStatusUnavailable:
		return true
	default:
		return false
	}
}

func (record RouteRecord) PayloadDigest() string {
	if strings.HasPrefix(record.PayloadFingerprint, "sha256:") {
		return record.PayloadFingerprint
	}

	var payload []byte
	if len(record.RouteJSON) > 0 {
		payload = record.RouteJSON
	} else if record.ResultAvailable {
		encoded, err := json.Marshal(record.Result)
		if err == nil {
			payload = encoded
		}
	}

	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:])
}
