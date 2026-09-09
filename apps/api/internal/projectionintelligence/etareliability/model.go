package etareliability

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

const (
	Version       = "eta-reliability-v1"
	EvidenceClass = "historically_recomputed_from_persisted_observations_with_endpoint_proxy"
)

type Status string

const (
	StatusUnavailable Status = "unavailable"
	StatusLimited     Status = "limited"
	StatusComplete    Status = "complete"
)

func (status Status) IsKnown() bool {
	return status == StatusUnavailable || status == StatusLimited || status == StatusComplete
}

type Notice struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Route struct {
	OriginICAOCode      string `json:"origin_icao_code"`
	DestinationICAOCode string `json:"destination_icao_code"`
}

type Metrics struct {
	SampleCount                int     `json:"sample_count"`
	MedianAbsoluteErrorSeconds float64 `json:"median_absolute_error_seconds"`
	P80AbsoluteErrorSeconds    float64 `json:"p80_absolute_error_seconds"`
	WithinFiveMinutesRatio     float64 `json:"within_five_minutes_ratio"`
	WithinTenMinutesRatio      float64 `json:"within_ten_minutes_ratio"`
	IntervalCoverageRatio      float64 `json:"interval_coverage_ratio"`
}

type Result struct {
	Version              string                    `json:"version"`
	Status               Status                    `json:"status"`
	TrajectoryID         string                    `json:"trajectory_id"`
	Route                Route                     `json:"route"`
	Method               projectioncontract.Method `json:"method"`
	TargetLeadSeconds    int64                     `json:"target_lead_seconds"`
	LeadToleranceSeconds int64                     `json:"lead_tolerance_seconds"`
	EndpointRadiusKM     float64                   `json:"endpoint_radius_km"`
	CandidateCount       int                       `json:"candidate_count"`
	EligibleSampleCount  int                       `json:"eligible_sample_count"`
	Metrics              *Metrics                  `json:"metrics,omitempty"`
	EvidenceClass        string                    `json:"evidence_class"`
	Limitations          []Notice                  `json:"limitations"`
	InputFingerprint     string                    `json:"input_fingerprint"`
	GeneratedAt          time.Time                 `json:"generated_at"`
}

var fingerprintPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
var icaoPattern = regexp.MustCompile(`^[A-Z0-9]{4}$`)

func (result Result) Validate() error {
	if result.Version != Version || !result.Status.IsKnown() {
		return fmt.Errorf("ETA reliability metadata is invalid")
	}
	if strings.TrimSpace(result.TrajectoryID) == "" {
		return fmt.Errorf("ETA reliability trajectory identifier is required")
	}
	if !icaoPattern.MatchString(result.Route.OriginICAOCode) || !icaoPattern.MatchString(result.Route.DestinationICAOCode) {
		return fmt.Errorf("ETA reliability route is invalid")
	}
	if strings.TrimSpace(result.Method.Name) == "" || strings.TrimSpace(result.Method.Version) == "" || !result.Method.DecisionClass.IsKnown() {
		return fmt.Errorf("ETA reliability method identity is invalid")
	}
	if result.TargetLeadSeconds <= 0 || result.LeadToleranceSeconds <= 0 || !finitePositive(result.EndpointRadiusKM) {
		return fmt.Errorf("ETA reliability comparison policy is invalid")
	}
	if result.CandidateCount < 0 || result.EligibleSampleCount < 0 || result.EligibleSampleCount > result.CandidateCount {
		return fmt.Errorf("ETA reliability sample counts are invalid")
	}
	if result.EvidenceClass != EvidenceClass || !fingerprintPattern.MatchString(result.InputFingerprint) || result.GeneratedAt.IsZero() {
		return fmt.Errorf("ETA reliability evidence metadata is invalid")
	}
	for _, limitation := range result.Limitations {
		if strings.TrimSpace(limitation.Code) == "" || strings.TrimSpace(limitation.Message) == "" {
			return fmt.Errorf("ETA reliability limitation is invalid")
		}
	}
	if result.Status == StatusUnavailable {
		if result.Metrics != nil {
			return fmt.Errorf("unavailable ETA reliability must not publish metrics")
		}
		return nil
	}
	if result.Metrics == nil || result.Metrics.SampleCount != result.EligibleSampleCount || result.Metrics.SampleCount < 1 {
		return fmt.Errorf("ETA reliability metrics are missing or inconsistent")
	}
	if !finiteNonNegative(result.Metrics.MedianAbsoluteErrorSeconds) ||
		!finiteNonNegative(result.Metrics.P80AbsoluteErrorSeconds) ||
		!ratio(result.Metrics.WithinFiveMinutesRatio) ||
		!ratio(result.Metrics.WithinTenMinutesRatio) ||
		!ratio(result.Metrics.IntervalCoverageRatio) {
		return fmt.Errorf("ETA reliability metrics are invalid")
	}
	if result.Metrics.P80AbsoluteErrorSeconds < result.Metrics.MedianAbsoluteErrorSeconds {
		return fmt.Errorf("ETA reliability percentile ordering is invalid")
	}
	return nil
}

func finitePositive(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value > 0
}
func finiteNonNegative(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0
}
func ratio(value float64) bool { return finiteNonNegative(value) && value <= 1 }
