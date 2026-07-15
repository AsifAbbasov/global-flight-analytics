package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/config"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/database"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalaggregate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalmaterialization"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalread"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalreplay"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load(".env")
	_ = godotenv.Load("apps/api/.env")

	os.Exit(
		runCommand(
			os.Args[1:],
			os.Stdout,
			os.Stderr,
			time.Now,
		),
	)
}

func runCommand(
	args []string,
	stdout io.Writer,
	stderr io.Writer,
	now func() time.Time,
) int {
	if now == nil {
		fmt.Fprintln(
			stderr,
			"ERROR: command clock is required",
		)
		return 1
	}

	options, err := parseCommandOptions(
		args,
		stderr,
		now().UTC(),
	)
	if errors.Is(err, flag.ErrHelp) {
		return 0
	}
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: parse command options: %v\n",
			err,
		)
		return 1
	}

	cfg, err :=
		config.LoadHistoricalMaterializationConfig()
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: load Historical Materialization configuration: %v\n",
			err,
		)
		return 1
	}

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

	readRepository, err := historicalread.NewPostgres(
		historicalread.PostgresConfig{
			Pool: pool,
		},
	)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: compose Historical Read Repository: %v\n",
			err,
		)
		return 1
	}

	aggregateStore, err :=
		historicalaggregate.NewPostgres(
			historicalaggregate.PostgresConfig{
				Pool: pool,
			},
		)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: compose Historical Aggregate Store: %v\n",
			err,
		)
		return 1
	}

	materializer, err :=
		historicalmaterialization.New(
			historicalmaterialization.Config{
				Repository: readRepository,
				Store:      aggregateStore,
				Now:        now,
			},
		)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: compose Historical Materializer: %v\n",
			err,
		)
		return 1
	}

	replayRunner, err := historicalreplay.New(
		historicalreplay.Config{
			Materializer: materializer,
			Now:          now,
		},
	)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: compose Historical Replay Runner: %v\n",
			err,
		)
		return 1
	}

	operation, err := newCommandOperation(
		materializer,
		replayRunner,
		now,
	)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: compose Historical Intelligence command: %v\n",
			err,
		)
		return 1
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		cfg.OperationTimeout,
	)
	defer cancel()

	report, err := operation.Execute(
		ctx,
		options,
	)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: execute Historical Intelligence %s: %v\n",
			options.Mode,
			err,
		)
		return 1
	}

	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: encode command report: %v\n",
			err,
		)
		return 1
	}

	return 0
}
