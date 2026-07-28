package projectionbaseline

import (
	"fmt"
	"hash"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

const FingerprintVersion = "projection-baseline-input-fingerprint-v2"

func writeTrajectoryFingerprintEvidence(
	digest hash.Hash,
	item trajectory.FlightTrajectory,
) {
	writeFingerprintString(digest, FingerprintVersion)
	writeFingerprintString(digest, item.IdentityKey)
	writeFingerprintString(digest, string(item.IdentityBasis))
	writeFingerprintString(digest, string(item.SplitReason))
	writeFingerprintString(digest, item.SourceName)
	writeFingerprintTime(digest, item.StartTime)
	writeFingerprintTime(digest, item.EndTime)
	writeFingerprintString(digest, fmt.Sprintf("%d", item.DurationSeconds))
	writeFingerprintString(digest, fmt.Sprintf("%d", item.PointCount))
	writeFingerprintString(digest, fmt.Sprintf("%d", item.SegmentCount))
	writeFingerprintString(digest, fmt.Sprintf("%d", item.CoverageGapCount))
	writeFingerprintFloat(digest, item.QualityScore)

	writeFingerprintString(digest, fmt.Sprintf("points:%d", len(item.Points)))
	for _, point := range item.Points {
		writeTrackPointFingerprintEvidence(digest, point)
	}

	writeFingerprintString(digest, fmt.Sprintf("segments:%d", len(item.Segments)))
	for _, segment := range item.Segments {
		writeSegmentFingerprintEvidence(digest, segment)
	}

	writeFingerprintString(digest, fmt.Sprintf("gaps:%d", len(item.CoverageGaps)))
	for _, gap := range item.CoverageGaps {
		writeGapFingerprintEvidence(digest, gap)
	}
}

func writeTrackPointFingerprintEvidence(
	digest hash.Hash,
	point trajectory.TrackPoint4D,
) {
	writeFingerprintString(digest, point.ID)
	writeFingerprintString(digest, point.FlightStateID)
	writeFingerprintTime(digest, point.ObservedAt)
	writeFingerprintFloat(digest, point.Latitude)
	writeFingerprintFloat(digest, point.Longitude)
	writeFingerprintFloat(digest, point.BarometricAltitudeM)
	writeFingerprintString(digest, string(point.BarometricAltitudeStatus))
	writeFingerprintFloat(digest, point.GeometricAltitudeM)
	writeFingerprintString(digest, string(point.GeometricAltitudeStatus))
	writeFingerprintFloat(digest, point.VelocityMPS)
	writeFingerprintString(digest, fmt.Sprintf("%t", point.VelocityAvailable))
	writeFingerprintFloat(digest, point.HeadingDegrees)
	writeFingerprintString(digest, fmt.Sprintf("%t", point.HeadingAvailable))
	writeFingerprintFloat(digest, point.VerticalRateMPS)
	writeFingerprintString(digest, fmt.Sprintf("%t", point.VerticalRateAvailable))
	writeFingerprintString(digest, fmt.Sprintf("%t", point.OnGround))
	writeFingerprintString(digest, fmt.Sprintf("%t", point.OnGroundAvailable))
	writeFingerprintString(digest, fmt.Sprintf("%t", point.TelemetryAvailabilityKnown))
	writeFingerprintString(digest, point.SourceName)
}

func writeSegmentFingerprintEvidence(
	digest hash.Hash,
	segment trajectory.TrajectorySegment,
) {
	writeFingerprintString(digest, segment.ID)
	writeFingerprintString(digest, fmt.Sprintf("%d", segment.SequenceNumber))
	writeFingerprintString(digest, string(segment.Status))
	writeFingerprintFloat(digest, segment.QualityScore)
	writeFingerprintTime(digest, segment.StartTime)
	writeFingerprintTime(digest, segment.EndTime)
	writeFingerprintString(digest, fmt.Sprintf("%d", segment.DurationSeconds))
	writeFingerprintString(digest, fmt.Sprintf("%d", segment.PointCount))
	writeFingerprintString(digest, segment.SourceName)
}

func writeGapFingerprintEvidence(
	digest hash.Hash,
	gap trajectory.CoverageGap,
) {
	writeFingerprintString(digest, gap.ID)
	writeFingerprintString(digest, gap.PreviousSegmentID)
	writeFingerprintString(digest, gap.NextSegmentID)
	writeFingerprintTime(digest, gap.StartTime)
	writeFingerprintTime(digest, gap.EndTime)
	writeFingerprintString(digest, fmt.Sprintf("%d", gap.DurationSeconds))
	writeFingerprintFloat(digest, gap.DistanceKm)
	writeFingerprintString(digest, string(gap.Reason))
	writeFingerprintString(digest, gap.FilledBy)
}
