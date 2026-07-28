package historicalsimilarity

import "math"

type Config struct {
	MinimumPointCount int
	SampleCount       int

	GeometryScoreScaleKM float64
	EndpointScoreScaleKM float64

	GeometryWeight   float64
	EndpointsWeight  float64
	PathLengthWeight float64
	DurationWeight   float64
}

func DefaultConfig() Config {
	return Config{
		MinimumPointCount: DefaultMinimumPointCount,
		SampleCount:       DefaultSampleCount,

		GeometryScoreScaleKM: 250,
		EndpointScoreScaleKM: 100,

		GeometryWeight:   0.55,
		EndpointsWeight:  0.20,
		PathLengthWeight: 0.15,
		DurationWeight:   0.10,
	}
}

func (config Config) Validate() error {
	if config.MinimumPointCount < 2 ||
		config.MinimumPointCount >
			MaximumInputPointCount {
		return ErrMinimumPointCountInvalid
	}
	if config.SampleCount < 2 ||
		config.SampleCount > MaximumSampleCount {
		return ErrSampleCountInvalid
	}
	if !finite(config.GeometryScoreScaleKM) ||
		config.GeometryScoreScaleKM <= 0 ||
		!finite(config.EndpointScoreScaleKM) ||
		config.EndpointScoreScaleKM <= 0 {
		return ErrDistanceScaleInvalid
	}

	weights := []float64{
		config.GeometryWeight,
		config.EndpointsWeight,
		config.PathLengthWeight,
		config.DurationWeight,
	}
	total := 0.0
	compensation := 0.0
	for _, weight := range weights {
		if !finite(weight) || weight < 0 {
			return ErrWeightInvalid
		}
		corrected := weight - compensation
		next := total + corrected
		compensation = (next - total) - corrected
		total = next
	}
	if math.Abs(total-1) > weightTolerance {
		return ErrWeightInvalid
	}

	return nil
}

func (config Config) scoringPolicy() ScoringPolicy {
	return ScoringPolicy{
		GeometryScoreScaleKM: config.
			GeometryScoreScaleKM,
		EndpointScoreScaleKM: config.
			EndpointScoreScaleKM,
		GeometryWeight:   config.GeometryWeight,
		EndpointsWeight:  config.EndpointsWeight,
		PathLengthWeight: config.PathLengthWeight,
		DurationWeight:   config.DurationWeight,
	}
}

type Engine struct {
	config Config
}

func New(config Config) (*Engine, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return &Engine{config: config}, nil
}

func NewDefault() *Engine {
	// DefaultConfig is a compile-time-owned package contract and is protected
	// by tests and the permanent review audit. No panic path is necessary.
	return &Engine{config: DefaultConfig()}
}
