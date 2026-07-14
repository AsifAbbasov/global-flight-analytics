package routestore

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

func TestNormalizeResultProducesCanonicalCopy(
	t *testing.T,
) {
	result := validRouteResult()
	result.ICAO24 = " abc123 "
	result.Callsign = " J2001 "
	result.Provenance.SourceNames = []string{
		" trajectory ",
		"ourairports",
		"trajectory",
		"",
	}

	normalized := normalizeResult(result)

	if normalized.ICAO24 != "ABC123" ||
		normalized.Callsign != "J2001" ||
		!reflect.DeepEqual(
			normalized.Provenance.SourceNames,
			[]string{
				"ourairports",
				"trajectory",
			},
		) {
		t.Fatalf(
			"unexpected normalized result: %#v",
			normalized,
		)
	}
	if result.ICAO24 != " abc123 " ||
		len(result.Provenance.SourceNames) != 4 {
		t.Fatal(
			"normalizeResult() mutated input",
		)
	}
}

func TestValidateStorableResultAcceptsValidResult(
	t *testing.T,
) {
	report, err := validateStorableResult(
		validRouteResult(),
	)
	if err != nil {
		t.Fatalf(
			"validateStorableResult() error = %v",
			err,
		)
	}
	if report.Status !=
		routecontract.ValidationStatusValid ||
		report.ErrorCount != 0 {
		t.Fatalf(
			"unexpected report: %#v",
			report,
		)
	}
}

func TestValidateStorableResultRejectsInvalidContract(
	t *testing.T,
) {
	result := validRouteResult()
	result.ICAO24 = "bad"

	report, err := validateStorableResult(result)
	if !errors.Is(err, ErrResultInvalid) {
		t.Fatalf(
			"error = %v, want %v",
			err,
			ErrResultInvalid,
		)
	}
	if report.ErrorCount == 0 {
		t.Fatalf(
			"unexpected report: %#v",
			report,
		)
	}

	var validationErr *ValidationError
	if !errors.As(err, &validationErr) ||
		validationErr.Report.ErrorCount == 0 {
		t.Fatalf(
			"unexpected typed error: %#v",
			err,
		)
	}
}

func TestNormalizeResultKeyAndListDefaults(
	t *testing.T,
) {
	asOfTime := time.Date(
		2026,
		time.July,
		14,
		17,
		0,
		0,
		123,
		time.FixedZone("test", 3600),
	)
	key, err := normalizeResultKey(ResultKey{
		TrajectoryID:  " trajectory-one ",
		SchemaVersion: routecontract.SchemaVersionV1,
		AsOfTime:      asOfTime,
	})
	if err != nil {
		t.Fatalf(
			"normalizeResultKey() error = %v",
			err,
		)
	}
	if key.TrajectoryID != "trajectory-one" ||
		key.AsOfTime.Location() != time.UTC {
		t.Fatalf("unexpected key: %#v", key)
	}

	query, err := normalizeListQuery(ListQuery{
		TrajectoryID:  " trajectory-one ",
		SchemaVersion: routecontract.SchemaVersionV1,
	})
	if err != nil {
		t.Fatalf(
			"normalizeListQuery() error = %v",
			err,
		)
	}
	if query.Limit != DefaultListLimit ||
		query.TrajectoryID !=
			"trajectory-one" {
		t.Fatalf(
			"unexpected query: %#v",
			query,
		)
	}
}

func TestMakeRecordIDIsDeterministic(
	t *testing.T,
) {
	result := validRouteResult()
	key := resultKey(result)
	first := makeRecordID(
		encodeResultKey(key),
		result.Provenance.InputFingerprint,
	)
	second := makeRecordID(
		encodeResultKey(key),
		result.Provenance.InputFingerprint,
	)

	if first != second ||
		!strings.HasPrefix(
			first,
			"route-record-",
		) ||
		len(first) != 77 {
		t.Fatalf(
			"record identifiers = %q %q",
			first,
			second,
		)
	}
}

func validRouteResult() routecontract.Result {
	asOfTime := time.Date(
		2026,
		time.July,
		14,
		17,
		0,
		0,
		123456789,
		time.UTC,
	)
	origin := validEndpoint(
		routecontract.EndpointRoleOrigin,
		"UBBB",
		"GYD",
		40.4675,
		50.0467,
		asOfTime.Add(-50*time.Minute),
		0.90,
	)
	destination := validEndpoint(
		routecontract.EndpointRoleDestination,
		"UGTB",
		"TBS",
		41.6692,
		44.9547,
		asOfTime.Add(-time.Minute),
		0.85,
	)

	return routecontract.Result{
		SchemaVersion: routecontract.SchemaVersionV1,
		Status:        routecontract.RouteStatusComplete,
		TrajectoryID:  "8a3d6e20-2c68-4b35-a512-7d91e6a90c31",
		IdentityKey: "flight-identity-" +
			strings.Repeat("c", 64),
		FlightID:   "flight-one",
		AircraftID: "aircraft-one",
		ICAO24:     "ABC123",
		Callsign:   "J2001",
		Window: routecontract.RouteWindow{
			StartTime: asOfTime.Add(-time.Hour),
			EndTime:   asOfTime.Add(-time.Minute),
			AsOfTime:  asOfTime,
		},
		Origin:      origin,
		Destination: destination,
		Summary: routecontract.RouteSummary{
			GreatCircleDistanceKM: 448.8,
			SameAirport:           false,
		},
		Confidence: routecontract.Confidence{
			Score: 0.85,
			Level: routecontract.
				ConfidenceLevelHigh,
			EvidenceCount: 2,
			Reasons: []routecontract.ConfidenceReason{
				{
					Code:         "both_route_endpoints_available",
					Message:      "Both route endpoints are available.",
					Contribution: 0.85,
				},
			},
		},
		Limitations: []routecontract.Limitation{
			{
				Code:    "probable_route_only",
				Message: "Route endpoints are inferred.",
				Scope:   "route",
			},
		},
		Provenance: routecontract.Provenance{
			ResolverVersion: "route-resolver-v1",
			InputFingerprint: "sha256:" +
				strings.Repeat("a", 64),
			TrajectoryUpdatedAt: asOfTime.Add(-time.Minute),
			SourceNames: []string{
				"ourairports",
				"trajectory",
			},
		},
		GeneratedAt: asOfTime.Add(time.Second),
	}
}

func validEndpoint(
	role routecontract.EndpointRole,
	icaoCode string,
	iataCode string,
	latitude float64,
	longitude float64,
	observedAt time.Time,
	score float64,
) *routecontract.EndpointInference {
	return &routecontract.EndpointInference{
		Role: role,
		Airport: routecontract.AirportReference{
			ICAOCode:   icaoCode,
			IATACode:   iataCode,
			Name:       icaoCode + " Airport",
			City:       "City",
			Country:    "Country",
			Latitude:   latitude,
			Longitude:  longitude,
			ElevationM: 10,
			Timezone:   "UTC",
		},
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
	}
}
