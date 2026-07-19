package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type fragmentCount struct {
	Fragment string
	Minimum  int
	Maximum  int
}

type fileRule struct {
	Name      string
	Path      string
	Required  []string
	Forbidden []string
	Counts    []fragmentCount
}

type auditFailure struct {
	Check  string
	Detail string
}

func main() {
	os.Exit(run(
		os.Args[1:],
		os.Stdout,
		os.Stderr,
	))
}

func run(
	args []string,
	stdout io.Writer,
	stderr io.Writer,
) int {
	flags := flag.NewFlagSet(
		"backendfinalaudit",
		flag.ContinueOnError,
	)
	flags.SetOutput(stderr)

	rootValue := flags.String(
		"root",
		"",
		"repository root; auto-detected when omitted",
	)
	strict := flags.Bool(
		"strict",
		true,
		"return a non-zero exit code when a correctness invariant fails",
	)

	if err := flags.Parse(args); err != nil {
		return 1
	}

	repositoryRoot, err := resolveRepositoryRoot(
		*rootValue,
	)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"locate repository root: %v\n",
			err,
		)
		return 1
	}

	failures := auditRepository(
		repositoryRoot,
		stdout,
	)
	if len(failures) == 0 {
		fmt.Fprintln(
			stdout,
			"Backend final correctness audit: PASS",
		)
		return 0
	}

	fmt.Fprintln(
		stderr,
		"Backend final correctness audit: FAIL",
	)
	for _, failure := range failures {
		fmt.Fprintf(
			stderr,
			"- %s: %s\n",
			failure.Check,
			failure.Detail,
		)
	}

	if *strict {
		return 1
	}
	return 0
}

func resolveRepositoryRoot(
	explicit string,
) (string, error) {
	if strings.TrimSpace(explicit) != "" {
		root, err := filepath.Abs(
			strings.TrimSpace(explicit),
		)
		if err != nil {
			return "", err
		}
		if err := validateRepositoryRoot(root); err != nil {
			return "", err
		}
		return root, nil
	}

	current, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if validateRepositoryRoot(current) == nil {
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", errors.New(
				"repository root containing apps/api/go.mod and apps/web/package.json was not found",
			)
		}
		current = parent
	}
}

func validateRepositoryRoot(
	root string,
) error {
	required := []string{
		filepath.Join(
			root,
			"apps",
			"api",
			"go.mod",
		),
		filepath.Join(
			root,
			"apps",
			"web",
			"package.json",
		),
	}
	for _, path := range required {
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf(
				"required repository file %s: %w",
				path,
				err,
			)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf(
				"required repository path is not a file: %s",
				path,
			)
		}
	}
	return nil
}

func auditRepository(
	root string,
	output io.Writer,
) []auditFailure {
	groups := []struct {
		name  string
		rules []fileRule
	}{
		{
			name:  "Projection read snapshot consistency",
			rules: projectionSnapshotRules(),
		},
		{
			name:  "Nullable telemetry integrity",
			rules: nullableTelemetryRules(),
		},
		{
			name:  "End-to-end telemetry availability",
			rules: endToEndTelemetryAvailabilityRules(),
		},
		{
			name:  "Historical pagination integrity",
			rules: historicalPaginationRules(),
		},
		{
			name:  "Weather composition boundary",
			rules: weatherCompositionRules(),
		},
		{
			name:  "Regression and verification evidence",
			rules: evidenceRules(),
		},
	}

	failures := make(
		[]auditFailure,
		0,
	)
	for _, group := range groups {
		groupFailures := auditRules(
			root,
			group.rules,
		)
		if len(groupFailures) == 0 {
			fmt.Fprintf(
				output,
				"%s: PASS\n",
				group.name,
			)
			continue
		}

		failures = append(
			failures,
			groupFailures...,
		)
	}

	return failures
}

func auditRules(
	root string,
	rules []fileRule,
) []auditFailure {
	failures := make(
		[]auditFailure,
		0,
	)

	for _, rule := range rules {
		content, err := os.ReadFile(
			filepath.Join(
				root,
				filepath.FromSlash(rule.Path),
			),
		)
		if err != nil {
			failures = append(
				failures,
				auditFailure{
					Check: rule.Name,
					Detail: fmt.Sprintf(
						"read %s: %v",
						rule.Path,
						err,
					),
				},
			)
			continue
		}

		text := string(content)
		for _, fragment := range rule.Required {
			if strings.Contains(text, fragment) {
				continue
			}
			failures = append(
				failures,
				auditFailure{
					Check: rule.Name,
					Detail: fmt.Sprintf(
						"%s is missing required fragment %q",
						rule.Path,
						fragment,
					),
				},
			)
		}

		for _, fragment := range rule.Forbidden {
			if !strings.Contains(text, fragment) {
				continue
			}
			failures = append(
				failures,
				auditFailure{
					Check: rule.Name,
					Detail: fmt.Sprintf(
						"%s contains forbidden fragment %q",
						rule.Path,
						fragment,
					),
				},
			)
		}

		for _, expectation := range rule.Counts {
			actual := strings.Count(
				text,
				expectation.Fragment,
			)
			if actual < expectation.Minimum {
				failures = append(
					failures,
					auditFailure{
						Check: rule.Name,
						Detail: fmt.Sprintf(
							"%s contains %q %d times, minimum is %d",
							rule.Path,
							expectation.Fragment,
							actual,
							expectation.Minimum,
						),
					},
				)
			}
			if expectation.Maximum >= 0 &&
				actual > expectation.Maximum {
				failures = append(
					failures,
					auditFailure{
						Check: rule.Name,
						Detail: fmt.Sprintf(
							"%s contains %q %d times, maximum is %d",
							rule.Path,
							expectation.Fragment,
							actual,
							expectation.Maximum,
						),
					},
				)
			}
		}
	}

	sort.Slice(
		failures,
		func(left int, right int) bool {
			if failures[left].Check ==
				failures[right].Check {
				return failures[left].Detail <
					failures[right].Detail
			}
			return failures[left].Check <
				failures[right].Check
		},
	)

	return failures
}

func projectionSnapshotRules() []fileRule {
	return []fileRule{
		{
			Name: "Projection DataSource exposes one snapshot read",
			Path: "apps/api/internal/projectionintelligence/projectionread/contracts.go",
			Required: []string{
				"type DataSource interface",
				"LoadSnapshot(",
				"SnapshotRequest",
			},
			Forbidden: []string{
				"LoadCurrentTrajectory(",
				"LoadRoute(",
				"LoadHistoricalCandidates(",
				"LoadRouteHistory(",
			},
			Counts: []fragmentCount{
				{
					Fragment: "LoadSnapshot(",
					Minimum:  1,
					Maximum:  1,
				},
			},
		},
		{
			Name: "Projection service performs one atomic source read",
			Path: "apps/api/internal/projectionintelligence/projectionread/service.go",
			Required: []string{
				"service.dataSource.LoadSnapshot(",
				"snapshot.CurrentTrajectory",
				"snapshot.HistoricalCandidates",
				"snapshot.RouteHistory",
			},
			Forbidden: []string{
				"dataSource.LoadCurrentTrajectory(",
				"dataSource.LoadRoute(",
				"dataSource.LoadHistoricalCandidates(",
				"dataSource.LoadRouteHistory(",
			},
			Counts: []fragmentCount{
				{
					Fragment: "service.dataSource.LoadSnapshot(",
					Minimum:  1,
					Maximum:  1,
				},
			},
		},
		{
			Name: "Projection PostgreSQL snapshot is repeatable and read-only",
			Path: "apps/api/internal/projectionintelligence/projectionread/postgres_snapshot.go",
			Required: []string{
				"pgx.RepeatableRead",
				"pgx.ReadOnly",
				"transaction.Client()",
				"transaction.TrajectoryRepository()",
				"context.WithoutCancel(ctx)",
				"transaction.Commit(ctx)",
				"transaction.Rollback(rollbackContext)",
			},
		},
		{
			Name: "Projection production wiring selects transactional snapshot executor",
			Path: "apps/api/internal/projectionintelligence/projectionread/postgres_config.go",
			Required: []string{
				"source.snapshotExecutor = repeatableReadSnapshotExecutor",
				"starter: pgxSnapshotTransactionStarter",
				"pool: config.Pool",
			},
		},
		{
			Name: "Projection snapshot session owns all dependent reads",
			Path: "apps/api/internal/projectionintelligence/projectionread/postgres_snapshot_source.go",
			Required: []string{
				"source.snapshotExecutor.Execute(",
				"client postgresClient",
				"repository trajectoryRepository",
				"session.loadSnapshotWithinSession(",
				"source.LoadCurrentTrajectory(",
				"source.LoadRoute(",
				"source.LoadHistoricalCandidates(",
				"source.LoadRouteHistory(",
			},
		},
	}
}

func nullableTelemetryRules() []fileRule {
	requiredColumns := []string{
		"latitude::float8",
		"longitude::float8",
		"velocity_mps::float8",
		"heading_degrees::float8",
		"vertical_rate_mps::float8",
	}
	requiredPredicates := []string{
		"latitude IS NOT NULL",
		"longitude IS NOT NULL",
		"velocity_mps IS NOT NULL",
		"heading_degrees IS NOT NULL",
		"vertical_rate_mps IS NOT NULL",
		"on_ground IS NOT NULL",
	}

	counts := make(
		[]fragmentCount,
		0,
		len(requiredColumns)+
			len(requiredPredicates),
	)
	for _, fragment := range requiredColumns {
		counts = append(
			counts,
			fragmentCount{
				Fragment: fragment,
				Minimum:  2,
				Maximum:  2,
			},
		)
	}
	for _, fragment := range requiredPredicates {
		counts = append(
			counts,
			fragmentCount{
				Fragment: fragment,
				Minimum:  2,
				Maximum:  2,
			},
		)
	}

	return []fileRule{
		{
			Name: "Projection trajectory SQL preserves nullable telemetry",
			Path: "apps/api/internal/projectionintelligence/projectionread/postgres_queries.go",
			Forbidden: []string{
				"COALESCE(latitude",
				"COALESCE(longitude",
				"COALESCE(velocity_mps",
				"COALESCE(heading_degrees",
				"COALESCE(vertical_rate_mps",
				"COALESCE(on_ground",
			},
			Counts: counts,
		},
		{
			Name: "Projection telemetry scanner rejects incomplete rows",
			Path: "apps/api/internal/projectionintelligence/projectionread/postgres_source.go",
			Required: []string{
				"func scanTrackPoint(",
				"var latitude pgtype.Float8",
				"var longitude pgtype.Float8",
				"var velocity pgtype.Float8",
				"var heading pgtype.Float8",
				"var verticalRate pgtype.Float8",
				"var onGround pgtype.Bool",
				"if !completeRequiredTelemetry(",
				"if !usable",
				"return latitude.Valid &&",
				"longitude.Valid &&",
				"velocity.Valid &&",
				"heading.Valid &&",
				"verticalRate.Valid &&",
				"onGround.Valid",
			},
		},
	}
}

func endToEndTelemetryAvailabilityRules() []fileRule {
	return []fileRule{
		{
			Name: "Flight State exposes explicit telemetry availability",
			Path: "apps/api/internal/domain/flightstate/model.go",
			Required: []string{
				"TelemetryAvailabilityKnown",
				"VelocityAvailable",
				"HeadingAvailable",
				"VerticalRateAvailable",
				"OnGroundAvailable",
			},
		},
		{
			Name: "Flight State availability helpers preserve explicit zero",
			Path: "apps/api/internal/domain/flightstate/telemetry_availability.go",
			Required: []string{
				"func (state FlightState) HasVelocity() bool",
				"func (state FlightState) HasHeading() bool",
				"func (state FlightState) HasVerticalRate() bool",
				"func (state FlightState) HasOnGroundState() bool",
				"func (state FlightState) HasCompleteKinematics() bool",
				"if !state.TelemetryAvailabilityKnown",
			},
		},
		{
			Name: "OpenSky preserves optional kinematic availability",
			Path: "apps/api/internal/integrations/opensky/provider.go",
			Required: []string{
				"optionalFiniteFloat64(",
				"VelocityAvailable:",
				"HeadingAvailable:",
				"VerticalRateAvailable:",
				"OnGroundAvailable:",
				"TelemetryAvailabilityKnown:",
				"return 0, false",
				"return *value, true",
			},
			Forbidden: []string{
				"optionalFloat64Value(",
			},
		},
		{
			Name: "Airplanes live declares provider telemetry availability",
			Path: "apps/api/internal/integrations/airplaneslive/mapper.go",
			Required: []string{
				"TelemetryAvailabilityKnown:",
				"VelocityAvailable:",
				"HeadingAvailable:",
				"VerticalRateAvailable:",
				"OnGroundAvailable:",
			},
		},
		{
			Name: "Flight State persistence writes nullable telemetry",
			Path: "apps/api/internal/repository/postgres/flightstate_repository.go",
			Required: []string{
				"telemetryFloatDatabaseValue(",
				"item.HasVelocity()",
				"item.HasHeading()",
				"item.HasVerticalRate()",
				"telemetryBoolDatabaseValue(",
				"item.HasOnGroundState()",
				"applyTelemetryDatabaseValues(",
				"velocity_mps::double precision",
				"heading_degrees::double precision",
				"vertical_rate_mps::double precision",
			},
			Forbidden: []string{
				"COALESCE(velocity_mps, 0)",
				"COALESCE(heading_degrees, 0)",
				"COALESCE(vertical_rate_mps, 0)",
				"COALESCE(on_ground, false)",
			},
		},
		{
			Name: "Reconciliation preserves nullable telemetry",
			Path: "apps/api/internal/repository/postgres/flightstate_reconciliation_repository.go",
			Required: []string{
				"velocity_mps::double precision",
				"heading_degrees::double precision",
				"vertical_rate_mps::double precision",
				"applyTelemetryDatabaseValues(",
			},
			Forbidden: []string{
				"COALESCE(velocity_mps, 0)",
				"COALESCE(heading_degrees, 0)",
				"COALESCE(vertical_rate_mps, 0)",
				"COALESCE(on_ground, false)",
			},
		},
		{
			Name: "Traffic excludes incomplete display kinematics",
			Path: "apps/api/internal/repository/postgres/traffic_repository.go",
			Required: []string{
				"fs.velocity_mps IS NOT NULL",
				"fs.heading_degrees IS NOT NULL",
				"fs.on_ground IS NOT NULL",
			},
			Forbidden: []string{
				"COALESCE(fs.velocity_mps, 0)",
				"COALESCE(fs.heading_degrees, 0)",
				"COALESCE(fs.on_ground, false)",
			},
			Counts: []fragmentCount{
				{
					Fragment: "fs.velocity_mps IS NOT NULL",
					Minimum:  2,
					Maximum:  2,
				},
				{
					Fragment: "fs.heading_degrees IS NOT NULL",
					Minimum:  2,
					Maximum:  2,
				},
				{
					Fragment: "fs.on_ground IS NOT NULL",
					Minimum:  2,
					Maximum:  2,
				},
			},
		},
		{
			Name: "Airspace excludes incomplete analytical kinematics",
			Path: "apps/api/internal/airspaceintelligence/airspaceproduction/postgres_reader.go",
			Required: []string{
				"fs.velocity_mps IS NOT NULL",
				"fs.heading_degrees IS NOT NULL",
				"fs.vertical_rate_mps IS NOT NULL",
				"fs.on_ground IS NOT NULL",
			},
			Forbidden: []string{
				"COALESCE(fs.velocity_mps, 0)",
				"COALESCE(fs.heading_degrees, 0)",
				"COALESCE(fs.vertical_rate_mps, 0)",
				"COALESCE(fs.on_ground, false)",
			},
		},
		{
			Name: "Traffic validation understands telemetry availability",
			Path: "apps/api/internal/services/traffic/validator/validator.go",
			Required: []string{
				"velocityAvailable := item.HasVelocity()",
				"headingAvailable := item.HasHeading()",
				"verticalRateAvailable := item.HasVerticalRate()",
				"onGroundAvailable := item.HasOnGroundState()",
				`"velocity_mps"`,
				`"heading_degrees"`,
				`"vertical_rate_mps"`,
				`"on_ground"`,
			},
		},
	}
}

func historicalPaginationRules() []fileRule {
	return []fileRule{
		{
			Name: "Historical cursor contains the complete ordering key",
			Path: "apps/api/internal/historicalintelligence/historicalaggregatecontract/pagination.go",
			Required: []string{
				"type ListCursor struct",
				"WindowEnd",
				"WindowStart",
				"AsOfTime",
				"ID",
				"NormalizeListCursor(",
				"MaximumListCursorIdentifierLength",
			},
		},
		{
			Name: "Historical page contract exposes composite cursor only",
			Path: "apps/api/internal/historicalintelligence/historicalaggregatecontract/contracts.go",
			Required: []string{
				"Cursor *ListCursor",
				"NextCursor *ListCursor",
				"cursor := page.NextCursor.Clone()",
			},
			Forbidden: []string{
				"BeforeWindowEnd",
				"NextBeforeWindowEnd",
			},
		},
		{
			Name: "Historical PostgreSQL keyset predicate matches ordering",
			Path: "apps/api/internal/historicalintelligence/historicalaggregate/postgres.go",
			Required: []string{
				"listResultsAfterCursorSQL",
				"window_end_unix_nano < $5",
				"window_end_unix_nano = $5",
				"window_start_unix_nano < $6",
				"window_start_unix_nano = $6",
				"as_of_time_unix_nano < $7",
				"as_of_time_unix_nano = $7",
				"id > $8",
				"LIMIT $9",
				"cursor.WindowEnd.UnixNano()",
				"cursor.WindowStart.UnixNano()",
				"cursor.AsOfTime.UnixNano()",
				"cursor.ID",
				"nextCursor = listCursorFromRecord(",
			},
			Counts: []fragmentCount{
				{
					Fragment: "window_end_unix_nano DESC",
					Minimum:  3,
					Maximum:  -1,
				},
				{
					Fragment: "window_start_unix_nano DESC",
					Minimum:  3,
					Maximum:  -1,
				},
				{
					Fragment: "as_of_time_unix_nano DESC",
					Minimum:  3,
					Maximum:  -1,
				},
				{
					Fragment: "id ASC",
					Minimum:  3,
					Maximum:  -1,
				},
			},
		},
		{
			Name: "Historical HTTP cursor is strict and opaque",
			Path: "apps/api/internal/http/historicalcursor/codec.go",
			Required: []string{
				`"historical-aggregate-cursor-v1"`,
				"base64.RawURLEncoding",
				"decoder.DisallowUnknownFields()",
				"rejectTrailingJSON(",
				"NormalizeListCursor(",
			},
		},
		{
			Name: "Historical HTTP handler accepts opaque cursor",
			Path: "apps/api/internal/http/handlers/historical_intelligence.go",
			Required: []string{
				`historicalCursorQuery`,
				`= "cursor"`,
				"parseHistoricalCursor(",
				"historicalcursor.Decode(",
			},
			Forbidden: []string{
				"BeforeWindowEnd",
				"before_window_end",
			},
		},
		{
			Name: "Historical DTO returns opaque next cursor",
			Path: "apps/api/internal/http/dto/historical_intelligence.go",
			Required: []string{
				`json:"next_cursor,omitempty"`,
				"NextCursor string",
				"historicalcursor.Encode(",
			},
			Forbidden: []string{
				"NextBeforeWindowEnd",
				"next_before_window_end",
			},
		},
	}
}

func weatherCompositionRules() []fileRule {
	return []fileRule{
		{
			Name: "Weather route remains a narrow coordinator",
			Path: "apps/api/internal/server/weather_route.go",
			Required: []string{
				"if openMeteoTimeout <= 0",
				`"open-meteo timeout must be greater than zero"`,
				"composeWeatherRouteDependencies(",
				"registerCurrentWeatherRoute(",
			},
			Forbidden: []string{
				"openmeteo",
				"providerbudget",
				"providerresponse",
				"weatherprovider",
				"repository/postgres",
				"services/weather",
				"NewWeatherRepository",
				"NewWeatherHandler",
				"v1.Get(",
			},
		},
		{
			Name: "Weather provider composition owns provider runtime only",
			Path: "apps/api/internal/server/weather_provider_composition.go",
			Required: []string{
				"providerbudget.New(",
				"providerresponse.New(",
				"NewIntegrationObserver(",
				"NewDefault[",
				"openmeteo.New(",
				"weatherprovider.New(",
			},
			Forbidden: []string{
				"NewWeatherRepository",
				"NewWeatherHandler",
				"v1.Get(",
			},
		},
		{
			Name: "Weather application composition owns repository service handler",
			Path: "apps/api/internal/server/weather_application_composition.go",
			Required: []string{
				"postgres.NewWeatherRepository(",
				"weatherservice.New(",
				"handlers.NewWeatherHandler(",
			},
			Forbidden: []string{
				"providerbudget.New(",
				"providerresponse.New(",
				"openmeteo.New(",
				"weatherprovider.New(",
				"v1.Get(",
			},
		},
		{
			Name: "Weather registration owns the HTTP boundary only",
			Path: "apps/api/internal/server/weather_route_registration.go",
			Required: []string{
				`CurrentWeatherPath = "/weather/current"`,
				"v1.Get(",
				"handler.GetCurrent",
			},
			Forbidden: []string{
				"providerbudget.New(",
				"providerresponse.New(",
				"openmeteo.New(",
				"weatherprovider.New(",
				"NewWeatherRepository",
				"weatherservice.New(",
			},
		},
	}
}

func evidenceRules() []fileRule {
	return []fileRule{
		{
			Name: "Projection snapshot regression evidence exists",
			Path: "apps/api/internal/projectionintelligence/projectionread/snapshot_architecture_test.go",
			Required: []string{
				"TestProjectionReadDataSourceExposesOneSnapshotOperation",
				"TestProjectionReadServiceDoesNotCoordinateIndependentDatabaseReads",
				"TestProductionPostgresDataSourceUsesRepeatableReadSnapshotExecutor",
				"LoadSnapshot",
				"repeatableReadSnapshotExecutor",
			},
		},
		{
			Name: "Nullable telemetry regression evidence exists",
			Path: "apps/api/internal/projectionintelligence/projectionread/nullable_telemetry_architecture_test.go",
			Required: []string{
				"TestProjectionReadNullableTelemetryBoundaryRemainsExplicit",
				"COALESCE(latitude, 0)",
				"var latitude pgtype.Float8",
				"completeRequiredTelemetry",
			},
		},
		{
			Name: "Historical pagination regression evidence exists",
			Path: "apps/api/internal/historicalintelligence/historicalaggregate/composite_pagination_architecture_test.go",
			Required: []string{
				"window_end_unix_nano DESC",
				"window_start_unix_nano DESC",
				"as_of_time_unix_nano DESC",
				"id ASC",
			},
		},
		{
			Name: "Weather composition regression evidence exists",
			Path: "apps/api/internal/server/weather_composition_architecture_test.go",
			Required: []string{
				"WeatherRouteFileRemainsCoordinatorOnly",
				"WeatherCompositionResponsibilitiesRemainSeparated",
			},
		},
		{
			Name: "Final verification script executes both audits",
			Path: "scripts/verify-backend-final-correctness.sh",
			Required: []string{
				"go run ./tools/backendfinalaudit -strict",
				"go run ./tools/projectaudit -mode all -strict",
				"go test ./...",
				"go vet ./...",
				"go build ./cmd/...",
			},
		},
	}
}
