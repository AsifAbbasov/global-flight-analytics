package airportresolver

import (
	"context"
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

func TestNewUsesDefaults(t *testing.T) {
	catalog := mustCatalog(
		t,
		[]airport.Airport{
			airportFixture(
				"UBBB",
				"GYD",
				"Baku",
				40.4675,
				50.0467,
			),
		},
	)

	resolver, err := New(Config{
		Catalog: catalog,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if resolver.maximumDistanceKM !=
		DefaultMaximumDistanceKM ||
		resolver.maximumCandidates !=
			DefaultMaximumCandidates {
		t.Fatalf(
			"defaults = distance %v candidates %d",
			resolver.maximumDistanceKM,
			resolver.maximumCandidates,
		)
	}
}

func TestNewRejectsInvalidConfiguration(
	t *testing.T,
) {
	catalog := mustCatalog(
		t,
		[]airport.Airport{
			airportFixture(
				"UBBB",
				"GYD",
				"Baku",
				40.4675,
				50.0467,
			),
		},
	)

	tests := []struct {
		name   string
		config Config
		want   error
	}{
		{
			name:   "catalog required",
			config: Config{},
			want:   ErrCatalogRequired,
		},
		{
			name: "negative distance",
			config: Config{
				Catalog:           catalog,
				MaximumDistanceKM: -1,
			},
			want: ErrInvalidMaximumDistance,
		},
		{
			name: "non-finite distance",
			config: Config{
				Catalog:           catalog,
				MaximumDistanceKM: math.Inf(1),
			},
			want: ErrInvalidMaximumDistance,
		},
		{
			name: "candidate limit too large",
			config: Config{
				Catalog:           catalog,
				MaximumCandidates: MaximumCandidateLimit + 1,
			},
			want: ErrInvalidMaximumCandidates,
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

func TestResolveRanksCandidatesAndAppliesLimits(
	t *testing.T,
) {
	catalog := mustCatalog(
		t,
		[]airport.Airport{
			airportFixture(
				"UBBB",
				"GYD",
				"Heydar Aliyev International Airport",
				40.4675,
				50.0467,
			),
			airportFixture(
				"UBBY",
				"ZTU",
				"Zaqatala International Airport",
				41.5622,
				46.6672,
			),
			airportFixture(
				"UBBG",
				"GNJ",
				"Ganja International Airport",
				40.7377,
				46.3176,
			),
			airportFixture(
				"UGTB",
				"TBS",
				"Tbilisi International Airport",
				41.6692,
				44.9547,
			),
		},
	)
	resolver, err := New(Config{
		Catalog:           catalog,
		MaximumDistanceKM: 500,
		MaximumCandidates: 2,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result, err := resolver.Resolve(
		context.Background(),
		Query{
			Role: routecontract.EndpointRoleOrigin,
			Point: Point{
				Latitude:  40.47,
				Longitude: 50.05,
			},
		},
	)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if result.Version != ResolverVersion ||
		result.Role !=
			routecontract.EndpointRoleOrigin ||
		result.MaximumDistanceKM != 500 ||
		result.MaximumCandidates != 2 ||
		result.CatalogVersion != CatalogVersion ||
		result.CatalogAirportCount != 4 ||
		result.EligibleCandidateCount != 4 ||
		result.FilteredByRadiusCount != 0 ||
		result.TruncatedCandidateCount != 2 ||
		len(result.Candidates) != 2 {
		t.Fatalf(
			"unexpected result: %#v",
			result,
		)
	}
	if result.Candidates[0].Rank != 1 ||
		result.Candidates[0].Airport.ICAOCode !=
			"UBBB" ||
		result.Candidates[1].Rank != 2 ||
		result.Candidates[1].Airport.ICAOCode !=
			"UBBY" {
		t.Fatalf(
			"unexpected ranking: %#v",
			result.Candidates,
		)
	}
	if result.Candidates[0].DistanceKM >=
		result.Candidates[1].DistanceKM ||
		result.Candidates[0].ProximityScore <=
			result.Candidates[1].ProximityScore {
		t.Fatalf(
			"unexpected distance scores: %#v",
			result.Candidates,
		)
	}
	if !strings.HasPrefix(
		result.InputFingerprint,
		"sha256:",
	) || len(result.InputFingerprint) != 71 {
		t.Fatalf(
			"input fingerprint = %q",
			result.InputFingerprint,
		)
	}
}

func TestResolveFiltersOutsideRadius(
	t *testing.T,
) {
	catalog := mustCatalog(
		t,
		[]airport.Airport{
			airportFixture(
				"UBBB",
				"GYD",
				"Baku",
				40.4675,
				50.0467,
			),
			airportFixture(
				"UGTB",
				"TBS",
				"Tbilisi",
				41.6692,
				44.9547,
			),
		},
	)
	resolver, err := New(Config{
		Catalog:           catalog,
		MaximumDistanceKM: 25,
		MaximumCandidates: 5,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result, err := resolver.Resolve(
		context.Background(),
		Query{
			Role: routecontract.EndpointRoleDestination,
			Point: Point{
				Latitude:  40.4675,
				Longitude: 50.0467,
			},
		},
	)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if result.EligibleCandidateCount != 1 ||
		result.FilteredByRadiusCount != 1 ||
		result.TruncatedCandidateCount != 0 ||
		len(result.Candidates) != 1 ||
		result.Candidates[0].Airport.ICAOCode !=
			"UBBB" ||
		result.Candidates[0].DistanceKM > 1e-9 ||
		math.Abs(
			result.Candidates[0].ProximityScore-1,
		) > 1e-12 {
		t.Fatalf(
			"unexpected radius result: %#v",
			result,
		)
	}
}

func TestResolveIsDeterministicAcrossCatalogInputOrder(
	t *testing.T,
) {
	items := []airport.Airport{
		airportFixture(
			"UBBB",
			"GYD",
			"Baku",
			40.4675,
			50.0467,
		),
		airportFixture(
			"UGTB",
			"TBS",
			"Tbilisi",
			41.6692,
			44.9547,
		),
		airportFixture(
			"UBBG",
			"GNJ",
			"Ganja",
			40.7377,
			46.3176,
		),
	}

	firstCatalog := mustCatalog(t, items)
	secondCatalog := mustCatalog(
		t,
		[]airport.Airport{
			items[2],
			items[0],
			items[1],
		},
	)
	firstResolver := mustResolver(
		t,
		Config{
			Catalog:           firstCatalog,
			MaximumDistanceKM: 600,
			MaximumCandidates: 3,
		},
	)
	secondResolver := mustResolver(
		t,
		Config{
			Catalog:           secondCatalog,
			MaximumDistanceKM: 600,
			MaximumCandidates: 3,
		},
	)
	query := Query{
		Role: routecontract.EndpointRoleOrigin,
		Point: Point{
			Latitude:  40.4,
			Longitude: 49.9,
		},
	}

	first, err := firstResolver.Resolve(
		context.Background(),
		query,
	)
	if err != nil {
		t.Fatalf(
			"first Resolve() error = %v",
			err,
		)
	}
	second, err := secondResolver.Resolve(
		context.Background(),
		query,
	)
	if err != nil {
		t.Fatalf(
			"second Resolve() error = %v",
			err,
		)
	}

	if !reflect.DeepEqual(first, second) {
		t.Fatalf(
			"results differ:\nfirst=%#v\nsecond=%#v",
			first,
			second,
		)
	}
}

func TestResolveUsesStableTieBreakers(
	t *testing.T,
) {
	catalog := mustCatalog(
		t,
		[]airport.Airport{
			airportFixture(
				"ZZZZ",
				"ZZZ",
				"Second",
				0,
				1,
			),
			airportFixture(
				"AAAA",
				"AAA",
				"First",
				0,
				-1,
			),
		},
	)
	resolver := mustResolver(
		t,
		Config{
			Catalog:           catalog,
			MaximumDistanceKM: 200,
			MaximumCandidates: 2,
		},
	)

	result, err := resolver.Resolve(
		context.Background(),
		Query{
			Role:  routecontract.EndpointRoleOrigin,
			Point: Point{},
		},
	)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if len(result.Candidates) != 2 ||
		result.Candidates[0].Airport.ICAOCode !=
			"AAAA" ||
		result.Candidates[1].Airport.ICAOCode !=
			"ZZZZ" {
		t.Fatalf(
			"unexpected tie order: %#v",
			result.Candidates,
		)
	}
}

func TestResolveRejectsInvalidQuery(
	t *testing.T,
) {
	resolver := mustResolver(
		t,
		Config{
			Catalog: mustCatalog(
				t,
				[]airport.Airport{
					airportFixture(
						"UBBB",
						"GYD",
						"Baku",
						40.4675,
						50.0467,
					),
				},
			),
		},
	)

	_, err := resolver.Resolve(
		context.Background(),
		Query{
			Role: "unknown",
			Point: Point{
				Latitude:  40,
				Longitude: 50,
			},
		},
	)
	if !errors.Is(err, ErrInvalidEndpointRole) {
		t.Fatalf(
			"role error = %v, want %v",
			err,
			ErrInvalidEndpointRole,
		)
	}

	_, err = resolver.Resolve(
		context.Background(),
		Query{
			Role: routecontract.EndpointRoleOrigin,
			Point: Point{
				Latitude:  91,
				Longitude: 50,
			},
		},
	)
	if !errors.Is(err, ErrInvalidPoint) {
		t.Fatalf(
			"point error = %v, want %v",
			err,
			ErrInvalidPoint,
		)
	}
}

func TestResolvePreservesContextCancellation(
	t *testing.T,
) {
	resolver := mustResolver(
		t,
		Config{
			Catalog: mustCatalog(
				t,
				[]airport.Airport{
					airportFixture(
						"UBBB",
						"GYD",
						"Baku",
						40.4675,
						50.0467,
					),
				},
			),
		},
	)
	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	_, err := resolver.Resolve(
		ctx,
		Query{
			Role: routecontract.EndpointRoleOrigin,
			Point: Point{
				Latitude:  40,
				Longitude: 50,
			},
		},
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"Resolve() error = %v, want context.Canceled",
			err,
		)
	}
}

func TestResultCloneDoesNotShareCandidates(
	t *testing.T,
) {
	result := Result{
		Candidates: []Candidate{
			{
				Airport: routecontract.AirportReference{
					ICAOCode: "UBBB",
				},
			},
		},
	}
	cloned := result.Clone()
	cloned.Candidates[0].Airport.ICAOCode =
		"UGTB"

	if result.Candidates[0].Airport.ICAOCode !=
		"UBBB" {
		t.Fatal(
			"Result.Clone() shared candidates",
		)
	}
}

func TestVersionConstantsRemainStable(
	t *testing.T,
) {
	if CatalogVersion !=
		"airport-candidate-catalog-v1" {
		t.Fatalf(
			"CatalogVersion = %q",
			CatalogVersion,
		)
	}
	if ResolverVersion !=
		"airport-candidate-resolver-v1" {
		t.Fatalf(
			"ResolverVersion = %q",
			ResolverVersion,
		)
	}
}

func mustCatalog(
	t *testing.T,
	items []airport.Airport,
) *Catalog {
	t.Helper()

	catalog, _, err := NewCatalog(items)
	if err != nil {
		t.Fatalf("NewCatalog() error = %v", err)
	}

	return catalog
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
