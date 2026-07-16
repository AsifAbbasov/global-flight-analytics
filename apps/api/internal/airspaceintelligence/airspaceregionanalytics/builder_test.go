package airspaceregionanalytics

import (
	"slices"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/interactiongraph"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/localtrafficscene"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/proximityscanner"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/separationrisk"
)

func TestBuildCreatesTemporalOccupancyComplexityAndRegionAnalytics(t *testing.T) {
	start := time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC)
	snapshots := []SnapshotInput{
		testSnapshot(start.Add(10*time.Second), []localtrafficscene.Aircraft{
			testAircraft("A", 40.10, 49.10, float64Pointer(9000), 0, 210, 0.92),
			testAircraft("B", 40.15, 49.15, float64Pointer(12000), 180, 220, 0.90),
		}, []proximityscanner.Candidate{
			{ID: "A--B", SourceNodeID: "A", TargetNodeID: "B", Kind: interactiongraph.InteractionKindConverging},
		}, []separationrisk.Assessment{
			{CandidateID: "A--B", SourceNodeID: "A", TargetNodeID: "B", Level: separationrisk.RiskLevelElevated, Kind: interactiongraph.InteractionKindConverging},
		}, "1"),
		testSnapshot(start.Add(70*time.Second), []localtrafficscene.Aircraft{
			testAircraft("A", 40.10, 49.10, float64Pointer(9000), 0, 210, 0.92),
			testAircraft("B", 40.15, 49.15, float64Pointer(12000), 180, 220, 0.90),
			testAircraft("C", 40.20, 49.20, float64Pointer(15000), 90, 250, 0.88),
		}, []proximityscanner.Candidate{
			{ID: "A--B", SourceNodeID: "A", TargetNodeID: "B", Kind: interactiongraph.InteractionKindConverging},
			{ID: "B--C", SourceNodeID: "B", TargetNodeID: "C", Kind: interactiongraph.InteractionKindNearby},
		}, []separationrisk.Assessment{
			{CandidateID: "A--B", SourceNodeID: "A", TargetNodeID: "B", Level: separationrisk.RiskLevelHigh, Kind: interactiongraph.InteractionKindConverging},
			{CandidateID: "B--C", SourceNodeID: "B", TargetNodeID: "C", Level: separationrisk.RiskLevelContextual, Kind: interactiongraph.InteractionKindNearby},
		}, "2"),
		testSnapshot(start.Add(130*time.Second), []localtrafficscene.Aircraft{
			testAircraft("A", 40.10, 49.10, float64Pointer(9000), 0, 210, 0.92),
			testAircraft("B", 40.15, 49.15, float64Pointer(12000), 180, 220, 0.90),
			testAircraft("C", 40.20, 49.20, float64Pointer(15000), 90, 250, 0.88),
			testAircraft("D", 40.25, 49.25, float64Pointer(18000), 270, 160, 0.86),
		}, []proximityscanner.Candidate{
			{ID: "A--B", SourceNodeID: "A", TargetNodeID: "B", Kind: interactiongraph.InteractionKindConverging},
			{ID: "A--C", SourceNodeID: "A", TargetNodeID: "C", Kind: interactiongraph.InteractionKindNearby},
			{ID: "B--D", SourceNodeID: "B", TargetNodeID: "D", Kind: interactiongraph.InteractionKindConverging},
		}, []separationrisk.Assessment{
			{CandidateID: "A--B", SourceNodeID: "A", TargetNodeID: "B", Level: separationrisk.RiskLevelHigh, Kind: interactiongraph.InteractionKindConverging},
			{CandidateID: "A--C", SourceNodeID: "A", TargetNodeID: "C", Level: separationrisk.RiskLevelElevated, Kind: interactiongraph.InteractionKindNearby},
			{CandidateID: "B--D", SourceNodeID: "B", TargetNodeID: "D", Level: separationrisk.RiskLevelHigh, Kind: interactiongraph.InteractionKindConverging},
		}, "3"),
	}

	result, err := Build(Request{
		RegionCode:  " az ",
		WindowStart: start,
		WindowEnd:   start.Add(3 * time.Minute),
		GeneratedAt: start.Add(4 * time.Minute),
		Snapshots:   snapshots,
	}, DefaultPolicy())
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if result.Status != ResultStatusComplete {
		t.Fatalf("Status = %q, want %q", result.Status, ResultStatusComplete)
	}
	if result.RegionCode != "AZ" {
		t.Fatalf("RegionCode = %q, want AZ", result.RegionCode)
	}
	if result.Occupancy.Metrics.BucketCount != 3 || result.Metrics.BucketCount != 3 {
		t.Fatalf("bucket counts = %+v", result.Occupancy.Metrics)
	}
	if result.Occupancy.Metrics.UniqueAircraftCount != 4 {
		t.Fatalf("UniqueAircraftCount = %d, want 4", result.Occupancy.Metrics.UniqueAircraftCount)
	}
	if result.Metrics.OccupancyTrend != OccupancyTrendRising {
		t.Fatalf("OccupancyTrend = %q, want %q", result.Metrics.OccupancyTrend, OccupancyTrendRising)
	}
	if result.Metrics.HighRiskCount != 3 || result.Metrics.ElevatedRiskCount != 2 {
		t.Fatalf("risk metrics = %+v", result.Metrics)
	}
	if result.Metrics.PeakComplexityScore <= 0 || result.Metrics.AirspacePressureIndex <= 0 {
		t.Fatalf("complexity metrics = %+v", result.Metrics)
	}
	if len(result.SectorComplexity) != 3 {
		t.Fatalf("SectorComplexity count = %d, want 3", len(result.SectorComplexity))
	}
	if result.Confidence.Level != ConfidenceLevelHigh {
		t.Fatalf("Confidence.Level = %q, want %q", result.Confidence.Level, ConfidenceLevelHigh)
	}
	if len(result.Provenance.InputFingerprint) != 64 {
		t.Fatalf("fingerprint length = %d", len(result.Provenance.InputFingerprint))
	}
}

func TestBuildComplexCrossingTrafficScoresAboveParallelTraffic(t *testing.T) {
	start := time.Date(2026, 7, 17, 13, 0, 0, 0, time.UTC)
	parallelAircraft := []localtrafficscene.Aircraft{
		testAircraft("A", 40.1, 49.1, float64Pointer(9000), 90, 220, 0.9),
		testAircraft("B", 40.2, 49.2, float64Pointer(9000), 90, 222, 0.9),
		testAircraft("C", 40.3, 49.3, float64Pointer(12000), 92, 218, 0.9),
		testAircraft("D", 40.4, 49.4, float64Pointer(12000), 88, 221, 0.9),
	}
	complexAircraft := []localtrafficscene.Aircraft{
		testAircraft("A", 40.1, 49.1, float64Pointer(9000), 0, 130, 0.9),
		testAircraft("B", 40.2, 49.2, float64Pointer(12000), 90, 220, 0.9),
		testAircraft("C", 40.3, 49.3, float64Pointer(15000), 180, 300, 0.9),
		testAircraft("D", 40.4, 49.4, float64Pointer(18000), 270, 170, 0.9),
	}
	parallel, err := Build(testRequest(start, testSnapshot(start.Add(10*time.Second), parallelAircraft, nil, nil, "parallel")), DefaultPolicy())
	if err != nil {
		t.Fatalf("parallel Build() error = %v", err)
	}
	candidates := []proximityscanner.Candidate{
		{ID: "A--B", SourceNodeID: "A", TargetNodeID: "B", Kind: interactiongraph.InteractionKindConverging},
		{ID: "A--C", SourceNodeID: "A", TargetNodeID: "C", Kind: interactiongraph.InteractionKindNearby},
		{ID: "B--D", SourceNodeID: "B", TargetNodeID: "D", Kind: interactiongraph.InteractionKindConverging},
		{ID: "C--D", SourceNodeID: "C", TargetNodeID: "D", Kind: interactiongraph.InteractionKindNearby},
	}
	risks := []separationrisk.Assessment{
		{CandidateID: "A--B", SourceNodeID: "A", TargetNodeID: "B", Level: separationrisk.RiskLevelHigh},
		{CandidateID: "A--C", SourceNodeID: "A", TargetNodeID: "C", Level: separationrisk.RiskLevelElevated},
		{CandidateID: "B--D", SourceNodeID: "B", TargetNodeID: "D", Level: separationrisk.RiskLevelHigh},
		{CandidateID: "C--D", SourceNodeID: "C", TargetNodeID: "D", Level: separationrisk.RiskLevelContextual},
	}
	complex, err := Build(testRequest(start, testSnapshot(start.Add(10*time.Second), complexAircraft, candidates, risks, "complex")), DefaultPolicy())
	if err != nil {
		t.Fatalf("complex Build() error = %v", err)
	}
	if complex.Metrics.PeakComplexityScore <= parallel.Metrics.PeakComplexityScore {
		t.Fatalf("complex score %v <= parallel score %v", complex.Metrics.PeakComplexityScore, parallel.Metrics.PeakComplexityScore)
	}
	if complex.SectorComplexity[0].HeadingDispersion <= parallel.SectorComplexity[0].HeadingDispersion {
		t.Fatalf("heading dispersion complex=%v parallel=%v", complex.SectorComplexity[0].HeadingDispersion, parallel.SectorComplexity[0].HeadingDispersion)
	}
}

func TestBuildUnknownAltitudeUsesExplicitBandAndLimitsResult(t *testing.T) {
	start := time.Date(2026, 7, 17, 14, 0, 0, 0, time.UTC)
	snapshot := testSnapshot(start.Add(10*time.Second), []localtrafficscene.Aircraft{
		testAircraft("A", 40.1, 49.1, nil, 45, 200, 0.8),
	}, nil, nil, "unknown")
	result, err := Build(testRequest(start, snapshot), DefaultPolicy())
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if result.Status != ResultStatusLimited {
		t.Fatalf("Status = %q, want %q", result.Status, ResultStatusLimited)
	}
	cell := result.Occupancy.Buckets[0].Cells[0]
	if cell.AltitudeKnown || cell.AltitudeBandIndex != -1 {
		t.Fatalf("unknown altitude cell = %+v", cell)
	}
	if result.Metrics.UnknownAltitudeCount != 1 || result.SectorComplexity[0].UnknownAltitudeCount != 1 {
		t.Fatalf("unknown altitude metrics = %+v", result.Metrics)
	}
	if !hasLimitation(result.Limitations, "unknown_altitude_occupancy_present") {
		t.Fatalf("limitations = %+v", result.Limitations)
	}
}

func TestBuildFingerprintIsDeterministicAcrossInputOrder(t *testing.T) {
	start := time.Date(2026, 7, 17, 15, 0, 0, 0, time.UTC)
	first := testSnapshot(start.Add(10*time.Second), []localtrafficscene.Aircraft{
		testAircraft("B", 40.2, 49.2, float64Pointer(12000), 180, 220, 0.9),
		testAircraft("A", 40.1, 49.1, float64Pointer(9000), 0, 210, 0.9),
	}, nil, nil, "first")
	second := testSnapshot(start.Add(70*time.Second), []localtrafficscene.Aircraft{
		testAircraft("D", 40.4, 49.4, float64Pointer(18000), 270, 160, 0.9),
		testAircraft("C", 40.3, 49.3, float64Pointer(15000), 90, 250, 0.9),
	}, nil, nil, "second")
	request := Request{
		RegionCode:  "AZ",
		WindowStart: start,
		WindowEnd:   start.Add(2 * time.Minute),
		GeneratedAt: start.Add(3 * time.Minute),
		Snapshots:   []SnapshotInput{first, second},
	}
	left, err := Build(request, DefaultPolicy())
	if err != nil {
		t.Fatalf("left Build() error = %v", err)
	}
	reversed := request
	reversed.Snapshots = []SnapshotInput{second, first}
	slices.Reverse(reversed.Snapshots[0].Scene.Aircraft)
	slices.Reverse(reversed.Snapshots[1].Scene.Aircraft)
	right, err := Build(reversed, DefaultPolicy())
	if err != nil {
		t.Fatalf("right Build() error = %v", err)
	}
	if left.Provenance.InputFingerprint != right.Provenance.InputFingerprint {
		t.Fatalf("fingerprints differ: %s != %s", left.Provenance.InputFingerprint, right.Provenance.InputFingerprint)
	}
}

func testRequest(start time.Time, snapshot SnapshotInput) Request {
	return Request{
		RegionCode:  "AZ",
		WindowStart: start,
		WindowEnd:   start.Add(time.Minute),
		GeneratedAt: start.Add(2 * time.Minute),
		Snapshots:   []SnapshotInput{snapshot},
	}
}

func testSnapshot(
	asOf time.Time,
	aircraft []localtrafficscene.Aircraft,
	candidates []proximityscanner.Candidate,
	assessments []separationrisk.Assessment,
	suffix string,
) SnapshotInput {
	for index := range aircraft {
		aircraft[index].ObservedAt = asOf.Add(-5 * time.Second)
	}
	sceneFingerprint := "scene-" + suffix
	scanFingerprint := "scan-" + suffix
	riskFingerprint := "risk-" + suffix
	contextualCount := 0
	elevatedCount := 0
	highCount := 0
	indeterminateCount := 0
	for _, assessment := range assessments {
		switch assessment.Level {
		case separationrisk.RiskLevelContextual:
			contextualCount++
		case separationrisk.RiskLevelElevated:
			elevatedCount++
		case separationrisk.RiskLevelHigh:
			highCount++
		case separationrisk.RiskLevelIndeterminate:
			indeterminateCount++
		}
	}
	return SnapshotInput{
		Scene: localtrafficscene.Result{
			Status:     localtrafficscene.ResultStatusComplete,
			RegionCode: "AZ",
			AsOfTime:   asOf,
			Aircraft:   aircraft,
			Confidence: localtrafficscene.Confidence{Score: 0.9},
			ScopeGuard: localtrafficscene.ScopeGuardResearchOnly,
			Provenance: localtrafficscene.Provenance{
				InputFingerprint: sceneFingerprint,
				SourceNames:      []string{"airplanes.live"},
				LatestObservedAt: asOf.Add(-5 * time.Second),
			},
			GeneratedAt: asOf.Add(time.Second),
		},
		Scan: proximityscanner.Result{
			Status:      proximityscanner.ResultStatusComplete,
			RegionCode:  "AZ",
			SceneStatus: localtrafficscene.ResultStatusComplete,
			AsOfTime:    asOf,
			Candidates:  candidates,
			Confidence:  proximityscanner.Confidence{Score: 0.88},
			ScopeGuard:  proximityscanner.ScopeGuardResearchOnly,
			Provenance: proximityscanner.Provenance{
				InputFingerprint: scanFingerprint,
				SceneFingerprint: sceneFingerprint,
				SourceNames:      []string{"airplanes.live"},
				LatestObservedAt: asOf.Add(-5 * time.Second),
			},
			GeneratedAt: asOf.Add(2 * time.Second),
		},
		Risk: separationrisk.Result{
			Status:      separationrisk.ResultStatusComplete,
			RegionCode:  "AZ",
			AsOfTime:    asOf,
			Assessments: assessments,
			Metrics: separationrisk.Metrics{
				ContextualCount:    contextualCount,
				ElevatedCount:      elevatedCount,
				HighCount:          highCount,
				IndeterminateCount: indeterminateCount,
			},
			Confidence: separationrisk.Confidence{Score: 0.86},
			ScopeGuard: separationrisk.ScopeGuardResearchOnly,
			Provenance: separationrisk.Provenance{
				InputFingerprint: riskFingerprint,
				ScanFingerprint:  scanFingerprint,
				SourceNames:      []string{"airplanes.live"},
				LatestObservedAt: asOf.Add(-5 * time.Second),
			},
			GeneratedAt: asOf.Add(3 * time.Second),
		},
	}
}

func testAircraft(
	nodeID string,
	latitude float64,
	longitude float64,
	altitude *float64,
	heading float64,
	speed float64,
	quality float64,
) localtrafficscene.Aircraft {
	reference := interactiongraph.AltitudeReferenceUnknown
	if altitude != nil {
		reference = interactiongraph.AltitudeReferenceBarometric
	}
	return localtrafficscene.Aircraft{
		NodeID:                  nodeID,
		ICAO24:                  nodeID,
		Latitude:                latitude,
		Longitude:               longitude,
		AltitudeMeters:          altitude,
		AltitudeReference:       reference,
		VelocityMetersPerSecond: speed,
		HeadingDegrees:          heading,
		SourceName:              "airplanes.live",
		QualityScore:            quality,
	}
}

func float64Pointer(value float64) *float64 { return &value }

func hasLimitation(limitations []Limitation, code string) bool {
	for _, limitation := range limitations {
		if limitation.Code == code {
			return true
		}
	}
	return false
}
