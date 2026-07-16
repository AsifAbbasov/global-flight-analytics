package airspaceregionanalytics

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/localtrafficscene"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/proximityscanner"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/separationrisk"
)

func Build(request Request, policy Policy) (Result, error) {
	if err := policy.Validate(); err != nil {
		return Result{}, err
	}
	normalized := normalizeRequest(request)
	if err := validateRequest(normalized, policy); err != nil {
		return Result{}, err
	}

	occupancy, context := buildOccupancyIndex(normalized, policy)
	sectorReports := buildSectorComplexity(context, policy)
	metrics := buildRegionMetrics(normalized, occupancy, sectorReports, policy)
	upstream := meanUpstreamConfidence(normalized.Snapshots)
	confidence := buildConfidence(
		upstream,
		context.meanDataQuality,
		occupancy.Metrics.TemporalCoverage,
		policy,
	)
	result := Result{
		SchemaVersion:    SchemaVersionV1,
		Status:           resultStatusFor(normalized, occupancy, confidence),
		RegionCode:       normalized.RegionCode,
		WindowStart:      normalized.WindowStart,
		WindowEnd:        normalized.WindowEnd,
		Occupancy:        occupancy,
		SectorComplexity: sectorReports,
		Metrics:          metrics,
		Confidence:       confidence,
		Limitations:      resultLimitations(normalized, occupancy),
		Explanations:     resultExplanations(),
		ScopeGuard:       ScopeGuardResearchOnly,
		Provenance:       buildProvenance(normalized, context),
		GeneratedAt:      normalized.GeneratedAt,
	}
	result.Provenance.InputFingerprint = inputFingerprint(result, policy)

	report := Validate(result, policy)
	if report.Status != ValidationStatusValid {
		return Result{}, fmt.Errorf("%w: issues=%v", ErrInvalidResult, report.Issues)
	}
	return result.Clone(), nil
}

func normalizeRequest(request Request) Request {
	normalized := request
	normalized.RegionCode = strings.ToUpper(strings.TrimSpace(request.RegionCode))
	normalized.WindowStart = request.WindowStart.UTC()
	normalized.WindowEnd = request.WindowEnd.UTC()
	normalized.GeneratedAt = request.GeneratedAt.UTC()
	normalized.Snapshots = make([]SnapshotInput, 0, len(request.Snapshots))
	for _, snapshot := range request.Snapshots {
		normalized.Snapshots = append(normalized.Snapshots, SnapshotInput{
			Scene: snapshot.Scene.Clone(),
			Scan:  snapshot.Scan.Clone(),
			Risk:  snapshot.Risk.Clone(),
		})
	}
	sort.Slice(normalized.Snapshots, func(left int, right int) bool {
		leftTime := normalized.Snapshots[left].Scene.AsOfTime
		rightTime := normalized.Snapshots[right].Scene.AsOfTime
		if !leftTime.Equal(rightTime) {
			return leftTime.Before(rightTime)
		}
		leftFingerprint := normalized.Snapshots[left].Scene.Provenance.InputFingerprint
		rightFingerprint := normalized.Snapshots[right].Scene.Provenance.InputFingerprint
		return leftFingerprint < rightFingerprint
	})
	return normalized
}

func validateRequest(request Request, policy Policy) error {
	if request.RegionCode == "" || request.WindowStart.IsZero() ||
		request.WindowEnd.IsZero() || request.GeneratedAt.IsZero() ||
		!request.WindowStart.Before(request.WindowEnd) ||
		request.GeneratedAt.Before(request.WindowEnd) {
		return fmt.Errorf("%w: region and bounded completed window are required", ErrInvalidRequest)
	}
	if len(request.Snapshots) > policy.MaximumSnapshots {
		return fmt.Errorf("%w: snapshot capacity exceeded", ErrInvalidRequest)
	}
	observationCount := 0
	seen := make(map[string]struct{}, len(request.Snapshots))
	for index, snapshot := range request.Snapshots {
		if snapshot.Scene.RegionCode != request.RegionCode ||
			snapshot.Scan.RegionCode != request.RegionCode ||
			snapshot.Risk.RegionCode != request.RegionCode {
			return fmt.Errorf("%w: snapshots[%d] region mismatch", ErrInvalidRequest, index)
		}
		if snapshot.Scene.AsOfTime.IsZero() ||
			!snapshot.Scene.AsOfTime.Equal(snapshot.Scan.AsOfTime) ||
			!snapshot.Scene.AsOfTime.Equal(snapshot.Risk.AsOfTime) ||
			snapshot.Scene.AsOfTime.Before(request.WindowStart) ||
			snapshot.Scene.AsOfTime.After(request.WindowEnd) {
			return fmt.Errorf("%w: snapshots[%d] temporal mismatch", ErrInvalidRequest, index)
		}
		if snapshot.Scene.GeneratedAt.After(request.GeneratedAt) ||
			snapshot.Scan.GeneratedAt.After(request.GeneratedAt) ||
			snapshot.Risk.GeneratedAt.After(request.GeneratedAt) {
			return fmt.Errorf("%w: snapshots[%d] future generated evidence", ErrInvalidRequest, index)
		}
		if !snapshot.Scene.Status.IsKnown() ||
			!snapshot.Scan.Status.IsKnown() ||
			!snapshot.Risk.Status.IsKnown() {
			return fmt.Errorf("%w: snapshots[%d] unknown upstream status", ErrInvalidRequest, index)
		}
		if snapshot.Scene.ScopeGuard != localtrafficscene.ScopeGuardResearchOnly ||
			snapshot.Scan.ScopeGuard != proximityscanner.ScopeGuardResearchOnly ||
			snapshot.Risk.ScopeGuard != separationrisk.ScopeGuardResearchOnly {
			return fmt.Errorf("%w: snapshots[%d] scope guard", ErrInvalidRequest, index)
		}
		if strings.TrimSpace(snapshot.Scene.Provenance.InputFingerprint) == "" ||
			strings.TrimSpace(snapshot.Scan.Provenance.InputFingerprint) == "" ||
			strings.TrimSpace(snapshot.Risk.Provenance.InputFingerprint) == "" ||
			snapshot.Scan.Provenance.SceneFingerprint != snapshot.Scene.Provenance.InputFingerprint ||
			snapshot.Risk.Provenance.ScanFingerprint != snapshot.Scan.Provenance.InputFingerprint {
			return fmt.Errorf("%w: snapshots[%d] provenance chain", ErrInvalidRequest, index)
		}
		key := snapshot.Scene.AsOfTime.UTC().Format(time.RFC3339Nano) + "|" +
			snapshot.Scene.Provenance.InputFingerprint
		if _, exists := seen[key]; exists {
			return fmt.Errorf("%w: duplicate snapshot %q", ErrInvalidRequest, key)
		}
		seen[key] = struct{}{}
		observationCount += len(snapshot.Scene.Aircraft)
		if observationCount > policy.MaximumAircraftObservations {
			return fmt.Errorf("%w: aircraft observation capacity exceeded", ErrInvalidRequest)
		}
	}
	return nil
}

func buildRegionMetrics(
	request Request,
	occupancy TemporalOccupancyIndex,
	reports []SectorComplexityReport,
	policy Policy,
) RegionMetrics {
	metrics := RegionMetrics{
		SnapshotCount:            len(request.Snapshots),
		BucketCount:              occupancy.Metrics.BucketCount,
		UniqueAircraftCount:      occupancy.Metrics.UniqueAircraftCount,
		AircraftObservationCount: occupancy.Metrics.AircraftObservationCount,
		OccupiedCellCount:        occupancy.Metrics.OccupiedCellCount,
		SectorReportCount:        len(reports),
		PeakAircraftPerBucket:    occupancy.Metrics.PeakAircraftPerBucket,
		MeanAircraftPerBucket:    occupancy.Metrics.MeanAircraftPerBucket,
		UnknownAltitudeCount:     occupancy.Metrics.UnknownAltitudeCount,
		TemporalCoverage:         occupancy.Metrics.TemporalCoverage,
		OccupancyTrend:           OccupancyTrendUnavailable,
		HighestComplexityLevel:   ComplexityLevelNone,
	}
	if len(occupancy.Buckets) > 0 {
		metrics.CurrentAircraftCount = occupancy.Buckets[len(occupancy.Buckets)-1].Metrics.AircraftCount
	}

	bucketScores := make(map[string][]float64)
	complexityTotal := 0.0
	for _, report := range reports {
		complexityTotal += report.Score
		bucketScores[report.BucketID] = append(bucketScores[report.BucketID], report.Score)
		metrics.PeakComplexityScore = math.Max(metrics.PeakComplexityScore, report.Score)
		metrics.HighestComplexityLevel = higherComplexityLevel(
			metrics.HighestComplexityLevel,
			report.Level,
		)
		switch report.Level {
		case ComplexityLevelModerate:
			metrics.ModerateSectorCount++
		case ComplexityLevelHigh:
			metrics.HighSectorCount++
		case ComplexityLevelSevere:
			metrics.SevereSectorCount++
		}
	}
	if len(reports) > 0 {
		metrics.MeanComplexityScore = complexityTotal / float64(len(reports))
	}

	pressures := make([]float64, 0, len(occupancy.Buckets))
	for _, bucket := range occupancy.Buckets {
		meanComplexity := mean(bucketScores[bucket.ID])
		density := clampUnit(float64(bucket.Metrics.AircraftCount) / float64(policy.DenseAircraftCount))
		pressure := clampUnit(0.40*density + 0.60*meanComplexity)
		pressures = append(pressures, pressure)
		metrics.PeakAirspacePressureIndex = math.Max(metrics.PeakAirspacePressureIndex, pressure)
	}
	metrics.AirspacePressureIndex = mean(pressures)
	metrics.OccupancyTrend = occupancyTrendFor(pressures, policy)

	for _, snapshot := range request.Snapshots {
		metrics.ContextualRiskCount += snapshot.Risk.Metrics.ContextualCount
		metrics.ElevatedRiskCount += snapshot.Risk.Metrics.ElevatedCount
		metrics.HighRiskCount += snapshot.Risk.Metrics.HighCount
		metrics.IndeterminateRiskCount += snapshot.Risk.Metrics.IndeterminateCount
	}
	return metrics
}

func occupancyTrendFor(pressures []float64, policy Policy) OccupancyTrend {
	if len(pressures) < 2 {
		return OccupancyTrendUnavailable
	}
	difference := pressures[len(pressures)-1] - pressures[0]
	switch {
	case difference >= policy.OccupancyTrendChangeThreshold:
		return OccupancyTrendRising
	case difference <= -policy.OccupancyTrendChangeThreshold:
		return OccupancyTrendFalling
	default:
		return OccupancyTrendStable
	}
}

func resultStatusFor(
	request Request,
	occupancy TemporalOccupancyIndex,
	confidence Confidence,
) ResultStatus {
	if occupancy.Metrics.BucketCount == 0 {
		return ResultStatusUnavailable
	}
	if occupancy.Metrics.TemporalCoverage < 1 ||
		occupancy.Metrics.UnknownAltitudeCount > 0 ||
		upstreamIsLimited(request.Snapshots) ||
		confidence.Level == ConfidenceLevelLow ||
		confidence.Level == ConfidenceLevelNone {
		return ResultStatusLimited
	}
	return ResultStatusComplete
}

func upstreamIsLimited(snapshots []SnapshotInput) bool {
	for _, snapshot := range snapshots {
		if snapshot.Scene.Status != localtrafficscene.ResultStatusComplete ||
			snapshot.Scan.Status != proximityscanner.ResultStatusComplete ||
			snapshot.Risk.Status != separationrisk.ResultStatusComplete {
			return true
		}
	}
	return false
}

func resultLimitations(
	request Request,
	occupancy TemporalOccupancyIndex,
) []Limitation {
	limitations := []Limitation{
		{
			Code:    "research_only_not_operational_airspace_management",
			Message: "Regional occupancy and complexity are research analytics and must not be used for operational separation, controller workload, or air traffic control decisions.",
			Scope:   "operational_use",
		},
		{
			Code:    "synthetic_grid_not_official_sectors",
			Message: "Spatial sectors are generated grid cells and do not represent official airspace sector boundaries.",
			Scope:   "sector_definition",
		},
		{
			Code:    "historical_complexity_baseline_unavailable",
			Message: "This version reports the requested window and does not claim comparison with an established historical complexity baseline.",
			Scope:   "historical_baseline",
		},
	}
	if occupancy.Metrics.BucketCount == 0 {
		limitations = append(limitations, Limitation{
			Code:    "occupancy_evidence_unavailable",
			Message: "No eligible aircraft occupancy evidence was available in the requested window.",
			Scope:   "occupancy_coverage",
		})
	}
	if occupancy.Metrics.TemporalCoverage < 1 {
		limitations = append(limitations, Limitation{
			Code:    "partial_temporal_coverage",
			Message: "One or more expected time buckets contain no eligible scene evidence.",
			Scope:   "temporal_coverage",
		})
	}
	if occupancy.Metrics.UnknownAltitudeCount > 0 {
		limitations = append(limitations, Limitation{
			Code:    "unknown_altitude_occupancy_present",
			Message: "Some aircraft are indexed in an unknown-altitude band and reduce vertical occupancy confidence.",
			Scope:   "vertical_evidence",
		})
	}
	if upstreamIsLimited(request.Snapshots) {
		limitations = append(limitations, Limitation{
			Code:    "limited_upstream_airspace_evidence",
			Message: "At least one local scene, proximity scan, or separation-risk result is limited or unavailable.",
			Scope:   "upstream_evidence",
		})
	}
	return limitations
}

func resultExplanations() []Explanation {
	return []Explanation{
		{
			Code:    "time_bucketed_three_dimensional_occupancy",
			Message: "Aircraft are indexed by time bucket, latitude-longitude grid cell, and altitude band so the same map area can have different occupancy at different times and levels.",
		},
		{
			Code:    "sector_complexity_is_multidimensional",
			Message: "Complexity is not a synonym for density; it also includes pair interactions, determinate risk evidence, heading dispersion, speed variability, and altitude mixing.",
		},
		{
			Code:    "regional_pressure_rollup",
			Message: "The regional pressure index combines normalized bucket density with mean explainable sector complexity across the requested window.",
		},
	}
}

func buildProvenance(request Request, context occupancyContext) Provenance {
	provenance := Provenance{
		SceneFingerprints: make([]string, 0, len(request.Snapshots)),
		ScanFingerprints:  make([]string, 0, len(request.Snapshots)),
		RiskFingerprints:  make([]string, 0, len(request.Snapshots)),
		SourceNames:       append([]string(nil), context.sourceNames...),
		LatestObservedAt:  context.latestObservedAt,
	}
	for _, snapshot := range request.Snapshots {
		provenance.SceneFingerprints = append(
			provenance.SceneFingerprints,
			snapshot.Scene.Provenance.InputFingerprint,
		)
		provenance.ScanFingerprints = append(
			provenance.ScanFingerprints,
			snapshot.Scan.Provenance.InputFingerprint,
		)
		provenance.RiskFingerprints = append(
			provenance.RiskFingerprints,
			snapshot.Risk.Provenance.InputFingerprint,
		)
	}
	return provenance
}

func higherComplexityLevel(left, right ComplexityLevel) ComplexityLevel {
	if complexityRank(right) > complexityRank(left) {
		return right
	}
	return left
}

func complexityRank(level ComplexityLevel) int {
	switch level {
	case ComplexityLevelLow:
		return 1
	case ComplexityLevelModerate:
		return 2
	case ComplexityLevelHigh:
		return 3
	case ComplexityLevelSevere:
		return 4
	default:
		return 0
	}
}

func mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	total := 0.0
	for _, value := range values {
		total += value
	}
	return total / float64(len(values))
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}

func clampUnit(value float64) float64 {
	return math.Min(math.Max(value, 0), 1)
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func positiveFinite(value float64) bool {
	return finite(value) && value > 0
}

func nonNegativeFinite(value float64) bool {
	return finite(value) && value >= 0
}

func unitInterval(value float64) bool {
	return finite(value) && value >= 0 && value <= 1
}
