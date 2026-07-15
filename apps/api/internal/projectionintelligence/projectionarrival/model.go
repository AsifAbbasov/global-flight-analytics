package projectionarrival

import "time"

const (
	Version    = "estimated-arrival-boundary-v1"
	MethodName = "estimated_arrival_boundary"

	FingerprintVersion            = "estimated-arrival-boundary-fingerprint-v1"
	UnavailableFingerprintVersion = "estimated-arrival-unavailable-fingerprint-v1"
)

type EstimateMode string

const (
	EstimateModeWithinProjection EstimateMode = "within_projection_horizon"
	EstimateModeExtrapolated     EstimateMode = "extrapolated_beyond_projection_horizon"
)

type positionSample struct {
	timeValue time.Time

	latitude  float64
	longitude float64

	horizontalUncertaintyM float64
}

type speedProfile struct {
	sampleCount int

	meanMPS    float64
	stdDevMPS  float64
	minimumMPS float64
	maximumMPS float64
}

type arrivalComputation struct {
	mode EstimateMode

	earliestTime  time.Time
	estimatedTime time.Time
	latestTime    time.Time

	estimatedGroundSpeedMPS float64
	groundSpeedStdDevMPS    float64
	speedSampleCount        int

	remainingDistanceM       float64
	lastPositionUncertaintyM float64
	extrapolationDuration    time.Duration
}
