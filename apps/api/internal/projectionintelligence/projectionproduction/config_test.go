package projectionproduction

import (
	"errors"
	"testing"
)

func TestConfigValidateAcceptsExplicitPolicy(
	t *testing.T,
) {
	config := productionTestConfig()

	if err := config.Validate(); err != nil {
		t.Fatalf(
			"Validate() error = %v",
			err,
		)
	}
}

func TestConfigValidateRejectsMissingDependencyAndPolicy(
	t *testing.T,
) {
	tests := []struct {
		name      string
		mutate    func(*Config)
		wantError error
	}{
		{
			name: "horizon planner",
			mutate: func(config *Config) {
				config.HorizonPlanner = nil
			},
			wantError: ErrHorizonPlannerRequired,
		},
		{
			name: "historical projector",
			mutate: func(config *Config) {
				config.HistoricalProjector = nil
			},
			wantError: ErrHistoricalProjectorRequired,
		},
		{
			name: "freshness evaluator",
			mutate: func(config *Config) {
				config.FreshnessEvaluator = nil
			},
			wantError: ErrFreshnessEvaluatorRequired,
		},
		{
			name: "limited evidence policy",
			mutate: func(config *Config) {
				config.FreshnessLimitedPolicy = ""
			},
			wantError: ErrLimitedEvidencePolicyInvalid,
		},
		{
			name: "dependency failure policy",
			mutate: func(config *Config) {
				config.DependencyFailurePolicy = ""
			},
			wantError: ErrDependencyFailurePolicyInvalid,
		},
		{
			name: "arrival failure policy",
			mutate: func(config *Config) {
				config.ArrivalFailurePolicy = ""
			},
			wantError: ErrArrivalFailurePolicyInvalid,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				config := productionTestConfig()
				test.mutate(&config)

				err := config.Validate()
				if !errors.Is(
					err,
					test.wantError,
				) {
					t.Fatalf(
						"error = %v, want %v",
						err,
						test.wantError,
					)
				}
			},
		)
	}
}
