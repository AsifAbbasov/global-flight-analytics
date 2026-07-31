package projectionarrival

import "time"

const (
	Version    = "estimated-arrival-boundary-v2"
	MethodName = "estimated_arrival_boundary"

	FingerprintVersion            = "estimated-arrival-boundary-fingerprint-v2"
	UnavailableFingerprintVersion = "estimated-arrival-unavailable-fingerprint-v2"
)

type EstimateMode string

const (
	EstimateModeWithinProjection EstimateMode = "within_projection_horizon"
	EstimateModeExtrapolated     EstimateMode = "extrapolated_beyond_projection_horizon"
)

type positionSampleSource string

const (
	positionSampleSourceCurrent    positionSampleSource = "current_trajectory"
	positionSampleSourceProjection positionSampleSource = "projection_point"
)

type positionSample struct {
	source     positionSampleSource
	sourceID   string
	sourceName string
	sequence   int

	timeValue time.Time

	latitude  float64
	longitude float64

	horizontalUncertaintyM float64
}

type positionEvidence struct {
	samples                []positionSample
	currentEndpoint        positionSample
	currentEndpointPresent bool
	fingerprint            string
}

type speedProfile struct {
	sampleCount int

	meanClosingSpeedMPS    float64
	closingSpeedStdDevMPS  float64
	minimumClosingSpeedMPS float64
	maximumClosingSpeedMPS float64
	maximumGroundSpeedMPS  float64
}

type arrivalComputation struct {
	mode EstimateMode

	earliestTime  time.Time
	estimatedTime time.Time
	latestTime    time.Time

	estimatedClosingSpeedMPS float64
	closingSpeedStdDevMPS    float64
	speedSampleCount         int

	remainingDistanceM       float64
	lastPositionUncertaintyM float64
	extrapolationDuration    time.Duration
}

type unavailableReason string

const (
	unavailableProjection            unavailableReason = "projection_unavailable"
	unavailableDestination           unavailableReason = "destination_unavailable"
	unavailableDestinationConfidence unavailableReason = "destination_confidence_below_minimum"
	unavailablePositionSamples       unavailableReason = "arrival_position_samples_unavailable"
	unavailableSpeedOrDuration       unavailableReason = "arrival_speed_or_duration_unavailable"
)
