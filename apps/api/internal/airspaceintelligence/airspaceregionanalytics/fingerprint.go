package airspaceregionanalytics

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

const fingerprintTimeLayout = time.RFC3339Nano

type fingerprintEnvelope struct {
	SchemaVersion     string
	PolicyVersion     string
	RegionCode        string
	WindowStart       string
	WindowEnd         string
	GeneratedAt       string
	SceneFingerprints []string
	ScanFingerprints  []string
	RiskFingerprints  []string
	Occupancy         fingerprintOccupancy
	SectorComplexity  []fingerprintSector
	Metrics           RegionMetrics
	ConfidenceScore   float64
	Status            string
	ScopeGuard        string
}

type fingerprintOccupancy struct {
	BucketDurationNanoseconds int64
	LatitudeCellDegrees       float64
	LongitudeCellDegrees      float64
	AltitudeBandMeters        float64
	Metrics                   OccupancyIndexMetrics
	Buckets                   []fingerprintBucket
}

type fingerprintBucket struct {
	ID        string
	StartTime string
	EndTime   string
	Metrics   OccupancyBucketMetrics
	Cells     []fingerprintCell
}

type fingerprintCell struct {
	ID                string
	LatitudeIndex     int
	LongitudeIndex    int
	AltitudeBandIndex int
	AltitudeKnown     bool
	AircraftNodeIDs   []string
	MeanQualityScore  float64
}

type fingerprintSector struct {
	ID                     string
	BucketID               string
	AircraftNodeIDs        []string
	AircraftCount          int
	AltitudeBandCount      int
	UnknownAltitudeCount   int
	CandidatePairCount     int
	ConvergingPairCount    int
	ContextualRiskCount    int
	ElevatedRiskCount      int
	HighRiskCount          int
	IndeterminateRiskCount int
	HeadingDispersion      float64
	SpeedVariability       float64
	Score                  float64
	Level                  string
	Components             []ScoreComponent
	ConfidenceScore        float64
}

func inputFingerprint(result Result, policy Policy) string {
	envelope := fingerprintEnvelope{
		SchemaVersion:     string(result.SchemaVersion),
		PolicyVersion:     policy.Version,
		RegionCode:        result.RegionCode,
		WindowStart:       result.WindowStart.UTC().Format(fingerprintTimeLayout),
		WindowEnd:         result.WindowEnd.UTC().Format(fingerprintTimeLayout),
		GeneratedAt:       result.GeneratedAt.UTC().Format(fingerprintTimeLayout),
		SceneFingerprints: append([]string(nil), result.Provenance.SceneFingerprints...),
		ScanFingerprints:  append([]string(nil), result.Provenance.ScanFingerprints...),
		RiskFingerprints:  append([]string(nil), result.Provenance.RiskFingerprints...),
		Metrics:           result.Metrics,
		ConfidenceScore:   result.Confidence.Score,
		Status:            string(result.Status),
		ScopeGuard:        string(result.ScopeGuard),
		Occupancy: fingerprintOccupancy{
			BucketDurationNanoseconds: result.Occupancy.BucketDuration.Nanoseconds(),
			LatitudeCellDegrees:       result.Occupancy.LatitudeCellDegrees,
			LongitudeCellDegrees:      result.Occupancy.LongitudeCellDegrees,
			AltitudeBandMeters:        result.Occupancy.AltitudeBandMeters,
			Metrics:                   result.Occupancy.Metrics,
			Buckets:                   make([]fingerprintBucket, 0, len(result.Occupancy.Buckets)),
		},
		SectorComplexity: make([]fingerprintSector, 0, len(result.SectorComplexity)),
	}
	for _, bucket := range result.Occupancy.Buckets {
		fingerprintBucketValue := fingerprintBucket{
			ID:        bucket.ID,
			StartTime: bucket.StartTime.UTC().Format(fingerprintTimeLayout),
			EndTime:   bucket.EndTime.UTC().Format(fingerprintTimeLayout),
			Metrics:   bucket.Metrics,
			Cells:     make([]fingerprintCell, 0, len(bucket.Cells)),
		}
		for _, cell := range bucket.Cells {
			fingerprintBucketValue.Cells = append(
				fingerprintBucketValue.Cells,
				fingerprintCell{
					ID:                cell.ID,
					LatitudeIndex:     cell.LatitudeIndex,
					LongitudeIndex:    cell.LongitudeIndex,
					AltitudeBandIndex: cell.AltitudeBandIndex,
					AltitudeKnown:     cell.AltitudeKnown,
					AircraftNodeIDs:   append([]string(nil), cell.AircraftNodeIDs...),
					MeanQualityScore:  cell.MeanQualityScore,
				},
			)
		}
		envelope.Occupancy.Buckets = append(envelope.Occupancy.Buckets, fingerprintBucketValue)
	}
	for _, report := range result.SectorComplexity {
		envelope.SectorComplexity = append(envelope.SectorComplexity, fingerprintSector{
			ID:                     report.ID,
			BucketID:               report.BucketID,
			AircraftNodeIDs:        append([]string(nil), report.AircraftNodeIDs...),
			AircraftCount:          report.AircraftCount,
			AltitudeBandCount:      report.AltitudeBandCount,
			UnknownAltitudeCount:   report.UnknownAltitudeCount,
			CandidatePairCount:     report.CandidatePairCount,
			ConvergingPairCount:    report.ConvergingPairCount,
			ContextualRiskCount:    report.ContextualRiskCount,
			ElevatedRiskCount:      report.ElevatedRiskCount,
			HighRiskCount:          report.HighRiskCount,
			IndeterminateRiskCount: report.IndeterminateRiskCount,
			HeadingDispersion:      report.HeadingDispersion,
			SpeedVariability:       report.SpeedVariability,
			Score:                  report.Score,
			Level:                  string(report.Level),
			Components:             append([]ScoreComponent(nil), report.Components...),
			ConfidenceScore:        report.Confidence.Score,
		})
	}
	encoded, err := json.Marshal(envelope)
	if err != nil {
		panic(err)
	}
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:])
}
