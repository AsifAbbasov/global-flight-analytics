package main

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalread"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalreplay"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalwindow"
)

func TestParseCommandOptionsMaterializeGlobal(
	t *testing.T,
) {
	now := commandTestTime()
	options, err := parseCommandOptions(
		[]string{
			"-metric", "flight_count",
			"-scope", "global",
			"-granularity", "hour",
			"-start", "2026-07-14T08:00:00Z",
			"-end", "2026-07-14T10:00:00Z",
			"-as-of", "2026-07-15T10:00:00Z",
		},
		&bytes.Buffer{},
		now,
	)
	if err != nil {
		t.Fatalf(
			"parse materialize options: %v",
			err,
		)
	}

	if options.Mode !=
		operationModeMaterialize ||
		options.MetricName !=
			historicalcontract.
				MetricNameFlightCount ||
		options.Scope.Type !=
			historicalcontract.ScopeTypeGlobal ||
		options.Granularity !=
			historicalcontract.GranularityHour ||
		options.DatasetLimit !=
			historicalread.DefaultDatasetLimit ||
		options.MaximumBucketCount !=
			historicalwindow.
				DefaultMaximumBucketCount ||
		options.MaximumWindowCount !=
			historicalreplay.
				DefaultMaximumWindowCount {
		t.Fatalf(
			"unexpected materialize options: %#v",
			options,
		)
	}
}

func TestParseCommandOptionsReplayRouteNormalizesCodes(
	t *testing.T,
) {
	options, err := parseCommandOptions(
		[]string{
			"-mode", "replay",
			"-metric", "route_observations",
			"-scope", "route",
			"-origin", "ubbb",
			"-destination", "ugtb",
			"-granularity", "day",
			"-start", "2026-07-01T00:00:00+04:00",
			"-end", "2026-07-05T00:00:00+04:00",
			"-as-of", "2026-07-15T12:00:00+04:00",
			"-dataset-limit", "5000",
			"-max-buckets", "50",
			"-max-windows", "10",
		},
		&bytes.Buffer{},
		commandTestTime(),
	)
	if err != nil {
		t.Fatalf(
			"parse replay options: %v",
			err,
		)
	}

	if options.Mode != operationModeReplay ||
		options.Scope.Type !=
			historicalcontract.ScopeTypeRoute ||
		options.Scope.OriginICAOCode != "UBBB" ||
		options.Scope.DestinationICAOCode !=
			"UGTB" ||
		options.DatasetLimit != 5000 ||
		options.MaximumBucketCount != 50 ||
		options.MaximumWindowCount != 10 {
		t.Fatalf(
			"unexpected replay options: %#v",
			options,
		)
	}

	expectedStart := time.Date(
		2026,
		time.June,
		30,
		20,
		0,
		0,
		0,
		time.UTC,
	)
	if !options.StartTime.Equal(expectedStart) {
		t.Fatalf(
			"start time = %s, want %s",
			options.StartTime,
			expectedStart,
		)
	}
}

func TestParseCommandOptionsRejectsUnsafeOrUnsupportedInputs(
	t *testing.T,
) {
	base := []string{
		"-metric", "flight_count",
		"-scope", "global",
		"-granularity", "hour",
		"-start", "2026-07-14T08:00:00Z",
		"-end", "2026-07-14T10:00:00Z",
		"-as-of", "2026-07-15T10:00:00Z",
	}

	tests := []struct {
		name          string
		args          []string
		errorFragment string
	}{
		{
			name: "as-of required",
			args: replaceFlagValue(
				base,
				"-as-of",
				"",
			),
			errorFragment: "as-of time is required",
		},
		{
			name: "end after as-of",
			args: replaceFlagValue(
				base,
				"-as-of",
				"2026-07-14T09:00:00Z",
			),
			errorFragment: "end time must not be after as-of time",
		},
		{
			name: "future as-of",
			args: replaceFlagValue(
				base,
				"-as-of",
				"2026-07-16T10:00:00Z",
			),
			errorFragment: "as-of time must not be in the future",
		},
		{
			name: "custom granularity rejected",
			args: replaceFlagValue(
				base,
				"-granularity",
				"custom",
			),
			errorFragment: "granularity must be hour, day, or week",
		},
		{
			name: "unmaterializable metric rejected",
			args: replaceFlagValue(
				base,
				"-metric",
				"data_freshness",
			),
			errorFragment: "metric is required and must be materializable",
		},
		{
			name: "traffic metric airport mismatch",
			args: []string{
				"-metric", "flight_count",
				"-scope", "airport",
				"-airport", "UBBB",
				"-granularity", "hour",
				"-start", "2026-07-14T08:00:00Z",
				"-end", "2026-07-14T10:00:00Z",
				"-as-of", "2026-07-15T10:00:00Z",
			},
			errorFragment: "traffic metrics require global scope",
		},
		{
			name: "invalid airport ICAO",
			args: []string{
				"-metric", "airport_departures",
				"-scope", "airport",
				"-airport", "BAD",
				"-granularity", "hour",
				"-start", "2026-07-14T08:00:00Z",
				"-end", "2026-07-14T10:00:00Z",
				"-as-of", "2026-07-15T10:00:00Z",
			},
			errorFragment: "airport scope requires exactly one valid four-character airport ICAO code",
		},
		{
			name: "dataset limit bounded",
			args: append(
				append(
					[]string(nil),
					base...,
				),
				"-dataset-limit",
				"100001",
			),
			errorFragment: "dataset limit must be between 1 and 100000",
		},
		{
			name: "window limit bounded",
			args: append(
				append(
					[]string(nil),
					base...,
				),
				"-max-windows",
				"10001",
			),
			errorFragment: "maximum window count must be between 1 and 10000",
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				_, err := parseCommandOptions(
					test.args,
					&bytes.Buffer{},
					commandTestTime(),
				)
				if err == nil {
					t.Fatal(
						"expected option validation error",
					)
				}
				if !strings.Contains(
					err.Error(),
					test.errorFragment,
				) {
					t.Fatalf(
						"error = %q, want fragment %q",
						err.Error(),
						test.errorFragment,
					)
				}
			},
		)
	}
}

func TestParseCommandOptionsHelp(
	t *testing.T,
) {
	var output bytes.Buffer
	_, err := parseCommandOptions(
		[]string{"-help"},
		&output,
		commandTestTime(),
	)
	if err == nil {
		t.Fatal(
			"expected flag help sentinel",
		)
	}
	if !strings.Contains(
		output.String(),
		"Usage: materialize-historical-intelligence",
	) {
		t.Fatalf(
			"help output = %q",
			output.String(),
		)
	}
}

func commandTestTime() time.Time {
	return time.Date(
		2026,
		time.July,
		15,
		12,
		0,
		0,
		0,
		time.UTC,
	)
}

func replaceFlagValue(
	args []string,
	name string,
	value string,
) []string {
	result := append(
		[]string(nil),
		args...,
	)
	for index := 0; index+1 < len(result); index++ {
		if result[index] == name {
			result[index+1] = value
			return result
		}
	}

	return result
}
