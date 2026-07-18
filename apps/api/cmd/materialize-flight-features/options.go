package main

import (
	"flag"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var materializationICAO24Pattern = regexp.MustCompile(
	`^[A-F0-9]{6}$`,
)

type commandOptions struct {
	TrajectoryID string
	ICAO24       string
	AsOfTime     time.Time
}

func parseCommandOptions(
	args []string,
	stderr io.Writer,
) (commandOptions, error) {
	flags := flag.NewFlagSet(
		"materialize-flight-features",
		flag.ContinueOnError,
	)
	flags.SetOutput(stderr)

	trajectoryIDValue := flags.String(
		"trajectory-id",
		"",
		"existing flight trajectory UUID",
	)
	icao24Value := flags.String(
		"icao24",
		"",
		"latest persisted trajectory for a six-character ICAO24 address",
	)
	asOfTimeValue := flags.String(
		"as-of-time",
		"",
		"optional RFC 3339 evidence cutoff; defaults to trajectory end time",
	)

	if err := flags.Parse(args); err != nil {
		return commandOptions{}, err
	}
	if flags.NArg() != 0 {
		return commandOptions{}, fmt.Errorf(
			"unexpected positional arguments: %v",
			flags.Args(),
		)
	}

	trajectoryID := strings.TrimSpace(*trajectoryIDValue)
	icao24 := strings.ToUpper(strings.TrimSpace(*icao24Value))
	if (trajectoryID == "") == (icao24 == "") {
		return commandOptions{}, fmt.Errorf(
			"exactly one of --trajectory-id or --icao24 is required",
		)
	}

	if trajectoryID != "" {
		parsed, err := uuid.Parse(trajectoryID)
		if err != nil {
			return commandOptions{}, fmt.Errorf(
				"trajectory identifier must be a valid UUID",
			)
		}
		trajectoryID = strings.ToLower(parsed.String())
	}
	if icao24 != "" && !materializationICAO24Pattern.MatchString(icao24) {
		return commandOptions{}, fmt.Errorf(
			"ICAO24 must contain exactly six hexadecimal characters",
		)
	}

	asOfTime := time.Time{}
	if normalized := strings.TrimSpace(*asOfTimeValue); normalized != "" {
		parsed, err := time.Parse(time.RFC3339Nano, normalized)
		if err != nil {
			return commandOptions{}, fmt.Errorf(
				"as-of time must be a valid RFC 3339 timestamp",
			)
		}
		asOfTime = parsed.UTC()
	}

	return commandOptions{
		TrajectoryID: trajectoryID,
		ICAO24:       icao24,
		AsOfTime:     asOfTime,
	}, nil
}
