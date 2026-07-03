package processor

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/dataquality"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/normalizer"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/trackbuilder"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/validator"
)

type Clock func() time.Time

type Config struct {
	Now                Clock
	TrackBuilderConfig trackbuilder.Config
}

type Processor struct {
	config       Config
	trackBuilder *trackbuilder.Builder
}

type ProcessedFlightState struct {
	State   flightstate.FlightState
	Quality dataquality.DataQuality
}

type ProcessingStats struct {
	ReceivedCount     int
	UsableCount       int
	InvalidCount      int
	ValidCount        int
	PartialCount      int
	TrajectoryCount   int
	CoverageGapCount  int
	TotalWarningCount int
}

type ProcessingResult struct {
	UsableStates  []ProcessedFlightState
	InvalidStates []ProcessedFlightState
	Trajectories  map[string]trajectory.FlightTrajectory
	Stats         ProcessingStats
	ProcessedAt   time.Time
}

func New(config Config) *Processor {
	if config.Now == nil {
		config.Now = func() time.Time {
			return time.Now().UTC()
		}
	}

	return &Processor{
		config:       config,
		trackBuilder: trackbuilder.NewBuilder(config.TrackBuilderConfig),
	}
}

func (processor *Processor) Process(states []flightstate.FlightState) ProcessingResult {
	normalizedStates := normalizer.NormalizeFlightStates(states)
	processedAt := processor.config.Now()

	result := ProcessingResult{
		UsableStates:  make([]ProcessedFlightState, 0, len(normalizedStates)),
		InvalidStates: make([]ProcessedFlightState, 0),
		Trajectories:  make(map[string]trajectory.FlightTrajectory),
		ProcessedAt:   processedAt,
		Stats: ProcessingStats{
			ReceivedCount: len(states),
		},
	}

	usableRawStates := make([]flightstate.FlightState, 0, len(normalizedStates))

	for _, state := range normalizedStates {
		quality := validator.EvaluateFlightState(state, processedAt)

		processedState := ProcessedFlightState{
			State:   state,
			Quality: quality,
		}

		result.Stats.TotalWarningCount += len(quality.Warnings)

		switch quality.ValidationStatus {
		case dataquality.ValidationStatusInvalid:
			result.InvalidStates = append(result.InvalidStates, processedState)
			result.Stats.InvalidCount++

		case dataquality.ValidationStatusPartial:
			result.UsableStates = append(result.UsableStates, processedState)
			usableRawStates = append(usableRawStates, state)
			result.Stats.PartialCount++
			result.Stats.UsableCount++

		default:
			result.UsableStates = append(result.UsableStates, processedState)
			usableRawStates = append(usableRawStates, state)
			result.Stats.ValidCount++
			result.Stats.UsableCount++
		}
	}

	result.Trajectories = processor.trackBuilder.BuildMany(usableRawStates)
	result.Stats.TrajectoryCount = len(result.Trajectories)
	result.Stats.CoverageGapCount = countCoverageGaps(result.Trajectories)

	return result
}

func countCoverageGaps(trajectories map[string]trajectory.FlightTrajectory) int {
	count := 0

	for _, item := range trajectories {
		count += item.CoverageGapCount
	}

	return count
}
