package processor

import (
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/dataquality"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/deduplicator"
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
	DuplicateCount    int
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

func New(
	config Config,
) (*Processor, error) {
	if config.Now == nil {
		config.Now = func() time.Time {
			return time.Now().UTC()
		}
	}

	trackBuilder, err := trackbuilder.NewBuilder(
		config.TrackBuilderConfig,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create track builder: %w",
			err,
		)
	}

	return &Processor{
		config:       config,
		trackBuilder: trackBuilder,
	}, nil
}

func (processor *Processor) Process(
	states []flightstate.FlightState,
) ProcessingResult {
	normalizedStates := normalizer.NormalizeFlightStates(
		states,
	)

	deduplicationResult := deduplicator.RemoveExactDuplicates(
		normalizedStates,
	)

	uniqueStates := deduplicationResult.UniqueStates
	processedAt := processor.config.Now()

	result := ProcessingResult{
		UsableStates: make(
			[]ProcessedFlightState,
			0,
			len(uniqueStates),
		),
		InvalidStates: make(
			[]ProcessedFlightState,
			0,
		),
		Trajectories: make(
			map[string]trajectory.FlightTrajectory,
		),
		ProcessedAt: processedAt,
		Stats: ProcessingStats{
			ReceivedCount:  len(states),
			DuplicateCount: deduplicationResult.DuplicateCount,
		},
	}

	usableInputs := make(
		[]trackbuilder.InputState,
		0,
		len(uniqueStates),
	)

	for _, state := range uniqueStates {
		quality := validator.EvaluateFlightState(
			state,
			processedAt,
		)

		processedState := ProcessedFlightState{
			State:   state,
			Quality: quality,
		}

		result.Stats.TotalWarningCount += len(
			quality.Warnings,
		)

		switch quality.ValidationStatus {
		case dataquality.ValidationStatusInvalid:
			result.InvalidStates = append(
				result.InvalidStates,
				processedState,
			)

			result.Stats.InvalidCount++

		case dataquality.ValidationStatusPartial:
			recordUsableState(
				&result,
				&usableInputs,
				processedState,
			)

			result.Stats.PartialCount++

		default:
			recordUsableState(
				&result,
				&usableInputs,
				processedState,
			)

			result.Stats.ValidCount++
		}
	}

	result.Trajectories = processor.trackBuilder.BuildMany(
		usableInputs,
	)

	result.Stats.TrajectoryCount = len(
		result.Trajectories,
	)

	result.Stats.CoverageGapCount = countCoverageGaps(
		result.Trajectories,
	)

	return result
}

func recordUsableState(
	result *ProcessingResult,
	usableInputs *[]trackbuilder.InputState,
	processedState ProcessedFlightState,
) {
	result.UsableStates = append(
		result.UsableStates,
		processedState,
	)

	*usableInputs = append(
		*usableInputs,
		trackbuilder.InputState{
			State:        processedState.State,
			QualityScore: processedState.Quality.Score,
		},
	)

	result.Stats.UsableCount++
}

func countCoverageGaps(
	trajectories map[string]trajectory.FlightTrajectory,
) int {
	count := 0

	for _, item := range trajectories {
		count += item.CoverageGapCount
	}

	return count
}
