package projectioncontinuation

type projectedSample struct {
	trajectoryID string
	weight       float64

	latitude  float64
	longitude float64
	altitudeM *float64
}
