package trajectory

// Clone returns a defensive copy of a trajectory and all mutable evidence
// slices. Evidence element types are value objects and contain no nested slices.
func (item FlightTrajectory) Clone() FlightTrajectory {
	cloned := item
	cloned.Points = append([]TrackPoint4D(nil), item.Points...)
	cloned.Segments = append([]TrajectorySegment(nil), item.Segments...)
	cloned.CoverageGaps = append([]CoverageGap(nil), item.CoverageGaps...)
	return cloned
}
