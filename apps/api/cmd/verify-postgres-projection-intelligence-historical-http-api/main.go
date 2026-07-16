package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/config"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/database"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionproduction"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionread"
	"github.com/joho/godotenv"
)

func main() {
	os.Exit(
		run(
			os.Stdout,
			os.Stderr,
		),
	)
}

func run(
	stdout io.Writer,
	stderr io.Writer,
) int {
	_ = godotenv.Load(".env")
	_ = godotenv.Load("apps/api/.env")

	cfg, err := config.LoadMigrationConfig()
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: load database configuration: %v\n",
			err,
		)
		return 1
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		cfg.MigrationTimeout,
	)
	defer cancel()

	pool, err := database.NewPostgresPool(
		cfg.Database.URL,
		cfg.Database.ConnectTimeout,
	)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: connect PostgreSQL: %v\n",
			err,
		)
		return 1
	}
	defer pool.Close()

	if err := verifySchema(
		ctx,
		pool,
	); err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: verify runtime schema: %v\n",
			err,
		)
		return 1
	}

	schedule, err :=
		buildVerificationSchedule(
			time.Now().UTC(),
		)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: build verification schedule: %v\n",
			err,
		)
		return 1
	}

	if err := cleanupFixture(
		ctx,
		pool,
	); err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: remove stale historical fixture: %v\n",
			err,
		)
		return 1
	}

	cleanupPending := true
	defer func() {
		if !cleanupPending {
			return
		}

		cleanupContext, cleanupCancel :=
			context.WithTimeout(
				context.Background(),
				30*time.Second,
			)
		defer cleanupCancel()

		if cleanupErr := cleanupFixture(
			cleanupContext,
			pool,
		); cleanupErr != nil {
			fmt.Fprintf(
				stderr,
				"ERROR: deferred historical fixture cleanup failed: %v\n",
				cleanupErr,
			)
		}
	}()

	if err := insertFixture(
		ctx,
		pool,
		schedule,
	); err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: insert historical runtime fixture: %v\n",
			err,
		)
		return 1
	}

	beforeCounts, err :=
		loadFixtureCounts(
			ctx,
			pool,
		)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: count inserted historical fixture: %v\n",
			err,
		)
		return 1
	}
	if beforeCounts != expectedFixtureCounts() {
		fmt.Fprintf(
			stderr,
			"ERROR: unexpected inserted fixture counts: %#v\n",
			beforeCounts,
		)
		return 1
	}

	composition, err :=
		projectionread.NewPostgres(
			projectionread.PostgresConfig{
				Pool: pool,
				Policy: projectionread.
					DefaultPolicy(),
				Now: func() time.Time {
					return schedule.GeneratedAt
				},
			},
		)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: compose production Projection Intelligence reader: %v\n",
			err,
		)
		return 1
	}

	app, err := buildRuntimeApp(
		composition.Service,
	)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: build runtime HTTP application: %v\n",
			err,
		)
		return 1
	}

	payload, err :=
		verifyHistoricalEndpoint(
			app,
			schedule,
		)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: verify Historical Neighbor Continuation endpoint: %v\n",
			err,
		)
		return 1
	}

	if err := cleanupFixture(
		ctx,
		pool,
	); err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: clean historical runtime fixture: %v\n",
			err,
		)
		return 1
	}

	afterCounts, err :=
		loadFixtureCounts(
			ctx,
			pool,
		)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: count historical fixture after cleanup: %v\n",
			err,
		)
		return 1
	}
	if afterCounts != (fixtureCounts{}) {
		fmt.Fprintf(
			stderr,
			"ERROR: historical runtime fixture remained after cleanup: %#v\n",
			afterCounts,
		)
		return 1
	}

	cleanupPending = false

	fmt.Fprintln(
		stdout,
		"PostgreSQL Projection Intelligence Historical HTTP Verification",
	)
	fmt.Fprintf(
		stdout,
		"Production composition: %s\n",
		projectionproduction.Version,
	)
	fmt.Fprintf(
		stdout,
		"Projection method: %s\n",
		payload.Data.Projection.Method.Name,
	)
	fmt.Fprintf(
		stdout,
		"Historical neighbors: %d\n",
		len(
			payload.Data.Evidence.
				NeighborSelection.Neighbors,
		),
	)
	fmt.Fprintf(
		stdout,
		"Forecast points: %d\n",
		len(
			payload.Data.Projection.Points,
		),
	)
	fmt.Fprintf(
		stdout,
		"Arrival airport: %s\n",
		payload.Data.Projection.Arrival.
			AirportICAOCode,
	)
	fmt.Fprintln(
		stdout,
		"Schema objects: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Deterministic multi-flight fixture: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Route Intelligence history loading: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Historical candidate loading: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Historical Neighbor Selection: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Pattern Confidence: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Pattern Freshness Guard: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Low-Frequency Route Guard: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Historical Neighbor Continuation: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Estimated Arrival attachment: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Projection HTTP contract: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Fixture cleanup: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Persistent verification rows: 0",
	)
	fmt.Fprintln(
		stdout,
		"Result: PASS",
	)

	return 0
}
