package researchbenchmark

import (
	"errors"
	"fmt"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/researchdataset"
)

type Kind string

const (
	KindTransponderEvidence  Kind = "transponder_evidence"
	KindClimbPrediction      Kind = "climb_prediction"
	KindOpenSkyCompatibility Kind = "opensky_compatibility"
	KindHistoricalReplay     Kind = "historical_replay"
	KindTakeoffWeight        Kind = "takeoff_weight"
)

var ErrPlanInvalid = errors.New(
	"research benchmark plan is invalid",
)

type Plan struct {
	ID        string
	Kind      Kind
	DatasetID researchdataset.ID

	Metrics []string

	MaximumRecords             int64
	OfflineOnly                bool
	ProductionDependency       bool
	RequiresApplicabilityGuard bool
	Limitations                []string
}

func DefaultPlans() []Plan {
	return []Plan{
		{
			ID:        "transponder-evidence-v1",
			Kind:      KindTransponderEvidence,
			DatasetID: researchdataset.IDEmergencyReference,
			Metrics: []string{
				"special_code_retention_ratio",
				"event_window_reconstruction_ratio",
				"unsupported_confirmed_incident_claim_count",
			},
			MaximumRecords:             250_000,
			OfflineOnly:                true,
			RequiresApplicabilityGuard: true,
			Limitations: []string{
				"Benchmark validates observed code evidence, not confirmed incident truth.",
			},
		},
		{
			ID:        "climb-prediction-v1",
			Kind:      KindClimbPrediction,
			DatasetID: researchdataset.IDClimbingAircraft,
			Metrics: []string{
				"altitude_error_2_min_m",
				"altitude_error_5_min_m",
				"altitude_error_10_min_m",
				"speed_error_10_min_mps",
				"prediction_interval_coverage",
			},
			MaximumRecords:             1_000_000,
			OfflineOnly:                true,
			RequiresApplicabilityGuard: true,
			Limitations: []string{
				"Historical 2017 climb distributions may not represent current regional operations.",
			},
		},
		{
			ID:        "opensky-schema-compatibility-v1",
			Kind:      KindOpenSkyCompatibility,
			DatasetID: researchdataset.IDTrinoSnapshot2026,
			Metrics: []string{
				"record_parse_success_ratio",
				"nullable_field_preservation_ratio",
				"position_source_preservation_ratio",
				"category_availability_preservation_ratio",
			},
			MaximumRecords:             1_000_000,
			OfflineOnly:                true,
			RequiresApplicabilityGuard: true,
			Limitations: []string{
				"Only allowlisted non-satellite tables may be evaluated.",
			},
		},
		{
			ID:        "external-historical-replay-v1",
			Kind:      KindHistoricalReplay,
			DatasetID: researchdataset.IDWeeklyStateVectors,
			Metrics: []string{
				"usable_observation_ratio",
				"trajectory_reconstruction_count",
				"coverage_gap_count",
				"deterministic_replay_fingerprint_match",
			},
			MaximumRecords:             1_000_000,
			OfflineOnly:                true,
			RequiresApplicabilityGuard: true,
			Limitations: []string{
				"Monday-only samples cannot support general weekly seasonality claims.",
			},
		},
		{
			ID:        "takeoff-weight-baseline-v1",
			Kind:      KindTakeoffWeight,
			DatasetID: researchdataset.IDPRCTakeoffWeight,
			Metrics: []string{
				"mean_absolute_error_kg",
				"root_mean_squared_error_kg",
				"prediction_interval_coverage",
				"unsupported_input_ratio",
				"out_of_distribution_ratio",
			},
			MaximumRecords:             500_000,
			OfflineOnly:                true,
			RequiresApplicabilityGuard: true,
			Limitations: []string{
				"Outputs are estimated mass ranges and never live measured take-off weight.",
			},
		},
	}
}

func Validate(
	plan Plan,
) error {
	if strings.TrimSpace(plan.ID) == "" ||
		plan.Kind == "" ||
		len(plan.Metrics) == 0 ||
		plan.MaximumRecords <= 0 ||
		!plan.OfflineOnly ||
		plan.ProductionDependency ||
		!plan.RequiresApplicabilityGuard {
		return ErrPlanInvalid
	}

	profile, err := researchdataset.ProfileByID(
		plan.DatasetID,
	)
	if err != nil {
		return err
	}
	if profile.Selection !=
		researchdataset.SelectionAdopted {
		return fmt.Errorf(
			"%w: dataset=%s selection=%s",
			ErrPlanInvalid,
			plan.DatasetID,
			profile.Selection,
		)
	}
	if plan.MaximumRecords >
		profile.MaximumRecords {
		return fmt.Errorf(
			"%w: maximum records=%d profile maximum=%d",
			ErrPlanInvalid,
			plan.MaximumRecords,
			profile.MaximumRecords,
		)
	}

	seen := make(map[string]struct{})
	for _, metric := range plan.Metrics {
		normalized := strings.TrimSpace(metric)
		if normalized == "" {
			return ErrPlanInvalid
		}
		if _, exists := seen[normalized]; exists {
			return ErrPlanInvalid
		}
		seen[normalized] = struct{}{}
	}

	return nil
}
