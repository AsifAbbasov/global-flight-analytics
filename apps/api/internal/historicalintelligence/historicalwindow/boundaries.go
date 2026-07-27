package historicalwindow

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
)

type granularityPolicy struct {
	floor func(time.Time) time.Time
	shift func(time.Time, int) time.Time
}

func granularityPolicyFor(
	granularity historicalcontract.Granularity,
) (granularityPolicy, bool) {
	switch granularity {
	case historicalcontract.GranularityHour:
		return granularityPolicy{
			floor: func(value time.Time) time.Time {
				return value.Truncate(time.Hour)
			},
			shift: func(value time.Time, steps int) time.Time {
				return value.Add(
					time.Duration(steps) * time.Hour,
				)
			},
		}, true

	case historicalcontract.GranularityDay:
		return granularityPolicy{
			floor: func(value time.Time) time.Time {
				return time.Date(
					value.Year(),
					value.Month(),
					value.Day(),
					0,
					0,
					0,
					0,
					time.UTC,
				)
			},
			shift: func(value time.Time, steps int) time.Time {
				return value.AddDate(0, 0, steps)
			},
		}, true

	case historicalcontract.GranularityWeek:
		return granularityPolicy{
			floor: func(value time.Time) time.Time {
				dayBoundary := time.Date(
					value.Year(),
					value.Month(),
					value.Day(),
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
				)
			},
			shift: func(value time.Time, steps int) time.Time {
				return value.AddDate(0, 0, 7*steps)
			},
		}, true

	case historicalcontract.GranularityCustom:
		return granularityPolicy{
			floor: func(value time.Time) time.Time {
				return value
			},
		}, true

	default:
		return granularityPolicy{}, false
	}
}

func FloorBoundary(
	value time.Time,
	granularity historicalcontract.Granularity,
) (time.Time, error) {
	if value.IsZero() {
		return time.Time{}, ErrStartTimeRequired
	}

	policy, exists := granularityPolicyFor(granularity)
	if !exists {
		return time.Time{}, ErrUnsupportedGranularity
	}

	return policy.floor(value.UTC()), nil
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

	policy, exists := granularityPolicyFor(granularity)
	if !exists || policy.shift == nil {
		return time.Time{}, ErrUnsupportedGranularity
	}

	normalized := value.UTC()
	floor := policy.floor(normalized)
	if !floor.Equal(normalized) {
		return time.Time{}, ErrBoundarySequenceInvalid
	}

	next := policy.shift(normalized, 1)
	if !next.After(normalized) {
		return time.Time{}, ErrBoundarySequenceInvalid
	}

	return next, nil
}

func shiftBoundary(
	value time.Time,
	granularity historicalcontract.Granularity,
	steps int,
) (time.Time, error) {
	policy, exists := granularityPolicyFor(granularity)
	if !exists || policy.shift == nil {
		return time.Time{}, ErrUnsupportedGranularity
	}

	normalized := value.UTC()
	floor := policy.floor(normalized)
	if !floor.Equal(normalized) {
		return time.Time{}, ErrBoundarySequenceInvalid
	}

	shifted := policy.shift(normalized, steps)
	if steps < 0 && !shifted.Before(normalized) {
		return time.Time{}, ErrBoundarySequenceInvalid
	}
	if steps > 0 && !shifted.After(normalized) {
		return time.Time{}, ErrBoundarySequenceInvalid
	}

	return shifted, nil
}

func isSupportedGranularity(
	granularity historicalcontract.Granularity,
) bool {
	_, exists := granularityPolicyFor(granularity)
	return exists
}
