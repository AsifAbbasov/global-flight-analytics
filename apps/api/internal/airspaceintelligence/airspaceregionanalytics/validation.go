package airspaceregionanalytics

import (
	"encoding/hex"
	"fmt"
	"math"
	"sort"
	"strings"
)

type ValidationStatus string

const (
	ValidationStatusInvalid ValidationStatus = "invalid"
	ValidationStatusValid   ValidationStatus = "valid"
)

type ValidationIssue struct {
	Path    string
	Message string
}

type ValidationReport struct {
	Status ValidationStatus
	Issues []ValidationIssue
}

func Validate(result Result, policy Policy) ValidationReport {
	issues := make([]ValidationIssue, 0)
	if err := policy.Validate(); err != nil {
		issues = append(issues, ValidationIssue{Path: "policy", Message: err.Error()})
	}
	if result.SchemaVersion != SchemaVersionV1 {
		issues = append(issues, ValidationIssue{Path: "schema_version", Message: "unexpected schema version"})
	}
	if !result.Status.IsKnown() {
		issues = append(issues, ValidationIssue{Path: "status", Message: "unknown result status"})
	}
	if strings.TrimSpace(result.RegionCode) == "" || result.RegionCode != strings.ToUpper(result.RegionCode) {
		issues = append(issues, ValidationIssue{Path: "region_code", Message: "canonical uppercase region code is required"})
	}
	if result.WindowStart.IsZero() || result.WindowEnd.IsZero() ||
		!result.WindowStart.Before(result.WindowEnd) ||
		result.GeneratedAt.Before(result.WindowEnd) {
		issues = append(issues, ValidationIssue{Path: "window", Message: "invalid completed analytical window"})
	}
	if result.ScopeGuard != ScopeGuardResearchOnly {
		issues = append(issues, ValidationIssue{Path: "scope_guard", Message: "research-only scope guard is required"})
	}
	if !validFingerprint(result.Provenance.InputFingerprint) {
		issues = append(issues, ValidationIssue{Path: "provenance.input_fingerprint", Message: "SHA-256 fingerprint is required"})
	}
	if len(result.Provenance.SceneFingerprints) != result.Metrics.SnapshotCount ||
		len(result.Provenance.ScanFingerprints) != result.Metrics.SnapshotCount ||
		len(result.Provenance.RiskFingerprints) != result.Metrics.SnapshotCount {
		issues = append(issues, ValidationIssue{Path: "provenance", Message: "upstream fingerprint count mismatch"})
	}
	if !sortedUniqueStrings(result.Provenance.SourceNames) {
		issues = append(issues, ValidationIssue{Path: "provenance.source_names", Message: "source names must be sorted and unique"})
	}
	issues = validateConfidence(issues, "confidence", result.Confidence)
	issues = validateOccupancy(issues, result, policy)
	issues = validateSectorReports(issues, result, policy)
	issues = validateRegionMetrics(issues, result)
	if len(issues) > 0 {
		return ValidationReport{Status: ValidationStatusInvalid, Issues: issues}
	}
	return ValidationReport{Status: ValidationStatusValid}
}

func validateOccupancy(
	issues []ValidationIssue,
	result Result,
	policy Policy,
) []ValidationIssue {
	index := result.Occupancy
	if index.BucketDuration != policy.TimeBucketDuration ||
		index.LatitudeCellDegrees != policy.LatitudeCellDegrees ||
		index.LongitudeCellDegrees != policy.LongitudeCellDegrees ||
		index.AltitudeBandMeters != policy.AltitudeBandMeters {
		issues = append(issues, ValidationIssue{Path: "occupancy", Message: "occupancy policy publication mismatch"})
	}
	if index.Metrics.BucketCount != len(index.Buckets) ||
		index.Metrics.ExpectedBucketCount < index.Metrics.BucketCount ||
		!unitInterval(index.Metrics.TemporalCoverage) ||
		!nonNegativeFinite(index.Metrics.MeanAircraftPerBucket) {
		issues = append(issues, ValidationIssue{Path: "occupancy.metrics", Message: "invalid occupancy metrics"})
	}

	totalCells := 0
	totalAircraft := 0
	totalUnknownAltitude := 0
	peakAircraft := 0
	peakCells := 0
	uniqueAircraft := make(map[string]struct{})
	previousBucketID := ""
	for bucketIndex, bucket := range index.Buckets {
		path := fmt.Sprintf("occupancy.buckets[%d]", bucketIndex)
		if bucket.ID == "" || (previousBucketID != "" && bucket.ID <= previousBucketID) {
			issues = append(issues, ValidationIssue{Path: path + ".id", Message: "bucket identifiers must be strictly sorted"})
		}
		previousBucketID = bucket.ID
		if bucket.StartTime.IsZero() || bucket.EndTime.Sub(bucket.StartTime) != index.BucketDuration {
			issues = append(issues, ValidationIssue{Path: path + ".time", Message: "bucket duration mismatch"})
		}
		if bucket.Metrics.OccupiedCellCount != len(bucket.Cells) ||
			bucket.Metrics.AircraftCount < 0 ||
			bucket.Metrics.UnknownAltitudeCount < 0 ||
			!unitInterval(bucket.Metrics.MeanQualityScore) {
			issues = append(issues, ValidationIssue{Path: path + ".metrics", Message: "invalid bucket metrics"})
		}
		bucketAircraft := 0
		bucketUnknown := 0
		previousCellID := ""
		for cellIndex, cell := range bucket.Cells {
			cellPath := fmt.Sprintf("%s.cells[%d]", path, cellIndex)
			if cell.ID == "" || cell.BucketID != bucket.ID ||
				(previousCellID != "" && cell.ID <= previousCellID) {
				issues = append(issues, ValidationIssue{Path: cellPath + ".id", Message: "cell identifiers must be canonical and sorted"})
			}
			previousCellID = cell.ID
			if cell.AircraftCount != len(cell.AircraftNodeIDs) || cell.AircraftCount <= 0 ||
				!unitInterval(cell.MeanQualityScore) ||
				!sortedUniqueStrings(cell.AircraftNodeIDs) {
				issues = append(issues, ValidationIssue{Path: cellPath, Message: "invalid cell membership"})
			}
			if !cell.AltitudeKnown && cell.AltitudeBandIndex != -1 {
				issues = append(issues, ValidationIssue{Path: cellPath + ".altitude_band", Message: "unknown altitude must use band -1"})
			}
			for _, nodeID := range cell.AircraftNodeIDs {
				uniqueAircraft[nodeID] = struct{}{}
			}
			bucketAircraft += cell.AircraftCount
			if !cell.AltitudeKnown {
				bucketUnknown += cell.AircraftCount
			}
		}
		if bucketAircraft != bucket.Metrics.AircraftCount ||
			bucketUnknown != bucket.Metrics.UnknownAltitudeCount {
			issues = append(issues, ValidationIssue{Path: path + ".metrics", Message: "bucket aggregate mismatch"})
		}
		totalCells += len(bucket.Cells)
		totalAircraft += bucketAircraft
		totalUnknownAltitude += bucketUnknown
		peakAircraft = maxInt(peakAircraft, bucketAircraft)
		peakCells = maxInt(peakCells, len(bucket.Cells))
	}
	if totalCells != index.Metrics.OccupiedCellCount ||
		totalAircraft != index.Metrics.AircraftObservationCount ||
		totalUnknownAltitude != index.Metrics.UnknownAltitudeCount ||
		len(uniqueAircraft) != index.Metrics.UniqueAircraftCount ||
		peakAircraft != index.Metrics.PeakAircraftPerBucket ||
		peakCells != index.Metrics.PeakOccupiedCells {
		issues = append(issues, ValidationIssue{Path: "occupancy.metrics", Message: "occupancy aggregate mismatch"})
	}
	return issues
}

func validateSectorReports(
	issues []ValidationIssue,
	result Result,
	policy Policy,
) []ValidationIssue {
	previousID := ""
	for index, report := range result.SectorComplexity {
		path := fmt.Sprintf("sector_complexity[%d]", index)
		if report.ID == "" || (previousID != "" && report.ID <= previousID) {
			issues = append(issues, ValidationIssue{Path: path + ".id", Message: "sector reports must be strictly sorted"})
		}
		previousID = report.ID
		if report.BucketID == "" || report.AircraftCount != len(report.AircraftNodeIDs) ||
			report.AircraftCount <= 0 || !sortedUniqueStrings(report.AircraftNodeIDs) {
			issues = append(issues, ValidationIssue{Path: path, Message: "invalid sector membership"})
		}
		if !unitInterval(report.Score) || report.Level != complexityLevelForScore(report.Score, policy) ||
			!unitInterval(report.HeadingDispersion) || !unitInterval(report.SpeedVariability) {
			issues = append(issues, ValidationIssue{Path: path + ".score", Message: "invalid complexity score or level"})
		}
		if len(report.Components) != 6 {
			issues = append(issues, ValidationIssue{Path: path + ".components", Message: "six complexity components are required"})
		} else {
			weightTotal := 0.0
			for componentIndex, component := range report.Components {
				if !unitInterval(component.Score) || !nonNegativeFinite(component.Weight) {
					issues = append(issues, ValidationIssue{
						Path:    fmt.Sprintf("%s.components[%d]", path, componentIndex),
						Message: "invalid score component",
					})
				}
				weightTotal += component.Weight
			}
			if math.Abs(weightTotal-1) > 1e-9 {
				issues = append(issues, ValidationIssue{Path: path + ".components", Message: "component weights must sum to one"})
			}
		}
		issues = validateConfidence(issues, path+".confidence", report.Confidence)
		if len(report.Limitations) == 0 || len(report.Explanations) == 0 {
			issues = append(issues, ValidationIssue{Path: path, Message: "limitations and explanations are required"})
		}
	}
	if len(result.SectorComplexity) != result.Metrics.SectorReportCount {
		issues = append(issues, ValidationIssue{Path: "metrics.sector_report_count", Message: "sector report count mismatch"})
	}
	return issues
}

func validateRegionMetrics(issues []ValidationIssue, result Result) []ValidationIssue {
	metrics := result.Metrics
	if metrics.BucketCount != result.Occupancy.Metrics.BucketCount ||
		metrics.UniqueAircraftCount != result.Occupancy.Metrics.UniqueAircraftCount ||
		metrics.AircraftObservationCount != result.Occupancy.Metrics.AircraftObservationCount ||
		metrics.OccupiedCellCount != result.Occupancy.Metrics.OccupiedCellCount ||
		metrics.PeakAircraftPerBucket != result.Occupancy.Metrics.PeakAircraftPerBucket ||
		metrics.UnknownAltitudeCount != result.Occupancy.Metrics.UnknownAltitudeCount ||
		metrics.TemporalCoverage != result.Occupancy.Metrics.TemporalCoverage {
		issues = append(issues, ValidationIssue{Path: "metrics", Message: "region and occupancy metrics mismatch"})
	}
	if !unitInterval(metrics.MeanComplexityScore) ||
		!unitInterval(metrics.PeakComplexityScore) ||
		!unitInterval(metrics.AirspacePressureIndex) ||
		!unitInterval(metrics.PeakAirspacePressureIndex) ||
		!metrics.OccupancyTrend.IsKnown() ||
		!metrics.HighestComplexityLevel.IsKnown() {
		issues = append(issues, ValidationIssue{Path: "metrics", Message: "invalid region analytics metrics"})
	}
	if len(result.Limitations) == 0 || len(result.Explanations) == 0 {
		issues = append(issues, ValidationIssue{Path: "result", Message: "limitations and explanations are required"})
	}
	return issues
}

func validateConfidence(
	issues []ValidationIssue,
	path string,
	confidence Confidence,
) []ValidationIssue {
	if !unitInterval(confidence.Score) || !confidence.Level.IsKnown() ||
		len(confidence.Components) != 5 || len(confidence.Reasons) == 0 {
		return append(issues, ValidationIssue{Path: path, Message: "invalid confidence contract"})
	}
	weightTotal := 0.0
	for index, component := range confidence.Components {
		if !unitInterval(component.Score) || !nonNegativeFinite(component.Weight) {
			issues = append(issues, ValidationIssue{
				Path:    fmt.Sprintf("%s.components[%d]", path, index),
				Message: "invalid confidence component",
			})
		}
		weightTotal += component.Weight
	}
	if math.Abs(weightTotal-1) > 1e-9 {
		issues = append(issues, ValidationIssue{Path: path + ".components", Message: "confidence weights must sum to one"})
	}
	return issues
}

func validFingerprint(value string) bool {
	if len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func sortedUniqueStrings(values []string) bool {
	if !sort.StringsAreSorted(values) {
		return false
	}
	for index, value := range values {
		if strings.TrimSpace(value) == "" || (index > 0 && values[index-1] == value) {
			return false
		}
	}
	return true
}
