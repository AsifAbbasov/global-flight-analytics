package historicalwindow

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
)

func FloorBoundary(
	value time.Time,
	granularity historicalcontract.Granularity,
) (time.Time, error) {
	if value.IsZero() {
		return time.Time{}, ErrStartTimeRequired
	}

	normalized := value.UTC()

	switch granularity {
	case historicalcontract.GranularityHour:
		return normalized.Truncate(time.Hour), nil

	case historicalcontract.GranularityDay:
		return time.Date(
			normalized.Year(),
			normalized.Month(),
			normalized.Day(),
			0,
			0,
			0,
			0,
			time.UTC,
		), nil

	case historicalcontract.GranularityWeek:
		dayBoundary := time.Date(
			normalized.Year(),
			normalized.Month(),
			normalized.Day(),
			0,
			0,
			0,
			0,
			time.UTC,
		)
		daysSinceMonday := (int(dayBoundary.Weekday()) -
			int(time.Monday) +
			7) % 7

		return dayBoundary.AddDate(
			0,
			0,
			-daysSinceMonday,
		), nil

	case historicalcontract.GranularityCustom:
		return normalized, nil

	default:
		return time.Time{},
			ErrUnsupportedGranularity
	}
}

func CeilBoundary(
	value time.Time,
	granularity historicalcontract.Granularity,
) (time.Time, error) {
	floor, err := FloorBoundary(
		value,
		granularity,
	)
	if err != nil {
		return time.Time{}, err
	}

	normalized := value.UTC()
	if floor.Equal(normalized) {
		return floor, nil
	}

	return NextBoundary(floor, granularity)
}

func NextBoundary(
	value time.Time,
	granularity historicalcontract.Granularity,
) (time.Time, error) {
	if value.IsZero() {
		return time.Time{}, ErrStartTimeRequired
	}

	normalized := value.UTC()

	switch granularity {
	case historicalcontract.GranularityHour:
		return normalized.Add(time.Hour), nil

	case historicalcontract.GranularityDay:
		return normalized.AddDate(0, 0, 1), nil

	case historicalcontract.GranularityWeek:
		return normalized.AddDate(0, 0, 7), nil

	case historicalcontract.GranularityCustom:
		return time.Time{},
			ErrUnsupportedGranularity

	default:
		return time.Time{},
			ErrUnsupportedGranularity
	}
}

func boundaryDuration(
	granularity historicalcontract.Granularity,
) (time.Duration, error) {
	switch granularity {
	case historicalcontract.GranularityHour:
		return time.Hour, nil
	case historicalcontract.GranularityDay:
		return 24 * time.Hour, nil
	case historicalcontract.GranularityWeek:
		return 7 * 24 * time.Hour, nil
	default:
		return 0, ErrUnsupportedGranularity
	}
}

func isSupportedGranularity(
	granularity historicalcontract.Granularity,
) bool {
	switch granularity {
	case historicalcontract.GranularityHour,
		historicalcontract.GranularityDay,
		historicalcontract.GranularityWeek,
		historicalcontract.GranularityCustom:
		return true
	default:
		return false
	}
}
