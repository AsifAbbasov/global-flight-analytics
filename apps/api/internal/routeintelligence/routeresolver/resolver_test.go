package routeresolver

import (
	"context"
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/endpointevidence"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

func TestNewUsesDefaults(t *testing.T) {
	resolver, err := New(Config{})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if resolver.partialConfidenceFactor !=
		DefaultPartialConfidenceFactor ||
		resolver.sameAirportConfidenceFactor !=
			DefaultSameAirportConfidenceFactor {
		t.Fatalf(
			"unexpected defaults: %#v",
			resolver,
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
			name: "partial factor",
			config: Config{
				PartialConfidenceFactor: math.Inf(1),
			},
			want: ErrInvalidPartialConfidenceFactor,
		},
		{
			name: "same-airport factor",
			config: Config{
				SameAirportConfidenceFactor: 1.1,
			},
			want: ErrInvalidSameAirportConfidenceFactor,
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

func TestResolveBuildsCompleteRoute(
	t *testing.T,
) {
	input, now := validInput()
	input.Origin = selectedEvidence(
		routecontract.EndpointRoleOrigin,
		airportReference(
			"UBBB",
			"GYD",
			40.4675,
			50.0467,
		),
		0.90,
		input.Window.StartTime,
		"a",
	)
	input.Destination = selectedEvidence(
		routecontract.EndpointRoleDestination,
		airportReference(
			"UGTB",
			"TBS",
			41.6692,
			44.9547,
		),
		0.85,
		input.Window.EndTime,
		"b",
	)

	resolver := mustResolver(
		t,
		Config{
			Now: func() time.Time {
				return now
			},
		},
	)

	resolution, err := resolver.Resolve(
		context.Background(),
		input,
	)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	result := resolution.Result
	if resolution.Version != Version ||
		result.Status !=
			routecontract.RouteStatusComplete ||
		result.Origin == nil ||
		result.Destination == nil ||
		result.Origin.Airport.ICAOCode !=
			"UBBB" ||
		result.Destination.Airport.ICAOCode !=
			"UGTB" ||
		result.Summary.SameAirport ||
		math.Abs(
			result.Summary.
				GreatCircleDistanceKM-448.8,
		) > 2 {
		t.Fatalf(
			"unexpected complete result: %#v",
			result,
		)
	}
	if math.Abs(
		result.Confidence.Score-0.85,
	) > 1e-12 ||
		result.Confidence.Level !=
			routecontract.ConfidenceLevelHigh ||
		result.Confidence.EvidenceCount != 2 {
		t.Fatalf(
			"unexpected confidence: %#v",
			result.Confidence,
		)
	}
	if result.ICAO24 != "ABC123" ||
		result.Callsign != "J2001" {
		t.Fatalf(
			"identity was not normalized: %#v",
			result,
		)
	}
	wantSources := []string{
		"ourairports",
		"trajectory",
		"trajectory_endpoint",
	}
	if !reflect.DeepEqual(
		result.Provenance.SourceNames,
		wantSources,
	) {
		t.Fatalf(
			"sources = %#v, want %#v",
			result.Provenance.SourceNames,
			wantSources,
		)
	}
	if result.Provenance.ResolverVersion != Version ||
		!strings.HasPrefix(
			result.Provenance.InputFingerprint,
			"sha256:",
		) ||
		len(result.Provenance.InputFingerprint) != 71 ||
		!result.GeneratedAt.Equal(now) {
		t.Fatalf(
			"unexpected provenance: %#v generated=%v",
			result.Provenance,
			result.GeneratedAt,
		)
	}
	if resolution.Validation.Status !=
		routecontract.ValidationStatusValid ||
		resolution.Validation.ErrorCount != 0 {
		t.Fatalf(
			"validation = %#v",
			resolution.Validation,
		)
	}
	assertLimitationCode(
		t,
		result.Limitations,
		"destination_not_planned_destination",
	)
	assertLimitationCode(
		t,
		result.Limitations,
		"probable_route_only",
	)
}

func TestResolveAppliesSameAirportPenalty(
	t *testing.T,
) {
	input, now := validInput()
	airport := airportReference(
		"UBBB",
		"GYD",
		40.4675,
		50.0467,
	)
	input.Origin = selectedEvidence(
		routecontract.EndpointRoleOrigin,
		airport,
		0.90,
		input.Window.StartTime,
		"a",
	)
	input.Destination = selectedEvidence(
		routecontract.EndpointRoleDestination,
		airport,
		0.80,
		input.Window.EndTime,
		"b",
	)

	resolution, err := mustResolver(
		t,
		Config{
			Now: func() time.Time {
				return now
			},
		},
	).Resolve(
		context.Background(),
		input,
	)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	result := resolution.Result
	if !result.Summary.SameAirport ||
		result.Summary.GreatCircleDistanceKM != 0 ||
		math.Abs(
			result.Confidence.Score-0.60,
		) > 1e-12 ||
		result.Confidence.Level !=
			routecontract.ConfidenceLevelMedium {
		t.Fatalf(
			"unexpected same-airport result: %#v",
			result,
		)
	}
	assertLimitationCode(
		t,
		result.Limitations,
		"same_airport_route_candidate",
	)
	assertReasonCode(
		t,
		result.Confidence.Reasons,
		"same_airport_candidate",
	)
}

func TestResolveBuildsPartialRoute(
	t *testing.T,
) {
	input, now := validInput()
	input.Origin = selectedEvidence(
		routecontract.EndpointRoleOrigin,
		airportReference(
			"UBBB",
			"GYD",
			40.4675,
			50.0467,
		),
		0.90,
		input.Window.StartTime,
		"a",
	)
	input.Destination = ambiguousEvidence(
		routecontract.EndpointRoleDestination,
		"b",
	)

	resolution, err := mustResolver(
		t,
		Config{
			Now: func() time.Time {
				return now
			},
		},
	).Resolve(
		context.Background(),
		input,
	)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	result := resolution.Result
	if result.Status !=
		routecontract.RouteStatusPartial ||
		result.Origin == nil ||
		result.Destination != nil ||
		result.Summary.GreatCircleDistanceKM != 0 ||
		result.Summary.SameAirport ||
		math.Abs(
			result.Confidence.Score-0.45,
		) > 1e-12 ||
		result.Confidence.EvidenceCount != 1 {
		t.Fatalf(
			"unexpected partial result: %#v",
			result,
		)
	}
	assertLimitationCode(
		t,
		result.Limitations,
		"destination_endpoint_ambiguous",
	)
	assertLimitationCode(
		t,
		result.Limitations,
		"route_incomplete",
	)
	if resolution.Validation.Status !=
		routecontract.ValidationStatusValid {
		t.Fatalf(
			"validation = %#v",
			resolution.Validation,
		)
	}
}

func TestResolveBuildsUnavailableRoute(
	t *testing.T,
) {
	input, now := validInput()
	input.Origin = unavailableEvidence(
		routecontract.EndpointRoleOrigin,
		"a",
	)
	input.Destination = unavailableEvidence(
		routecontract.EndpointRoleDestination,
		"b",
	)

	resolution, err := mustResolver(
		t,
		Config{
			Now: func() time.Time {
				return now
			},
		},
	).Resolve(
		context.Background(),
		input,
	)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	result := resolution.Result
	if result.Status !=
		routecontract.RouteStatusUnavailable ||
		result.Origin != nil ||
		result.Destination != nil ||
		result.Confidence.Score != 0 ||
		result.Confidence.Level !=
			routecontract.ConfidenceLevelNone ||
		result.Confidence.EvidenceCount != 0 {
		t.Fatalf(
			"unexpected unavailable result: %#v",
			result,
		)
	}
	assertLimitationCode(
		t,
		result.Limitations,
		"origin_endpoint_unavailable",
	)
	assertLimitationCode(
		t,
		result.Limitations,
		"destination_endpoint_unavailable",
	)
	if resolution.Validation.Status !=
		routecontract.ValidationStatusValid {
		t.Fatalf(
			"validation = %#v",
			resolution.Validation,
		)
	}
}

func TestResolveRejectsInvalidEndpointEvidence(
	t *testing.T,
) {
	input, now := validInput()
	input.Origin = unavailableEvidence(
		routecontract.EndpointRoleDestination,
		"a",
	)

	_, err := mustResolver(
		t,
		Config{
			Now: func() time.Time {
				return now
			},
		},
	).Resolve(
		context.Background(),
		input,
	)
	if !errors.Is(
		err,
		ErrInvalidEndpointEvidence,
	) {
		t.Fatalf(
			"Resolve() error = %v, want %v",
			err,
			ErrInvalidEndpointEvidence,
		)
	}
}

func TestResolveRequiresProvenanceSources(
	t *testing.T,
) {
	input, now := validInput()
	input.SourceNames = nil
	input.Origin = unavailableEvidence(
		routecontract.EndpointRoleOrigin,
		"a",
	)
	input.Destination = unavailableEvidence(
		routecontract.EndpointRoleDestination,
		"b",
	)

	_, err := mustResolver(
		t,
		Config{
			Now: func() time.Time {
				return now
			},
		},
	).Resolve(
		context.Background(),
		input,
	)
	if !errors.Is(err, ErrSourceNamesRequired) {
		t.Fatalf(
			"Resolve() error = %v, want %v",
			err,
			ErrSourceNamesRequired,
		)
	}
}

func TestResolveRejectsGeneratedTimeBeforeAsOfTime(
	t *testing.T,
) {
	input, _ := validInput()
	now := input.Window.AsOfTime.Add(-time.Second)

	_, err := mustResolver(
		t,
		Config{
			Now: func() time.Time {
				return now
			},
		},
	).Resolve(
		context.Background(),
		input,
	)
	if !errors.Is(
		err,
		ErrGeneratedBeforeAsOfTime,
	) {
		t.Fatalf(
			"Resolve() error = %v, want %v",
			err,
			ErrGeneratedBeforeAsOfTime,
		)
	}
}

func TestResolveRejectsInvalidFinalContract(
	t *testing.T,
) {
	input, now := validInput()
	input.ICAO24 = "bad"

	_, err := mustResolver(
		t,
		Config{
			Now: func() time.Time {
				return now
			},
		},
	).Resolve(
		context.Background(),
		input,
	)
	if !errors.Is(err, ErrContractValidation) {
		t.Fatalf(
			"Resolve() error = %v, want %v",
			err,
			ErrContractValidation,
		)
	}

	var validationErr *ContractValidationError
	if !errors.As(err, &validationErr) ||
		validationErr.Report.ErrorCount == 0 {
		t.Fatalf(
			"unexpected contract error: %#v",
			err,
		)
	}
}

func TestResolveRejectsFutureEndpointEvidence(
	t *testing.T,
) {
	input, now := validInput()
	input.Origin = selectedEvidence(
		routecontract.EndpointRoleOrigin,
		airportReference(
			"UBBB",
			"GYD",
			40.4675,
			50.0467,
		),
		0.90,
		input.Window.AsOfTime.Add(time.Second),
		"a",
	)

	_, err := mustResolver(
		t,
		Config{
			Now: func() time.Time {
				return now
			},
		},
	).Resolve(
		context.Background(),
		input,
	)
	if !errors.Is(err, ErrContractValidation) {
		t.Fatalf(
			"Resolve() error = %v, want %v",
			err,
			ErrContractValidation,
		)
	}
}

func TestResolveIsDeterministic(
	t *testing.T,
) {
	input, now := validInput()
	resolver := mustResolver(
		t,
		Config{
			Now: func() time.Time {
				return now
			},
		},
	)

	first, err := resolver.Resolve(
		context.Background(),
		input,
	)
	if err != nil {
		t.Fatalf(
			"first Resolve() error = %v",
			err,
		)
	}
	second, err := resolver.Resolve(
		context.Background(),
		input,
	)
	if err != nil {
		t.Fatalf(
			"second Resolve() error = %v",
			err,
		)
	}

	if !reflect.DeepEqual(first, second) {
		t.Fatalf(
			"resolutions differ:\nfirst=%#v\nsecond=%#v",
			first,
			second,
		)
	}
}

func TestResolvePreservesContextCancellation(
	t *testing.T,
) {
	input, now := validInput()
	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	_, err := mustResolver(
		t,
		Config{
			Now: func() time.Time {
				return now
			},
		},
	).Resolve(ctx, input)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"Resolve() error = %v, want context.Canceled",
			err,
		)
	}
}

func validInput() (Input, time.Time) {
	asOfTime := time.Date(
		2026,
		time.July,
		14,
		16,
		45,
		0,
		123456789,
		time.UTC,
	)
	now := asOfTime.Add(time.Second)

	return Input{
		TrajectoryID: "8a3d6e20-2c68-4b35-a512-7d91e6a90c31",
		IdentityKey: "flight-identity-" +
			strings.Repeat("c", 64),
		FlightID:   "flight-one",
		AircraftID: "aircraft-one",
		ICAO24:     " abc123 ",
		Callsign:   " J2001 ",
		Window: routecontract.RouteWindow{
			StartTime: asOfTime.Add(-time.Hour),
			EndTime:   asOfTime.Add(-time.Minute),
			AsOfTime:  asOfTime,
		},
		TrajectoryUpdatedAt: asOfTime.Add(-time.Minute),
		Origin: unavailableEvidence(
			routecontract.EndpointRoleOrigin,
			"a",
		),
		Destination: unavailableEvidence(
			routecontract.EndpointRoleDestination,
			"b",
		),
		SourceNames: []string{
			" trajectory ",
			"ourairports",
			"trajectory",
			"",
		},
	}, now
}

func selectedEvidence(
	role routecontract.EndpointRole,
	airport routecontract.AirportReference,
	score float64,
	observedAt time.Time,
	fingerprintCharacter string,
) endpointevidence.Result {
	return endpointevidence.Result{
		Version:                endpointevidence.Version,
		Status:                 endpointevidence.SelectionStatusSelected,
		Role:                   role,
		CandidateCount:         1,
		SelectedCandidateRank:  1,
		SelectedCandidateScore: score,
		CandidateScoreGap:      score,
		InputFingerprint: "sha256:" +
			strings.Repeat(
				fingerprintCharacter,
				64,
			),
		Endpoint: &routecontract.EndpointInference{
			Role:       role,
			Airport:    airport,
			DistanceKM: 2,
			Confidence: routecontract.Confidence{
				Score: score,
				Level: routecontract.
					ConfidenceLevelForScore(
						score,
					),
				EvidenceCount: 1,
				Reasons: []routecontract.ConfidenceReason{
					{
						Code: string(role) +
							"_airport_proximity",
						Message:      "Endpoint proximity evidence.",
						Contribution: score,
					},
				},
			},
			Evidence: []routecontract.Evidence{
				{
					Type: routecontract.
						EvidenceTypeTrajectoryEndpointProximity,
					SourceName:    "trajectory_endpoint",
					SourceVersion: "airport-candidate-resolver-v1",
					Score:         score,
					Weight:        1,
					ObservedAt:    observedAt,
					Summary:       "Trajectory endpoint evidence.",
					Attributes: []routecontract.EvidenceAttribute{
						{
							Key:   "distance_km",
							Value: "2.000000",
						},
						{
							Key:   "rank",
							Value: "1",
						},
					},
				},
			},
			Limitations: []routecontract.Limitation{
				{
					Code:    "probable_endpoint_only",
					Message: "Endpoint is inferred.",
					Scope:   string(role),
				},
			},
		},
		Limitations: []routecontract.Limitation{
			{
				Code:    "probable_endpoint_only",
				Message: "Endpoint is inferred.",
				Scope:   string(role),
			},
		},
	}
}

func unavailableEvidence(
	role routecontract.EndpointRole,
	fingerprintCharacter string,
) endpointevidence.Result {
	return endpointevidence.Result{
		Version: endpointevidence.Version,
		Status: endpointevidence.
			SelectionStatusUnavailable,
		Role: role,
		InputFingerprint: "sha256:" +
			strings.Repeat(
				fingerprintCharacter,
				64,
			),
		Limitations: []routecontract.Limitation{
			{
				Code:    "airport_candidate_unavailable",
				Message: "No airport candidate is available.",
				Scope:   string(role),
			},
		},
	}
}

func ambiguousEvidence(
	role routecontract.EndpointRole,
	fingerprintCharacter string,
) endpointevidence.Result {
	return endpointevidence.Result{
		Version: endpointevidence.Version,
		Status: endpointevidence.
			SelectionStatusAmbiguous,
		Role:                   role,
		CandidateCount:         2,
		SelectedCandidateRank:  1,
		SelectedCandidateScore: 0.80,
		RunnerUpCandidateScore: 0.78,
		CandidateScoreGap:      0.02,
		InputFingerprint: "sha256:" +
			strings.Repeat(
				fingerprintCharacter,
				64,
			),
		Limitations: []routecontract.Limitation{
			{
				Code:    "airport_candidate_ambiguous",
				Message: "Airport candidates are ambiguous.",
				Scope:   string(role),
			},
		},
	}
}

func airportReference(
	icaoCode string,
	iataCode string,
	latitude float64,
	longitude float64,
) routecontract.AirportReference {
	return routecontract.AirportReference{
		ICAOCode:   icaoCode,
		IATACode:   iataCode,
		Name:       icaoCode + " Airport",
		City:       "City",
		Country:    "Country",
		Latitude:   latitude,
		Longitude:  longitude,
		ElevationM: 10,
		Timezone:   "UTC",
	}
}

func mustResolver(
	t *testing.T,
	config Config,
) *Resolver {
	t.Helper()

	resolver, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	return resolver
}

func assertLimitationCode(
	t *testing.T,
	items []routecontract.Limitation,
	code string,
) {
	t.Helper()

	for _, item := range items {
		if item.Code == code {
			return
		}
	}

	t.Fatalf(
		"limitation %q not found in %#v",
		code,
		items,
	)
}

func assertReasonCode(
	t *testing.T,
	items []routecontract.ConfidenceReason,
	code string,
) {
	t.Helper()

	for _, item := range items {
		if item.Code == code {
			return
		}
	}

	t.Fatalf(
		"reason %q not found in %#v",
		code,
		items,
	)
}
