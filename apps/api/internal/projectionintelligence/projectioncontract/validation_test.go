package projectioncontract

import (
	"testing"
	"time"
)

func TestValidateAcceptsCompleteProjectionContract(
	t *testing.T,
) {
	report := Validate(
		validProjectionResult(),
	)

	if report.Status !=
		ValidationStatusValid {
		t.Fatalf(
			"validation status = %q, issues = %#v",
			report.Status,
			report.Issues,
		)
	}
	if len(report.Issues) != 0 {
		t.Fatalf(
			"expected no validation issues, got %#v",
			report.Issues,
		)
	}
}

func TestValidateAcceptsUnavailableProjectionContract(
	t *testing.T,
) {
	asOfTime := projectionTestAsOfTime()
	result := Result{
		SchemaVersion: SchemaVersionV1,
		Status:        ResultStatusUnavailable,
		TrajectoryID:  "trajectory-unavailable",
		Method: Method{
			Name:          "short_horizon_baseline",
			Version:       "short-horizon-baseline-v1",
			DecisionClass: DecisionClassProjectDerived,
		},
		Horizon: Horizon{
			AsOfTime: asOfTime,
			EndTime: asOfTime.Add(
				10 * time.Minute,
			),
			Step: time.Minute,
		},
		Confidence: Confidence{
			Score: 0,
			Level: ConfidenceLevelNone,
		},
		Limitations: []Limitation{
			{
				Code:    "projection_not_allowed",
				Message: "Trajectory eligibility denied projection.",
				Scope:   "result",
			},
		},
		ScopeGuard: ScopeGuardResearchOnly,
		GeneratedAt: asOfTime.Add(
			time.Second,
		),
	}

	report := Validate(result)
	if report.Status !=
		ValidationStatusValid {
		t.Fatalf(
			"validation status = %q, issues = %#v",
			report.Status,
			report.Issues,
		)
	}
}

func TestValidateRejectsProjectionContractViolations(
	t *testing.T,
) {
	tests := []struct {
		name     string
		mutate   func(*Result)
		wantCode string
	}{
		{
			name: "schema version",
			mutate: func(result *Result) {
				result.SchemaVersion = "future"
			},
			wantCode: IssueSchemaVersionInvalid,
		},
		{
			name: "scope guard",
			mutate: func(result *Result) {
				result.ScopeGuard = "operational"
			},
			wantCode: IssueScopeGuardInvalid,
		},
		{
			name: "future input evidence",
			mutate: func(result *Result) {
				future := result.Horizon.
					AsOfTime.Add(
					time.Second,
				)
				result.Provenance.
					Inputs[0].ObservedAt = future
				result.Provenance.
					LatestInputObservedAt = future
			},
			wantCode: IssueFutureInputEvidence,
		},
		{
			name: "latest input mismatch",
			mutate: func(result *Result) {
				result.Provenance.
					LatestInputObservedAt = result.
					Provenance.
					LatestInputObservedAt.Add(
					-time.Second,
				)
			},
			wantCode: IssueLatestInputMismatch,
		},
		{
			name: "missing horizontal uncertainty",
			mutate: func(result *Result) {
				result.Points[0].
					Uncertainty.
					HorizontalRadiusM = 0
			},
			wantCode: IssuePointUncertaintyInvalid,
		},
		{
			name: "missing vertical uncertainty",
			mutate: func(result *Result) {
				result.Points[0].
					Uncertainty.
					VerticalRadiusM = nil
			},
			wantCode: IssuePointUncertaintyInvalid,
		},
		{
			name: "unordered point time",
			mutate: func(result *Result) {
				result.Points[1].
					ForecastTime = result.
					Points[0].ForecastTime
			},
			wantCode: IssuePointTimeInvalid,
		},
		{
			name: "unavailable with values",
			mutate: func(result *Result) {
				result.Status =
					ResultStatusUnavailable
			},
			wantCode: IssueUnavailableContractInvalid,
		},
		{
			name: "invalid arrival interval",
			mutate: func(result *Result) {
				result.Arrival.EarliestTime =
					result.Arrival.LatestTime.Add(
						time.Minute,
					)
			},
			wantCode: IssueArrivalIntervalInvalid,
		},
		{
			name: "complete horizon not reached",
			mutate: func(result *Result) {
				result.Points = result.Points[:1]
			},
			wantCode: IssueCompleteHorizonNotReached,
		},
		{
			name: "estimated input without limitation",
			mutate: func(result *Result) {
				result.Provenance.Inputs[0].
					Classification =
					InputClassificationEstimated
				result.Provenance.Inputs[0].
					Limitation = ""
			},
			wantCode: IssueInputInvalid,
		},
		{
			name: "usable result without explanation",
			mutate: func(result *Result) {
				result.Explanations = nil
			},
			wantCode: IssueExplanationRequired,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				result := validProjectionResult()
				test.mutate(&result)

				report := Validate(result)
				if report.Status !=
					ValidationStatusInvalid {
					t.Fatalf(
						"validation status = %q, want invalid",
						report.Status,
					)
				}
				if !report.HasCode(
					test.wantCode,
				) {
					t.Fatalf(
						"issues = %#v, want code %q",
						report.Issues,
						test.wantCode,
					)
				}
			},
		)
	}
}

func TestValidationReportCloneDoesNotShareIssues(
	t *testing.T,
) {
	report := ValidationReport{
		Status: ValidationStatusInvalid,
		Issues: []ValidationIssue{
			{
				Code: "test",
			},
		},
	}
	cloned := report.Clone()
	cloned.Issues[0].Code = "changed"

	if report.Issues[0].Code != "test" {
		t.Fatal(
			"ValidationReport.Clone() shared issues",
		)
	}
}

func validProjectionResult() Result {
	asOfTime := projectionTestAsOfTime()
	firstAltitude := 11000.0
	secondAltitude := 10950.0
	firstVerticalUncertainty := 120.0
	secondVerticalUncertainty := 150.0

	return Result{
		SchemaVersion: SchemaVersionV1,
		Status:        ResultStatusComplete,

		TrajectoryID: "trajectory-001",
		FlightID:     "flight-001",
		AircraftID:   "aircraft-001",
		ICAO24:       "4K1234",
		Callsign:     "AHY123",

		Method: Method{
			Name:          "short_horizon_baseline",
			Version:       "short-horizon-baseline-v1",
			DecisionClass: DecisionClassProjectDerived,
		},
		Horizon: Horizon{
			AsOfTime: asOfTime,
			EndTime: asOfTime.Add(
				10 * time.Minute,
			),
			Step: 5 * time.Minute,
		},
		Points: []ProjectionPoint{
			{
				Sequence: 0,
				ForecastTime: asOfTime.Add(
					5 * time.Minute,
				),
				Position: Position{
					Latitude:  40.5,
					Longitude: 49.9,
					AltitudeM: &firstAltitude,
				},
				Uncertainty: Uncertainty{
					HorizontalRadiusM: 850,
					VerticalRadiusM:   &firstVerticalUncertainty,
				},
				Confidence: Confidence{
					Score: 0.72,
					Level: ConfidenceLevelMedium,
					Reasons: []ConfidenceReason{
						{
							Code:         "trajectory_quality",
							Message:      "Recent trajectory continuity supports a short projection.",
							Contribution: 0.72,
						},
					},
				},
			},
			{
				Sequence: 1,
				ForecastTime: asOfTime.Add(
					10 * time.Minute,
				),
				Position: Position{
					Latitude:  40.6,
					Longitude: 50.0,
					AltitudeM: &secondAltitude,
				},
				Uncertainty: Uncertainty{
					HorizontalRadiusM: 1600,
					VerticalRadiusM:   &secondVerticalUncertainty,
				},
				Confidence: Confidence{
					Score: 0.63,
					Level: ConfidenceLevelMedium,
					Reasons: []ConfidenceReason{
						{
							Code:         "horizon_uncertainty",
							Message:      "Uncertainty increases with forecast horizon.",
							Contribution: 0.63,
						},
					},
				},
			},
		},
		Arrival: &ArrivalEstimate{
			AirportICAOCode: "UBBB",
			EarliestTime: asOfTime.Add(
				15 * time.Minute,
			),
			EstimatedTime: asOfTime.Add(
				20 * time.Minute,
			),
			LatestTime: asOfTime.Add(
				25 * time.Minute,
			),
			Confidence: Confidence{
				Score: 0.61,
				Level: ConfidenceLevelMedium,
				Reasons: []ConfidenceReason{
					{
						Code:         "projection_baseline",
						Message:      "Arrival estimate uses the selected projection baseline.",
						Contribution: 0.61,
					},
				},
			},
			Limitations: []Limitation{
				{
					Code:    "no_official_flight_plan",
					Message: "Official flight plan and Air Traffic Control intent are unavailable.",
					Scope:   "arrival",
				},
			},
		},
		Confidence: Confidence{
			Score: 0.7,
			Level: ConfidenceLevelMedium,
			Reasons: []ConfidenceReason{
				{
					Code:         "bounded_short_horizon",
					Message:      "Projection remains inside the configured short horizon.",
					Contribution: 0.7,
				},
			},
		},
		Limitations: []Limitation{
			{
				Code:    "research_only",
				Message: "Projection is a research estimate, not an operational instruction.",
				Scope:   "result",
			},
		},
		Explanations: []Explanation{
			{
				Code:    "short_horizon_baseline",
				Message: "Positions are estimated from information available at the as-of time.",
			},
		},
		ScopeGuard: ScopeGuardResearchOnly,
		Provenance: Provenance{
			InputFingerprint: "sha256:projection-test",
			Inputs: []InputReference{
				{
					Name:           "latest_trajectory_point",
					Classification: InputClassificationObserved,
					SourceName:     "flight_trajectories",
					ObservedAt: asOfTime.Add(
						-time.Minute,
					),
					RetrievedAt: asOfTime.Add(
						10 * time.Second,
					),
				},
				{
					Name:           "derived_ground_track",
					Classification: InputClassificationDerived,
					SourceName:     "trajectory_features",
					ObservedAt: asOfTime.Add(
						-time.Minute,
					),
					RetrievedAt: asOfTime.Add(
						10 * time.Second,
					),
				},
			},
			LatestInputObservedAt: asOfTime.Add(
				-time.Minute,
			),
		},
		GeneratedAt: asOfTime.Add(
			20 * time.Second,
		),
	}
}

func projectionTestAsOfTime() time.Time {
	return time.Date(
		2026,
		time.July,
		15,
		17,
		0,
		0,
		0,
		time.UTC,
	)
}
