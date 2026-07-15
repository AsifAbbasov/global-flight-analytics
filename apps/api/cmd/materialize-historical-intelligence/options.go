package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalread"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalreplay"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalwindow"
)

var commandICAOPattern = regexp.MustCompile(
	`^[A-Z0-9]{4}$`,
)

func parseCommandOptions(
	args []string,
	output io.Writer,
	now time.Time,
) (commandOptions, error) {
	flagSet := flag.NewFlagSet(
		"materialize-historical-intelligence",
		flag.ContinueOnError,
	)
	flagSet.SetOutput(output)

	modeValue := flagSet.String(
		"mode",
		string(operationModeMaterialize),
		"operation mode: materialize or replay",
	)
	metricValue := flagSet.String(
		"metric",
		"",
		"Historical Intelligence metric name",
	)
	scopeValue := flagSet.String(
		"scope",
		"",
		"scope: global, airport, or route",
	)
	granularityValue := flagSet.String(
		"granularity",
		"",
		"granularity: hour, day, or week",
	)
	startValue := flagSet.String(
		"start",
		"",
		"inclusive RFC 3339 start time",
	)
	endValue := flagSet.String(
		"end",
		"",
		"exclusive RFC 3339 end time",
	)
	asOfValue := flagSet.String(
		"as-of",
		"",
		"required RFC 3339 knowledge cutoff",
	)
	airportValue := flagSet.String(
		"airport",
		"",
		"airport ICAO code for airport scope",
	)
	originValue := flagSet.String(
		"origin",
		"",
		"origin airport ICAO code for route scope",
	)
	destinationValue := flagSet.String(
		"destination",
		"",
		"destination airport ICAO code for route scope",
	)
	datasetLimit := flagSet.Int(
		"dataset-limit",
		historicalread.DefaultDatasetLimit,
		"maximum rows read from each historical source dataset",
	)
	maximumBucketCount := flagSet.Int(
		"max-buckets",
		historicalwindow.DefaultMaximumBucketCount,
		"maximum analytical buckets allowed in one plan",
	)
	maximumWindowCount := flagSet.Int(
		"max-windows",
		historicalreplay.DefaultMaximumWindowCount,
		"maximum replay windows allowed",
	)

	flagSet.Usage = func() {
		fmt.Fprintln(
			output,
			"Usage: materialize-historical-intelligence [flags]",
		)
		flagSet.PrintDefaults()
	}

	if err := flagSet.Parse(args); err != nil {
		return commandOptions{}, err
	}
	if flagSet.NArg() != 0 {
		return commandOptions{},
			fmt.Errorf(
				"unexpected positional arguments: %s",
				strings.Join(flagSet.Args(), " "),
			)
	}

	mode, err := parseOperationMode(
		*modeValue,
	)
	if err != nil {
		return commandOptions{}, err
	}
	metricName, err := parseMetricName(
		*metricValue,
	)
	if err != nil {
		return commandOptions{}, err
	}
	granularity, err := parseGranularity(
		*granularityValue,
	)
	if err != nil {
		return commandOptions{}, err
	}
	startTime, err := parseRequiredTime(
		"start",
		*startValue,
	)
	if err != nil {
		return commandOptions{}, err
	}
	endTime, err := parseRequiredTime(
		"end",
		*endValue,
	)
	if err != nil {
		return commandOptions{}, err
	}
	asOfTime, err := parseRequiredTime(
		"as-of",
		*asOfValue,
	)
	if err != nil {
		return commandOptions{}, err
	}

	now = now.UTC()
	if now.IsZero() {
		return commandOptions{},
			errors.New(
				"current command time is required",
			)
	}
	if !startTime.Before(endTime) {
		return commandOptions{},
			errors.New(
				"start time must be before end time",
			)
	}
	if endTime.After(asOfTime) {
		return commandOptions{},
			errors.New(
				"end time must not be after as-of time",
			)
	}
	if asOfTime.After(now) {
		return commandOptions{},
			errors.New(
				"as-of time must not be in the future",
			)
	}

	scope, err := parseCommandScope(
		*scopeValue,
		*airportValue,
		*originValue,
		*destinationValue,
	)
	if err != nil {
		return commandOptions{}, err
	}
	if err := validateMetricScope(
		metricName,
		scope,
	); err != nil {
		return commandOptions{}, err
	}

	if *datasetLimit < 1 ||
		*datasetLimit >
			historicalread.MaximumDatasetLimit {
		return commandOptions{},
			fmt.Errorf(
				"dataset limit must be between 1 and %d",
				historicalread.MaximumDatasetLimit,
			)
	}
	if *maximumBucketCount < 1 ||
		*maximumBucketCount >
			historicalwindow.MaximumBucketCount {
		return commandOptions{},
			fmt.Errorf(
				"maximum bucket count must be between 1 and %d",
				historicalwindow.MaximumBucketCount,
			)
	}
	if *maximumWindowCount < 1 ||
		*maximumWindowCount >
			historicalreplay.MaximumWindowCount {
		return commandOptions{},
			fmt.Errorf(
				"maximum window count must be between 1 and %d",
				historicalreplay.MaximumWindowCount,
			)
	}

	return commandOptions{
		Mode: mode,

		StartTime: startTime,
		EndTime:   endTime,
		AsOfTime:  asOfTime,

		Granularity: granularity,
		MetricName:  metricName,
		Scope:       scope,

		DatasetLimit:       *datasetLimit,
		MaximumBucketCount: *maximumBucketCount,
		MaximumWindowCount: *maximumWindowCount,
	}, nil
}

func parseOperationMode(
	value string,
) (operationMode, error) {
	switch operationMode(
		strings.ToLower(
			strings.TrimSpace(value),
		),
	) {
	case operationModeMaterialize:
		return operationModeMaterialize, nil
	case operationModeReplay:
		return operationModeReplay, nil
	default:
		return "",
			errors.New(
				"mode must be materialize or replay",
			)
	}
}

func parseMetricName(
	value string,
) (historicalcontract.MetricName, error) {
	normalized := historicalcontract.MetricName(
		strings.ToLower(
			strings.TrimSpace(value),
		),
	)
	switch normalized {
	case historicalcontract.MetricNameActiveAircraft,
		historicalcontract.MetricNameFlightCount,
		historicalcontract.MetricNameTrajectoryCount,
		historicalcontract.MetricNameObservationCount,
		historicalcontract.MetricNameTrafficDensity,
		historicalcontract.MetricNameAirportDepartures,
		historicalcontract.MetricNameAirportArrivals,
		historicalcontract.MetricNameAirportOperations,
		historicalcontract.MetricNameUniqueAircraft,
		historicalcontract.MetricNameActiveRoutes,
		historicalcontract.MetricNameRouteObservations,
		historicalcontract.MetricNameRouteConfidence,
		historicalcontract.MetricNameCompleteRouteRatio,
		historicalcontract.MetricNamePartialRouteRatio,
		historicalcontract.MetricNameUnavailableRouteRatio,
		historicalcontract.MetricNameGreatCircleDistanceKM:
		return normalized, nil
	default:
		return "",
			errors.New(
				"metric is required and must be materializable",
			)
	}
}

func parseGranularity(
	value string,
) (historicalcontract.Granularity, error) {
	normalized := historicalcontract.Granularity(
		strings.ToLower(
			strings.TrimSpace(value),
		),
	)
	switch normalized {
	case historicalcontract.GranularityHour,
		historicalcontract.GranularityDay,
		historicalcontract.GranularityWeek:
		return normalized, nil
	default:
		return "",
			errors.New(
				"granularity must be hour, day, or week",
			)
	}
}

func parseRequiredTime(
	name string,
	value string,
) (time.Time, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return time.Time{},
			fmt.Errorf(
				"%s time is required",
				name,
			)
	}

	parsed, err := time.Parse(
		time.RFC3339Nano,
		normalized,
	)
	if err != nil {
		return time.Time{},
			fmt.Errorf(
				"%s time must be valid RFC 3339: %w",
				name,
				err,
			)
	}

	return parsed.UTC(), nil
}

func parseCommandScope(
	scopeValue string,
	airportValue string,
	originValue string,
	destinationValue string,
) (historicalcontract.Scope, error) {
	scopeType := historicalcontract.ScopeType(
		strings.ToLower(
			strings.TrimSpace(scopeValue),
		),
	)
	airportICAO := strings.ToUpper(
		strings.TrimSpace(airportValue),
	)
	originICAO := strings.ToUpper(
		strings.TrimSpace(originValue),
	)
	destinationICAO := strings.ToUpper(
		strings.TrimSpace(destinationValue),
	)

	switch scopeType {
	case historicalcontract.ScopeTypeGlobal:
		if airportICAO != "" ||
			originICAO != "" ||
			destinationICAO != "" {
			return historicalcontract.Scope{},
				errors.New(
					"global scope must not include airport or route identifiers",
				)
		}

	case historicalcontract.ScopeTypeAirport:
		if !commandICAOPattern.MatchString(
			airportICAO,
		) ||
			originICAO != "" ||
			destinationICAO != "" {
			return historicalcontract.Scope{},
				errors.New(
					"airport scope requires exactly one valid four-character airport ICAO code",
				)
		}

	case historicalcontract.ScopeTypeRoute:
		if !commandICAOPattern.MatchString(
			originICAO,
		) ||
			!commandICAOPattern.MatchString(
				destinationICAO,
			) ||
			airportICAO != "" {
			return historicalcontract.Scope{},
				errors.New(
					"route scope requires valid origin and destination ICAO codes",
				)
		}

	default:
		return historicalcontract.Scope{},
			errors.New(
				"scope must be global, airport, or route",
			)
	}

	return historicalcontract.Scope{
		Type:                scopeType,
		AirportICAOCode:     airportICAO,
		OriginICAOCode:      originICAO,
		DestinationICAOCode: destinationICAO,
	}, nil
}

func validateMetricScope(
	metricName historicalcontract.MetricName,
	scope historicalcontract.Scope,
) error {
	switch metricName {
	case historicalcontract.MetricNameActiveAircraft,
		historicalcontract.MetricNameFlightCount,
		historicalcontract.MetricNameTrajectoryCount,
		historicalcontract.MetricNameObservationCount,
		historicalcontract.MetricNameTrafficDensity:
		if scope.Type !=
			historicalcontract.ScopeTypeGlobal {
			return errors.New(
				"traffic metrics require global scope",
			)
		}

	case historicalcontract.MetricNameAirportDepartures,
		historicalcontract.MetricNameAirportArrivals,
		historicalcontract.MetricNameAirportOperations,
		historicalcontract.MetricNameUniqueAircraft:
		if scope.Type !=
			historicalcontract.ScopeTypeAirport {
			return errors.New(
				"airport metrics require airport scope",
			)
		}

	case historicalcontract.MetricNameActiveRoutes,
		historicalcontract.MetricNameRouteObservations,
		historicalcontract.MetricNameRouteConfidence,
		historicalcontract.MetricNameCompleteRouteRatio,
		historicalcontract.MetricNamePartialRouteRatio,
		historicalcontract.MetricNameUnavailableRouteRatio,
		historicalcontract.MetricNameGreatCircleDistanceKM:
		if scope.Type !=
			historicalcontract.ScopeTypeGlobal &&
			scope.Type !=
				historicalcontract.ScopeTypeRoute {
			return errors.New(
				"route metrics require global or route scope",
			)
		}

	default:
		return errors.New(
			"metric is not materializable",
		)
	}

	return nil
}
