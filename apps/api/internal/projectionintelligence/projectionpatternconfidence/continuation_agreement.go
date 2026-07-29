package projectionpatternconfidence

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
)

const earthRadiusM = 6_371_008.8

var (
	ErrContinuationCandidatesInvalid = errors.New(
		"historical continuation candidates are invalid",
	)
	ErrContinuationEvidenceUnavailable = errors.New(
		"historical continuation agreement evidence is unavailable",
	)
)

type continuationVector struct {
	trajectoryID string
	sampleIndex  int
	elapsedS     float64

	anchorLatitude    float64
	anchorLongitude   float64
	endpointLatitude  float64
	endpointLongitude float64

	eastM  float64
	northM float64
}

type continuationAgreementEvidence struct {
	known bool

	sampleCount     int
	pairCount       int
	comparisonCount int
	horizonSeconds  float64

	meanSpreadM          float64
	maximumSpreadM       float64
	meanDivergenceMPS    float64
	maximumDivergenceMPS float64
	score                float64

	vectors []continuationVector
}

func extractContinuationAgreement(
	selection projectionneighbors.Result,
	candidates []trajectory.FlightTrajectory,
	config Config,
) (continuationAgreementEvidence, error) {
	if len(selection.Neighbors) < 2 {
		return continuationAgreementEvidence{}, fmt.Errorf(
			"%w: at least two selected neighbors are required",
			ErrContinuationEvidenceUnavailable,
		)
	}
	if selection.RequiredContinuationDuration <= 0 {
		return continuationAgreementEvidence{}, fmt.Errorf(
			"%w: continuation duration must be positive",
			ErrContinuationEvidenceUnavailable,
		)
	}

	candidateByID, err := indexSelectedCandidates(selection, candidates)
	if err != nil {
		return continuationAgreementEvidence{}, err
	}

	vectorsByTrajectory := make(
		map[string][]continuationVector,
		len(selection.Neighbors),
	)
	allVectors := make(
		[]continuationVector,
		0,
		len(selection.Neighbors)*config.ContinuationAgreementSampleCount,
	)

	for _, neighbor := range selection.Neighbors {
		trajectoryID := strings.TrimSpace(neighbor.TrajectoryID)
		candidate := candidateByID[trajectoryID]
		vectors, vectorErr := continuationVectors(
			candidate,
			neighbor,
			selection.RequiredContinuationDuration,
			config.ContinuationAgreementSampleCount,
		)
		if vectorErr != nil {
			return continuationAgreementEvidence{}, fmt.Errorf(
				"%w: trajectory=%q: %v",
				ErrContinuationEvidenceUnavailable,
				trajectoryID,
				vectorErr,
			)
		}
		vectorsByTrajectory[trajectoryID] = vectors
		allVectors = append(allVectors, vectors...)
	}

	trajectoryIDs := make([]string, 0, len(vectorsByTrajectory))
	for trajectoryID := range vectorsByTrajectory {
		trajectoryIDs = append(trajectoryIDs, trajectoryID)
	}
	sort.Strings(trajectoryIDs)
	sort.SliceStable(allVectors, func(left int, right int) bool {
		if allVectors[left].trajectoryID != allVectors[right].trajectoryID {
			return allVectors[left].trajectoryID < allVectors[right].trajectoryID
		}
		return allVectors[left].sampleIndex < allVectors[right].sampleIndex
	})

	pairCount := len(trajectoryIDs) * (len(trajectoryIDs) - 1) / 2
	comparisonCount := pairCount * config.ContinuationAgreementSampleCount
	if pairCount == 0 || comparisonCount == 0 {
		return continuationAgreementEvidence{}, fmt.Errorf(
			"%w: no pairwise continuation comparisons were produced",
			ErrContinuationEvidenceUnavailable,
		)
	}

	totalSpread := 0.0
	totalDivergence := 0.0
	maximumSpread := 0.0
	maximumDivergence := 0.0

	for left := 0; left < len(trajectoryIDs); left++ {
		for right := left + 1; right < len(trajectoryIDs); right++ {
			leftVectors := vectorsByTrajectory[trajectoryIDs[left]]
			rightVectors := vectorsByTrajectory[trajectoryIDs[right]]
			if len(leftVectors) != config.ContinuationAgreementSampleCount ||
				len(rightVectors) != config.ContinuationAgreementSampleCount {
				return continuationAgreementEvidence{}, fmt.Errorf(
					"%w: incomplete sample vectors",
					ErrContinuationEvidenceUnavailable,
				)
			}
			for sampleIndex := 0; sampleIndex < config.ContinuationAgreementSampleCount; sampleIndex++ {
				leftVector := leftVectors[sampleIndex]
				rightVector := rightVectors[sampleIndex]
				if math.Abs(leftVector.elapsedS-rightVector.elapsedS) > scoreComparisonTolerance ||
					leftVector.elapsedS <= 0 {
					return continuationAgreementEvidence{}, fmt.Errorf(
						"%w: sample horizons are inconsistent",
						ErrContinuationEvidenceUnavailable,
					)
				}
				spread := math.Hypot(
					leftVector.eastM-rightVector.eastM,
					leftVector.northM-rightVector.northM,
				)
				divergence := spread / leftVector.elapsedS
				totalSpread += spread
				totalDivergence += divergence
				if spread > maximumSpread {
					maximumSpread = spread
				}
				if divergence > maximumDivergence {
					maximumDivergence = divergence
				}
			}
		}
	}

	meanSpread := totalSpread / float64(comparisonCount)
	meanDivergence := totalDivergence / float64(comparisonCount)

	return continuationAgreementEvidence{
		known:                true,
		sampleCount:          config.ContinuationAgreementSampleCount,
		pairCount:            pairCount,
		comparisonCount:      comparisonCount,
		horizonSeconds:       selection.RequiredContinuationDuration.Seconds(),
		meanSpreadM:          meanSpread,
		maximumSpreadM:       maximumSpread,
		meanDivergenceMPS:    meanDivergence,
		maximumDivergenceMPS: maximumDivergence,
		score: clampUnit(
			1 - meanDivergence/config.ContinuationDivergenceNormalizationMPS,
		),
		vectors: allVectors,
	}, nil
}

func indexSelectedCandidates(
	selection projectionneighbors.Result,
	candidates []trajectory.FlightTrajectory,
) (map[string]trajectory.FlightTrajectory, error) {
	selected := make(map[string]struct{}, len(selection.Neighbors))
	for _, neighbor := range selection.Neighbors {
		selected[strings.TrimSpace(neighbor.TrajectoryID)] = struct{}{}
	}

	result := make(map[string]trajectory.FlightTrajectory, len(selected))
	for _, candidate := range candidates {
		trajectoryID := strings.TrimSpace(candidate.ID)
		if _, required := selected[trajectoryID]; !required {
			continue
		}
		if trajectoryID == "" || candidate.ID != trajectoryID {
			return nil, fmt.Errorf(
				"%w: selected candidate identifier is not normalized",
				ErrContinuationCandidatesInvalid,
			)
		}
		if _, exists := result[trajectoryID]; exists {
			return nil, fmt.Errorf(
				"%w: duplicate selected candidate %q",
				ErrContinuationCandidatesInvalid,
				trajectoryID,
			)
		}
		result[trajectoryID] = candidate
	}

	for trajectoryID := range selected {
		if _, exists := result[trajectoryID]; !exists {
			return nil, fmt.Errorf(
				"%w: selected candidate %q is missing",
				ErrContinuationCandidatesInvalid,
				trajectoryID,
			)
		}
	}
	return result, nil
}

func continuationVectors(
	candidate trajectory.FlightTrajectory,
	neighbor projectionneighbors.Neighbor,
	horizon time.Duration,
	sampleCount int,
) ([]continuationVector, error) {
	if neighbor.AnchorPointIndex < 0 || neighbor.AnchorPointIndex >= len(candidate.Points) {
		return nil, fmt.Errorf("anchor point index is outside candidate points")
	}
	anchor := candidate.Points[neighbor.AnchorPointIndex]
	if !anchor.ObservedAt.Equal(neighbor.AnchorObservedAt) {
		return nil, fmt.Errorf("anchor timestamp does not match selected-neighbor evidence")
	}
	if neighbor.ContinuationEndTime.Before(anchor.ObservedAt.Add(horizon)) {
		return nil, fmt.Errorf("selected continuation does not cover the required horizon")
	}
	if err := validateCoordinate(anchor.Latitude, anchor.Longitude); err != nil {
		return nil, fmt.Errorf("anchor coordinate: %w", err)
	}

	vectors := make([]continuationVector, 0, sampleCount)
	for sample := 1; sample <= sampleCount; sample++ {
		offset := time.Duration(
			int64(horizon) * int64(sample) / int64(sampleCount),
		)
		if offset <= 0 {
			return nil, fmt.Errorf("sample duration is not positive")
		}
		target := anchor.ObservedAt.Add(offset)
		latitude, longitude, err := interpolatePosition(
			candidate.Points,
			neighbor.AnchorPointIndex,
			target,
		)
		if err != nil {
			return nil, err
		}
		eastM, northM := localDisplacementM(
			anchor.Latitude,
			anchor.Longitude,
			latitude,
			longitude,
		)
		vectors = append(vectors, continuationVector{
			trajectoryID:      strings.TrimSpace(neighbor.TrajectoryID),
			sampleIndex:       sample,
			elapsedS:          offset.Seconds(),
			anchorLatitude:    anchor.Latitude,
			anchorLongitude:   anchor.Longitude,
			endpointLatitude:  latitude,
			endpointLongitude: longitude,
			eastM:             eastM,
			northM:            northM,
		})
	}
	return vectors, nil
}

func interpolatePosition(
	points []trajectory.TrackPoint4D,
	anchorIndex int,
	target time.Time,
) (float64, float64, error) {
	if anchorIndex < 0 || anchorIndex >= len(points) {
		return 0, 0, fmt.Errorf("anchor index is invalid")
	}
	if target.Before(points[anchorIndex].ObservedAt) {
		return 0, 0, fmt.Errorf("sample target precedes anchor")
	}

	lower := points[anchorIndex]
	if target.Equal(lower.ObservedAt) {
		return lower.Latitude, lower.Longitude, nil
	}
	for index := anchorIndex + 1; index < len(points); index++ {
		upper := points[index]
		if !upper.ObservedAt.After(lower.ObservedAt) {
			return 0, 0, fmt.Errorf("candidate continuation timestamps are not strictly increasing")
		}
		if target.Equal(upper.ObservedAt) {
			if err := validateCoordinate(upper.Latitude, upper.Longitude); err != nil {
				return 0, 0, err
			}
			return upper.Latitude, upper.Longitude, nil
		}
		if target.Before(upper.ObservedAt) {
			if err := validateCoordinate(lower.Latitude, lower.Longitude); err != nil {
				return 0, 0, err
			}
			if err := validateCoordinate(upper.Latitude, upper.Longitude); err != nil {
				return 0, 0, err
			}
			span := upper.ObservedAt.Sub(lower.ObservedAt)
			fraction := float64(target.Sub(lower.ObservedAt)) / float64(span)
			latitude := lower.Latitude + (upper.Latitude-lower.Latitude)*fraction
			longitudeDelta := normalizeLongitudeDegrees(upper.Longitude - lower.Longitude)
			longitude := normalizeLongitudeDegrees(lower.Longitude + longitudeDelta*fraction)
			if err := validateCoordinate(latitude, longitude); err != nil {
				return 0, 0, err
			}
			return latitude, longitude, nil
		}
		lower = upper
	}
	return 0, 0, fmt.Errorf("candidate continuation does not reach sample target")
}

func validateCoordinate(latitude float64, longitude float64) error {
	if !finite(latitude) || !finite(longitude) ||
		latitude < -90 || latitude > 90 ||
		longitude < -180 || longitude > 180 {
		return fmt.Errorf("coordinate is invalid")
	}
	return nil
}

func localDisplacementM(
	startLatitude float64,
	startLongitude float64,
	endLatitude float64,
	endLongitude float64,
) (float64, float64) {
	startLatitudeRadians := startLatitude * math.Pi / 180
	endLatitudeRadians := endLatitude * math.Pi / 180
	latitudeDelta := endLatitudeRadians - startLatitudeRadians
	longitudeDelta := normalizeLongitudeRadians(
		(endLongitude - startLongitude) * math.Pi / 180,
	)
	meanLatitude := (startLatitudeRadians + endLatitudeRadians) / 2
	return earthRadiusM * longitudeDelta * math.Cos(meanLatitude),
		earthRadiusM * latitudeDelta
}

func normalizeLongitudeDegrees(value float64) float64 {
	for value > 180 {
		value -= 360
	}
	for value < -180 {
		value += 360
	}
	return value
}

func normalizeLongitudeRadians(value float64) float64 {
	for value > math.Pi {
		value -= 2 * math.Pi
	}
	for value < -math.Pi {
		value += 2 * math.Pi
	}
	return value
}
