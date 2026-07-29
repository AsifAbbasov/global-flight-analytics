package projectionneighbors

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

type anchorSearchFailure string

const (
	anchorSearchFailureNone          anchorSearchFailure = ""
	anchorSearchFailureUnavailable   anchorSearchFailure = "continuation_unavailable"
	anchorSearchFailureDiscontinuous anchorSearchFailure = "continuation_discontinuous"
)

type AnchorEvidence struct {
	PointIndex int
	ObservedAt time.Time
	DistanceKM float64

	ContinuationPointCount int
	ContinuationEndTime    time.Time
	MaximumObservedGap     time.Duration
}

type anchorSearchResult struct {
	Evidence AnchorEvidence
	Failure  anchorSearchFailure
}

func (result anchorSearchResult) Found() bool {
	return result.Failure == anchorSearchFailureNone &&
		result.Evidence.PointIndex >= 0
}

func findAnchor(
	currentEndpoint trajectory.TrackPoint4D,
	candidatePoints []trajectory.TrackPoint4D,
	minimumPrefixPointCount int,
	requiredContinuationDuration time.Duration,
	maximumContinuationGap time.Duration,
) anchorSearchResult {
	if len(candidatePoints) < minimumPrefixPointCount+1 ||
		requiredContinuationDuration <= 0 ||
		maximumContinuationGap <= 0 {
		return anchorSearchResult{
			Failure: anchorSearchFailureUnavailable,
		}
	}

	firstAnchorIndex := minimumPrefixPointCount - 1
	lastPointTime := candidatePoints[len(candidatePoints)-1].ObservedAt.UTC()
	durationAvailable := false
	for index := firstAnchorIndex; index < len(candidatePoints)-1; index++ {
		requiredEndTime := candidatePoints[index].ObservedAt.UTC().Add(
			requiredContinuationDuration,
		)
		if !lastPointTime.Before(requiredEndTime) {
			durationAvailable = true
			break
		}
	}

	currentGeoPoint := geoPoint{
		latitude:  currentEndpoint.Latitude,
		longitude: currentEndpoint.Longitude,
	}

	bestIndex := -1
	bestEndIndex := -1
	bestDistance := 0.0

	for segmentStart := 0; segmentStart < len(candidatePoints); {
		segmentEnd := segmentStart
		for segmentEnd+1 < len(candidatePoints) {
			gap := candidatePoints[segmentEnd+1].ObservedAt.UTC().Sub(
				candidatePoints[segmentEnd].ObservedAt.UTC(),
			)
			if gap < 0 || gap > maximumContinuationGap {
				break
			}
			segmentEnd++
		}

		anchorStart := segmentStart
		if anchorStart < firstAnchorIndex {
			anchorStart = firstAnchorIndex
		}
		if anchorStart < segmentEnd {
			continuationEndIndex := anchorStart + 1
			for anchorIndex := anchorStart; anchorIndex < segmentEnd; anchorIndex++ {
				if continuationEndIndex < anchorIndex+1 {
					continuationEndIndex = anchorIndex + 1
				}
				requiredEndTime := candidatePoints[anchorIndex].ObservedAt.UTC().Add(
					requiredContinuationDuration,
				)
				for continuationEndIndex <= segmentEnd &&
					candidatePoints[continuationEndIndex].ObservedAt.UTC().Before(requiredEndTime) {
					continuationEndIndex++
				}
				if continuationEndIndex > segmentEnd {
					break
				}

				distance := haversineKM(
					currentGeoPoint,
					geoPoint{
						latitude:  candidatePoints[anchorIndex].Latitude,
						longitude: candidatePoints[anchorIndex].Longitude,
					},
				)
				if bestIndex < 0 ||
					distance < bestDistance ||
					(distance == bestDistance &&
						candidatePoints[anchorIndex].ObservedAt.UTC().Before(
							candidatePoints[bestIndex].ObservedAt.UTC(),
						)) {
					bestIndex = anchorIndex
					bestEndIndex = continuationEndIndex
					bestDistance = distance
				}
			}
		}

		segmentStart = segmentEnd + 1
	}

	if bestIndex < 0 {
		failure := anchorSearchFailureUnavailable
		if durationAvailable {
			failure = anchorSearchFailureDiscontinuous
		}
		return anchorSearchResult{Failure: failure}
	}

	return anchorSearchResult{
		Evidence: AnchorEvidence{
			PointIndex:             bestIndex,
			ObservedAt:             candidatePoints[bestIndex].ObservedAt.UTC(),
			DistanceKM:             bestDistance,
			ContinuationPointCount: bestEndIndex - bestIndex,
			ContinuationEndTime:    candidatePoints[bestEndIndex].ObservedAt.UTC(),
			MaximumObservedGap: maximumObservedGap(
				candidatePoints[bestIndex : bestEndIndex+1],
			),
		},
		Failure: anchorSearchFailureNone,
	}
}

func maximumObservedGap(
	points []trajectory.TrackPoint4D,
) time.Duration {
	maximum := time.Duration(0)
	for index := 1; index < len(points); index++ {
		gap := points[index].ObservedAt.UTC().Sub(
			points[index-1].ObservedAt.UTC(),
		)
		if gap > maximum {
			maximum = gap
		}
	}
	return maximum
}
