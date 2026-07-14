package endpointevidence

import (
	"context"
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/airportresolver"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

func TestNewUsesDefaults(t *testing.T) {
	builder, err := New(Config{})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if builder.minimumSelectionScore !=
		DefaultMinimumSelectionScore ||
		builder.minimumCandidateScoreGap !=
			DefaultMinimumCandidateScoreGap {
		t.Fatalf(
			"unexpected defaults: %#v",
			builder,
		)
	}
}

func TestNewRejectsInvalidConfiguration(
	t *testing.T,
) {
	tests := []struct {
		name   string
		config Config
		want   error
	}{
		{
			name: "selection score",
			config: Config{
				MinimumSelectionScore: math.Inf(1),
			},
			want: ErrInvalidMinimumSelectionScore,
		},
		{
			name: "candidate score gap",
			config: Config{
				MinimumCandidateScoreGap: 1.1,
			},
			want: ErrInvalidMinimumCandidateScoreGap,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := New(test.config)
			if !errors.Is(err, test.want) {
				t.Fatalf(
					"New() error = %v, want %v",
					err,
					test.want,
				)
			}
		})
	}
}

func TestBuildSelectsContractSafeOrigin(
	t *testing.T,
) {
	builder := mustBuilder(t, Config{})
	input := validInput(
		routecontract.EndpointRoleOrigin,
		[]airportresolver.Candidate{
			candidate(
				1,
				"UBBB",
				"GYD",
				2.5,
				0.95,
			),
			candidate(
				2,
				"UBBG",
				"GNJ",
				85,
				0.25,
			),
		},
	)
	input.TrajectoryQuality = 0.9
	input.SegmentStatus =
		trajectory.SegmentStatusObserved
	input.SegmentPointCount = 8

	result, err := builder.Build(
		context.Background(),
		input,
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if result.Status != SelectionStatusSelected ||
		result.Endpoint == nil ||
		result.Role !=
			routecontract.EndpointRoleOrigin ||
		result.Endpoint.Role !=
			routecontract.EndpointRoleOrigin ||
		result.Endpoint.Airport.ICAOCode !=
			"UBBB" ||
		result.SelectedCandidateRank != 1 ||
		result.SelectedCandidateScore < 0.8 ||
		result.CandidateScoreGap <=
			DefaultMinimumCandidateScoreGap {
		t.Fatalf(
			"unexpected result: %#v",
			result,
		)
	}
	if result.Endpoint.Confidence.Level !=
		routecontract.ConfidenceLevelHigh ||
		result.Endpoint.Confidence.EvidenceCount != 1 ||
		len(result.Endpoint.Evidence) != 1 {
		t.Fatalf(
			"unexpected endpoint confidence: %#v",
			result.Endpoint,
		)
	}
	evidence := result.Endpoint.Evidence[0]
	if evidence.Type !=
		routecontract.
			EvidenceTypeTrajectoryEndpointProximity ||
		evidence.SourceVersion !=
			airportresolver.ResolverVersion ||
		!evidence.ObservedAt.Equal(input.ObservedAt) {
		t.Fatalf(
			"unexpected evidence: %#v",
			evidence,
		)
	}
	assertSortedAttributes(t, evidence.Attributes)
	assertUniqueSortedLimitations(
		t,
		result.Endpoint.Limitations,
	)
	if !strings.HasPrefix(
		result.InputFingerprint,
		"sha256:",
	) || len(result.InputFingerprint) != 71 {
		t.Fatalf(
			"fingerprint = %q",
			result.InputFingerprint,
		)
	}

	routeResult := routecontract.Result{
		SchemaVersion: routecontract.SchemaVersionV1,
		Status:        routecontract.RouteStatusPartial,
		TrajectoryID:  "trajectory-one",
		ICAO24:        "ABC123",
		Window: routecontract.RouteWindow{
			StartTime: input.ObservedAt.Add(-time.Hour),
			EndTime:   input.ObservedAt,
			AsOfTime:  input.ObservedAt,
		},
		Origin: result.Endpoint,
		Confidence: routecontract.Confidence{
			Score: result.Endpoint.Confidence.Score *
				0.5,
			Level: routecontract.
				ConfidenceLevelForScore(
					result.Endpoint.
						Confidence.Score *
						0.5,
				),
			EvidenceCount: 1,
			Reasons: []routecontract.ConfidenceReason{
				{
					Code:    "single_route_endpoint_available",
					Message: "Only one route endpoint is available.",
					Contribution: result.Endpoint.
						Confidence.Score *
						0.5,
				},
			},
		},
		Limitations: []routecontract.Limitation{
			{
				Code:    "destination_unavailable",
				Message: "Destination is not available.",
				Scope:   "route",
			},
		},
		Provenance: routecontract.Provenance{
			ResolverVersion: "route-resolver-v1",
			InputFingerprint: "sha256:" +
				strings.Repeat("a", 64),
			TrajectoryUpdatedAt: input.ObservedAt,
			SourceNames: []string{
				"trajectory",
			},
		},
		GeneratedAt: input.ObservedAt.Add(time.Second),
	}
	report := routecontract.Validate(routeResult)
	if report.Status !=
		routecontract.ValidationStatusValid {
		t.Fatalf(
			"route contract report = %#v",
			report,
		)
	}
}

func TestBuildAddsDestinationAndQualityLimitations(
	t *testing.T,
) {
	builder := mustBuilder(t, Config{})
	input := validInput(
		routecontract.EndpointRoleDestination,
		[]airportresolver.Candidate{
			candidate(
				1,
				"UGTB",
				"TBS",
				4,
				0.92,
			),
		},
	)
	input.CoverageGapCount = 2
	input.Candidates.TruncatedCandidateCount = 3
	input.Candidates.EligibleCandidateCount = 4
	input.Candidates.CatalogAirportCount = 4

	result, err := builder.Build(
		context.Background(),
		input,
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if result.Status != SelectionStatusSelected ||
		result.Endpoint == nil {
		t.Fatalf(
			"unexpected result: %#v",
			result,
		)
	}

	codes := limitationCodes(
		result.Endpoint.Limitations,
	)
	for _, code := range []string{
		"airport_candidates_truncated",
		"destination_not_planned_destination",
		"probable_endpoint_only",
		"trajectory_coverage_gaps",
	} {
		if _, exists := codes[code]; !exists {
			t.Fatalf(
				"limitation %q not found in %#v",
				code,
				result.Endpoint.Limitations,
			)
		}
	}
}

func TestBuildRejectsAmbiguousCandidates(
	t *testing.T,
) {
	builder := mustBuilder(t, Config{})
	input := validInput(
		routecontract.EndpointRoleOrigin,
		[]airportresolver.Candidate{
			candidate(
				1,
				"UBBB",
				"GYD",
				5,
				0.90,
			),
			candidate(
				2,
				"UBBG",
				"GNJ",
				6,
				0.86,
			),
		},
	)

	result, err := builder.Build(
		context.Background(),
		input,
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if result.Status !=
		SelectionStatusAmbiguous ||
		result.Endpoint != nil ||
		result.CandidateScoreGap >=
			DefaultMinimumCandidateScoreGap {
		t.Fatalf(
			"unexpected result: %#v",
			result,
		)
	}
	assertLimitationCode(
		t,
		result.Limitations,
		"airport_candidate_ambiguous",
	)
}

func TestBuildRejectsInsufficientCandidate(
	t *testing.T,
) {
	builder := mustBuilder(t, Config{})
	input := validInput(
		routecontract.EndpointRoleOrigin,
		[]airportresolver.Candidate{
			candidate(
				1,
				"UBBB",
				"GYD",
				115,
				0.04,
			),
		},
	)
	input.TrajectoryQuality = 0.1
	input.SegmentStatus =
		trajectory.SegmentStatusEstimated
	input.SegmentPointCount = 1

	result, err := builder.Build(
		context.Background(),
		input,
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if result.Status !=
		SelectionStatusInsufficient ||
		result.Endpoint != nil ||
		result.SelectedCandidateScore >=
			DefaultMinimumSelectionScore {
		t.Fatalf(
			"unexpected result: %#v",
			result,
		)
	}
	assertLimitationCode(
		t,
		result.Limitations,
		"endpoint_confidence_insufficient",
	)
}

func TestBuildReturnsUnavailableWithoutCandidates(
	t *testing.T,
) {
	builder := mustBuilder(t, Config{})
	input := validInput(
		routecontract.EndpointRoleOrigin,
		nil,
	)

	result, err := builder.Build(
		context.Background(),
		input,
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if result.Status !=
		SelectionStatusUnavailable ||
		result.Endpoint != nil ||
		result.CandidateCount != 0 {
		t.Fatalf(
			"unexpected result: %#v",
			result,
		)
	}
	assertLimitationCode(
		t,
		result.Limitations,
		"airport_candidate_unavailable",
	)
}

func TestBuildIsDeterministic(t *testing.T) {
	builder := mustBuilder(t, Config{})
	input := validInput(
		routecontract.EndpointRoleOrigin,
		[]airportresolver.Candidate{
			candidate(
				1,
				"UBBB",
				"GYD",
				2,
				0.95,
			),
		},
	)

	first, err := builder.Build(
		context.Background(),
		input,
	)
	if err != nil {
		t.Fatalf("first Build() error = %v", err)
	}
	second, err := builder.Build(
		context.Background(),
		input,
	)
	if err != nil {
		t.Fatalf("second Build() error = %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf(
			"results differ:\nfirst=%#v\nsecond=%#v",
			first,
			second,
		)
	}
}

func TestBuildPreservesContextCancellation(
	t *testing.T,
) {
	builder := mustBuilder(t, Config{})
	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	_, err := builder.Build(
		ctx,
		validInput(
			routecontract.EndpointRoleOrigin,
			nil,
		),
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"Build() error = %v, want context.Canceled",
			err,
		)
	}
}

func TestBuildRejectsInvalidInputs(t *testing.T) {
	builder := mustBuilder(t, Config{})
	base := validInput(
		routecontract.EndpointRoleOrigin,
		[]airportresolver.Candidate{
			candidate(
				1,
				"UBBB",
				"GYD",
				2,
				0.9,
			),
		},
	)

	tests := []struct {
		name   string
		mutate func(*Input)
		want   error
	}{
		{
			name: "candidate result",
			mutate: func(input *Input) {
				input.Candidates.Version =
					"unknown"
			},
			want: ErrInvalidCandidateResult,
		},
		{
			name: "observed at required",
			mutate: func(input *Input) {
				input.ObservedAt = time.Time{}
			},
			want: ErrObservedAtRequired,
		},
		{
			name: "observed at UTC",
			mutate: func(input *Input) {
				input.ObservedAt =
					input.ObservedAt.In(
						time.FixedZone(
							"test",
							3600,
						),
					)
			},
			want: ErrObservedAtNotUTC,
		},
		{
			name: "trajectory quality",
			mutate: func(input *Input) {
				input.TrajectoryQuality =
					math.NaN()
			},
			want: ErrInvalidTrajectoryQuality,
		},
		{
			name: "segment status",
			mutate: func(input *Input) {
				input.SegmentStatus =
					trajectory.SegmentStatusInvalid
			},
			want: ErrInvalidSegmentStatus,
		},
		{
			name: "segment point count",
			mutate: func(input *Input) {
				input.SegmentPointCount = 0
			},
			want: ErrInvalidSegmentPointCount,
		},
		{
			name: "coverage gap count",
			mutate: func(input *Input) {
				input.CoverageGapCount = -1
			},
			want: ErrInvalidCoverageGapCount,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := base
			input.Candidates =
				base.Candidates.Clone()
			test.mutate(&input)

			_, err := builder.Build(
				context.Background(),
				input,
			)
			if !errors.Is(err, test.want) {
				t.Fatalf(
					"Build() error = %v, want %v",
					err,
					test.want,
				)
			}
		})
	}
}

func validInput(
	role routecontract.EndpointRole,
	candidates []airportresolver.Candidate,
) Input {
	eligibleCount := len(candidates)
	catalogAirportCount := eligibleCount
	filteredByRadiusCount := 0
	if eligibleCount == 0 {
		catalogAirportCount = 1
		filteredByRadiusCount = 1
	}

	return Input{
		Candidates: airportresolver.Result{
			Version: airportresolver.ResolverVersion,
			Role:    role,
			Point: airportresolver.Point{
				Latitude:  40.4,
				Longitude: 49.9,
			},
			MaximumDistanceKM: airportresolver.
				DefaultMaximumDistanceKM,
			MaximumCandidates: airportresolver.
				DefaultMaximumCandidates,
			CatalogVersion: airportresolver.CatalogVersion,
			CatalogFingerprint: "sha256:" +
				strings.Repeat("a", 64),
			CatalogAirportCount:    catalogAirportCount,
			EligibleCandidateCount: eligibleCount,
			FilteredByRadiusCount:  filteredByRadiusCount,
			Candidates: append(
				[]airportresolver.Candidate(nil),
				candidates...,
			),
			InputFingerprint: "sha256:" +
				strings.Repeat("b", 64),
		},
		ObservedAt: time.Date(
			2026,
			time.July,
			14,
			16,
			0,
			0,
			123456789,
			time.UTC,
		),
		TrajectoryQuality: 0.8,
		SegmentStatus:     trajectory.SegmentStatusObserved,
		SegmentPointCount: 5,
	}
}

func candidate(
	rank int,
	icaoCode string,
	iataCode string,
	distanceKM float64,
	proximityScore float64,
) airportresolver.Candidate {
	return airportresolver.Candidate{
		Rank: rank,
		Airport: routecontract.AirportReference{
			ICAOCode:   icaoCode,
			IATACode:   iataCode,
			Name:       icaoCode + " Airport",
			City:       "City",
			Country:    "Country",
			Latitude:   40,
			Longitude:  50,
			ElevationM: 10,
			Timezone:   "UTC",
		},
		DistanceKM:     distanceKM,
		ProximityScore: proximityScore,
	}
}

func mustBuilder(
	t *testing.T,
	config Config,
) *Builder {
	t.Helper()

	builder, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	return builder
}

func assertSortedAttributes(
	t *testing.T,
	items []routecontract.EvidenceAttribute,
) {
	t.Helper()

	for index := 1; index < len(items); index++ {
		if items[index-1].Key >= items[index].Key {
			t.Fatalf(
				"attributes are not sorted and unique: %#v",
				items,
			)
		}
	}
}

func assertUniqueSortedLimitations(
	t *testing.T,
	items []routecontract.Limitation,
) {
	t.Helper()

	for index := 1; index < len(items); index++ {
		if items[index-1].Code >= items[index].Code {
			t.Fatalf(
				"limitations are not sorted and unique: %#v",
				items,
			)
		}
	}
}

func limitationCodes(
	items []routecontract.Limitation,
) map[string]struct{} {
	result := make(map[string]struct{})
	for _, item := range items {
		result[item.Code] = struct{}{}
	}

	return result
}

func assertLimitationCode(
	t *testing.T,
	items []routecontract.Limitation,
	code string,
) {
	t.Helper()

	if _, exists := limitationCodes(items)[code]; !exists {
		t.Fatalf(
			"limitation %q not found in %#v",
			code,
			items,
		)
	}
}
