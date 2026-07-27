package historicalread

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
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

	result.Window.StartTime = result.Window.StartTime.UTC()
	result.Window.EndTime = result.Window.EndTime.UTC()
	result.Window.AsOfTime = result.Window.AsOfTime.UTC()
	result.GeneratedAt = result.GeneratedAt.UTC()
	result.Provenance.TrajectoryUpdatedAt =
		result.Provenance.TrajectoryUpdatedAt.UTC()

	return result.Clone(), true
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
