package airspaceregionanalytics

import (
	"math"
	"sort"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/interactiongraph"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/separationrisk"
)

type sectorAccumulator struct {
	placements             []placement
	candidatePairCount     int
	convergingPairCount    int
	contextualRiskCount    int
	elevatedRiskCount      int
	highRiskCount          int
	indeterminateRiskCount int
}

type upstreamConfidence struct {
	scene float64
	scan  float64
	risk  float64
}

func buildSectorComplexity(
	context occupancyContext,
	policy Policy,
) []SectorComplexityReport {
	accumulators := make(map[string]*sectorAccumulator, len(context.placementsBySector))
	for sectorID, placements := range context.placementsBySector {
		copied := append([]placement(nil), placements...)
		sort.Slice(copied, func(left int, right int) bool {
			return copied[left].aircraft.NodeID < copied[right].aircraft.NodeID
		})
		accumulators[sectorID] = &sectorAccumulator{placements: copied}
	}

	for bucketID, snapshots := range context.snapshotsByBucket {
		for _, snapshot := range snapshots {
			for _, candidate := range snapshot.Scan.Candidates {
				for _, sectorID := range affectedSectors(
					bucketID,
					candidate.SourceNodeID,
					candidate.TargetNodeID,
					context.nodeSectorByBucket,
				) {
					accumulator := accumulators[sectorID]
					if accumulator == nil {
						continue
					}
					accumulator.candidatePairCount++
					if candidate.Kind == interactiongraph.InteractionKindConverging {
						accumulator.convergingPairCount++
					}
				}
			}
			for _, assessment := range snapshot.Risk.Assessments {
				for _, sectorID := range affectedSectors(
					bucketID,
					assessment.SourceNodeID,
					assessment.TargetNodeID,
					context.nodeSectorByBucket,
				) {
					accumulator := accumulators[sectorID]
					if accumulator == nil {
						continue
					}
					switch assessment.Level {
					case separationrisk.RiskLevelContextual:
						accumulator.contextualRiskCount++
					case separationrisk.RiskLevelElevated:
						accumulator.elevatedRiskCount++
					case separationrisk.RiskLevelHigh:
						accumulator.highRiskCount++
					case separationrisk.RiskLevelIndeterminate:
						accumulator.indeterminateRiskCount++
					}
				}
			}
		}
	}

	sectorIDs := make([]string, 0, len(accumulators))
	for sectorID := range accumulators {
		sectorIDs = append(sectorIDs, sectorID)
	}
	sort.Strings(sectorIDs)

	reports := make([]SectorComplexityReport, 0, len(sectorIDs))
	for _, sectorID := range sectorIDs {
		accumulator := accumulators[sectorID]
		first := accumulator.placements[0]
		nodeIDs := make([]string, 0, len(accumulator.placements))
		qualityTotal := 0.0
		unknownAltitudeCount := 0
		altitudeBands := make(map[int]struct{})
		headings := make([]float64, 0, len(accumulator.placements))
		speeds := make([]float64, 0, len(accumulator.placements))
		for _, item := range accumulator.placements {
			nodeIDs = append(nodeIDs, item.aircraft.NodeID)
			qualityTotal += item.aircraft.QualityScore
			headings = append(headings, item.aircraft.HeadingDegrees)
			speeds = append(speeds, item.aircraft.VelocityMetersPerSecond)
			if item.altitudeKnown {
				altitudeBands[item.altitudeBandIndex] = struct{}{}
			} else {
				unknownAltitudeCount++
			}
		}

		aircraftCount := len(accumulator.placements)
		meanQuality := qualityTotal / float64(aircraftCount)
		possiblePairs := aircraftCount * (aircraftCount - 1) / 2
		densityScore := clampUnit(float64(aircraftCount) / float64(policy.DenseAircraftCount))
		pairInteractionScore := 0.0
		if possiblePairs > 0 {
			pairInteractionScore = clampUnit(
				float64(accumulator.candidatePairCount) / float64(possiblePairs),
			)
		}
		determinateRiskScore := riskIntensity(accumulator)
		headingScore := headingDispersion(headings)
		speedScore := clampUnit(
			standardDeviation(speeds) / policy.SpeedVariabilityScaleMetersPerSecond,
		)
		altitudeMixingScore := 0.0
		if len(altitudeBands) > 1 {
			altitudeMixingScore = clampUnit(
				float64(len(altitudeBands)-1) / float64(policy.MixedAltitudeBandCount-1),
			)
		}
		components := []ScoreComponent{
			{Name: "aircraft_density", Score: densityScore, Weight: policy.ComplexityWeights.Density},
			{Name: "pair_interaction", Score: pairInteractionScore, Weight: policy.ComplexityWeights.PairInteraction},
			{Name: "determinate_risk", Score: determinateRiskScore, Weight: policy.ComplexityWeights.DeterminateRisk},
			{Name: "heading_dispersion", Score: headingScore, Weight: policy.ComplexityWeights.HeadingDispersion},
			{Name: "speed_variability", Score: speedScore, Weight: policy.ComplexityWeights.SpeedVariability},
			{Name: "altitude_mixing", Score: altitudeMixingScore, Weight: policy.ComplexityWeights.AltitudeMixing},
		}
		score := weightedScore(components)
		upstream := meanUpstreamConfidence(context.snapshotsByBucket[first.bucketID])
		verticalCompleteness := clampUnit(
			float64(aircraftCount-unknownAltitudeCount) / float64(aircraftCount),
		)
		confidence := buildConfidence(
			upstream,
			meanQuality,
			verticalCompleteness,
			policy,
		)
		report := SectorComplexityReport{
			ID:                     sectorID,
			BucketID:               first.bucketID,
			BucketStart:            first.bucketStart,
			BucketEnd:              first.bucketEnd,
			LatitudeIndex:          first.latitudeIndex,
			LongitudeIndex:         first.longitudeIndex,
			AircraftNodeIDs:        nodeIDs,
			AircraftCount:          aircraftCount,
			AltitudeBandCount:      len(altitudeBands),
			UnknownAltitudeCount:   unknownAltitudeCount,
			CandidatePairCount:     accumulator.candidatePairCount,
			ConvergingPairCount:    accumulator.convergingPairCount,
			ContextualRiskCount:    accumulator.contextualRiskCount,
			ElevatedRiskCount:      accumulator.elevatedRiskCount,
			HighRiskCount:          accumulator.highRiskCount,
			IndeterminateRiskCount: accumulator.indeterminateRiskCount,
			HeadingDispersion:      headingScore,
			SpeedVariability:       speedScore,
			Score:                  score,
			Level:                  complexityLevelForScore(score, policy),
			Components:             components,
			Confidence:             confidence,
			Limitations: sectorLimitations(
				aircraftCount,
				unknownAltitudeCount,
				accumulator,
			),
			Explanations: sectorExplanations(components),
		}
		reports = append(reports, report)
	}
	return reports
}

func affectedSectors(
	bucketID string,
	sourceNodeID string,
	targetNodeID string,
	nodeSectorByBucket map[string]map[string]string,
) []string {
	sectorSet := make(map[string]struct{}, 2)
	if sourceSector := nodeSectorByBucket[bucketID][sourceNodeID]; sourceSector != "" {
		sectorSet[sourceSector] = struct{}{}
	}
	if targetSector := nodeSectorByBucket[bucketID][targetNodeID]; targetSector != "" {
		sectorSet[targetSector] = struct{}{}
	}
	sectors := make([]string, 0, len(sectorSet))
	for sectorID := range sectorSet {
		sectors = append(sectors, sectorID)
	}
	sort.Strings(sectors)
	return sectors
}

func riskIntensity(accumulator *sectorAccumulator) float64 {
	total := accumulator.contextualRiskCount +
		accumulator.elevatedRiskCount +
		accumulator.highRiskCount
	if total == 0 {
		return 0
	}
	weighted := float64(accumulator.contextualRiskCount)*0.15 +
		float64(accumulator.elevatedRiskCount)*0.65 +
		float64(accumulator.highRiskCount)
	return clampUnit(weighted / float64(total))
}

func headingDispersion(headings []float64) float64 {
	if len(headings) <= 1 {
		return 0
	}
	sineTotal := 0.0
	cosineTotal := 0.0
	for _, heading := range headings {
		radians := heading * math.Pi / 180
		sineTotal += math.Sin(radians)
		cosineTotal += math.Cos(radians)
	}
	resultantLength := math.Hypot(sineTotal, cosineTotal) / float64(len(headings))
	return clampUnit(1 - resultantLength)
}

func standardDeviation(values []float64) float64 {
	if len(values) <= 1 {
		return 0
	}
	mean := 0.0
	for _, value := range values {
		mean += value
	}
	mean /= float64(len(values))
	variance := 0.0
	for _, value := range values {
		difference := value - mean
		variance += difference * difference
	}
	variance /= float64(len(values))
	return math.Sqrt(variance)
}

func meanUpstreamConfidence(snapshots []SnapshotInput) upstreamConfidence {
	if len(snapshots) == 0 {
		return upstreamConfidence{}
	}
	result := upstreamConfidence{}
	for _, snapshot := range snapshots {
		result.scene += snapshot.Scene.Confidence.Score
		result.scan += snapshot.Scan.Confidence.Score
		result.risk += snapshot.Risk.Confidence.Score
	}
	count := float64(len(snapshots))
	result.scene /= count
	result.scan /= count
	result.risk /= count
	return result
}

func buildConfidence(
	upstream upstreamConfidence,
	dataQuality float64,
	coverage float64,
	policy Policy,
) Confidence {
	components := []ScoreComponent{
		{Name: "scene_confidence", Score: upstream.scene, Weight: policy.ConfidenceWeights.SceneConfidence},
		{Name: "scan_confidence", Score: upstream.scan, Weight: policy.ConfidenceWeights.ScanConfidence},
		{Name: "risk_confidence", Score: upstream.risk, Weight: policy.ConfidenceWeights.RiskConfidence},
		{Name: "data_quality", Score: dataQuality, Weight: policy.ConfidenceWeights.DataQuality},
		{Name: "evidence_coverage", Score: coverage, Weight: policy.ConfidenceWeights.TemporalCoverage},
	}
	score := weightedScore(components)
	return Confidence{
		Score:      score,
		Level:      confidenceLevelForScore(score, policy),
		Components: components,
		Reasons: []ConfidenceReason{
			{Code: "upstream_scene_confidence", Message: "Local traffic scene confidence contributes to the analytical confidence.", Contribution: upstream.scene},
			{Code: "upstream_scan_confidence", Message: "Pairwise proximity scan confidence contributes to the analytical confidence.", Contribution: upstream.scan},
			{Code: "upstream_risk_confidence", Message: "Separation risk evidence confidence contributes to the analytical confidence.", Contribution: upstream.risk},
			{Code: "prepared_data_quality", Message: "Mean prepared aircraft quality contributes to the analytical confidence.", Contribution: dataQuality},
			{Code: "evidence_coverage", Message: "Temporal or vertical evidence coverage contributes to the analytical confidence.", Contribution: coverage},
		},
	}
}

func sectorLimitations(
	aircraftCount int,
	unknownAltitudeCount int,
	accumulator *sectorAccumulator,
) []Limitation {
	limitations := []Limitation{
		{
			Code:    "synthetic_research_sector",
			Message: "The sector is a generated analytical grid cell and is not an official air traffic control sector.",
			Scope:   "sector_definition",
		},
		{
			Code:    "research_only_not_operational_complexity",
			Message: "The complexity score is research context and must not be interpreted as controller workload or operational safety logic.",
			Scope:   "operational_use",
		},
	}
	if unknownAltitudeCount > 0 {
		limitations = append(limitations, Limitation{
			Code:    "unknown_altitude_evidence_present",
			Message: "One or more aircraft lack comparable altitude evidence, reducing vertical-layer confidence.",
			Scope:   "vertical_evidence",
		})
	}
	if aircraftCount > 1 && accumulator.candidatePairCount == 0 {
		limitations = append(limitations, Limitation{
			Code:    "pairwise_candidates_unavailable",
			Message: "Multiple aircraft are present but no pairwise candidate evidence was published for this sector.",
			Scope:   "interaction_evidence",
		})
	}
	if accumulator.indeterminateRiskCount > 0 {
		limitations = append(limitations, Limitation{
			Code:    "indeterminate_risk_evidence_present",
			Message: "One or more pair assessments are indeterminate and do not support a determinate risk conclusion.",
			Scope:   "risk_evidence",
		})
	}
	return limitations
}

func sectorExplanations(components []ScoreComponent) []Explanation {
	return []Explanation{
		{
			Code:    "multidimensional_sector_complexity",
			Message: "Sector complexity combines aircraft density, pair interactions, determinate risk evidence, heading dispersion, speed variability, and altitude mixing.",
		},
		{
			Code:    "density_is_not_complexity",
			Message: "Aircraft count is only one component; a smaller crossing or converging traffic scene may be more complex than a larger parallel scene.",
		},
		{
			Code:    dominantComponentCode(components),
			Message: "The dominant weighted component is published so the complexity score remains explainable.",
		},
	}
}

func dominantComponentCode(components []ScoreComponent) string {
	code := "dominant_component_unavailable"
	contribution := -1.0
	for _, component := range components {
		value := component.Score * component.Weight
		if value > contribution {
			contribution = value
			code = "dominant_component_" + component.Name
		}
	}
	return code
}

func complexityLevelForScore(score float64, policy Policy) ComplexityLevel {
	switch {
	case score <= 0:
		return ComplexityLevelNone
	case score < policy.ModerateComplexityMinimumScore:
		return ComplexityLevelLow
	case score < policy.HighComplexityMinimumScore:
		return ComplexityLevelModerate
	case score < policy.SevereComplexityMinimumScore:
		return ComplexityLevelHigh
	default:
		return ComplexityLevelSevere
	}
}

func confidenceLevelForScore(score float64, policy Policy) ConfidenceLevel {
	switch {
	case score <= 0:
		return ConfidenceLevelNone
	case score < policy.MediumConfidenceMinimumScore:
		return ConfidenceLevelLow
	case score < policy.HighConfidenceMinimumScore:
		return ConfidenceLevelMedium
	default:
		return ConfidenceLevelHigh
	}
}

func weightedScore(components []ScoreComponent) float64 {
	total := 0.0
	for _, component := range components {
		total += component.Score * component.Weight
	}
	return clampUnit(total)
}
