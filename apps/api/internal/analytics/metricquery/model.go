package metricquery

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	DefaultWindowMinutes = 15
	MinimumWindowMinutes = 1
	MaximumWindowMinutes = 180

	DefaultResultLimit = 1000
	MinimumResultLimit = 1
	MaximumResultLimit = 5000

	MaximumTrajectoryIDCount = 500
)

var (
	ErrRepositoryRequired = errors.New(
		"analytical trajectory repository is required",
	)
	ErrWindowMinutesInvalid = errors.New(
		"analytical trajectory window minutes must be between one and one hundred eighty",
	)
	ErrResultLimitInvalid = errors.New(
		"analytical trajectory result limit must be between one and five thousand",
	)
	ErrReferenceTimeRequired = errors.New(
		"analytical trajectory reference time is required",
	)
	ErrTrajectoryIDsMissing = errors.New(
		"at least one trajectory id is required",
	)
	ErrTrajectoryIDCountExceeded = errors.New(
		"trajectory id count exceeds the maximum of five hundred",
	)
	ErrTrajectoryIDInvalid = errors.New(
		"trajectory id is invalid",
	)
)

type RecentRequest struct {
	WindowMinutes int
	Limit         int
}

type Window struct {
	ObservedFrom time.Time
	ObservedTo   time.Time
	Limit        int
}

func (
	request RecentRequest,
) Normalize(
	now time.Time,
) (Window, error) {
	if now.IsZero() {
		return Window{}, ErrReferenceTimeRequired
	}

	windowMinutes := request.WindowMinutes
	if windowMinutes == 0 {
		windowMinutes = DefaultWindowMinutes
	}
	if windowMinutes < MinimumWindowMinutes ||
		windowMinutes > MaximumWindowMinutes {
		return Window{}, ErrWindowMinutesInvalid
	}

	limit := request.Limit
	if limit == 0 {
		limit = DefaultResultLimit
	}
	if limit < MinimumResultLimit ||
		limit > MaximumResultLimit {
		return Window{}, ErrResultLimitInvalid
	}

	referenceTime := now.UTC()

	return Window{
		ObservedFrom: referenceTime.Add(
			-time.Duration(windowMinutes) *
				time.Minute,
		),
		ObservedTo: referenceTime,
		Limit:      limit,
	}, nil
}

func NormalizeTrajectoryIDs(
	values []string,
) ([]string, error) {
	if len(values) == 0 {
		return nil, ErrTrajectoryIDsMissing
	}
	if len(values) > MaximumTrajectoryIDCount {
		return nil, ErrTrajectoryIDCountExceeded
	}

	seen := make(
		map[string]struct{},
		len(values),
	)
	result := make(
		[]string,
		0,
		len(values),
	)

	for index, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return nil, fmt.Errorf(
				"%w: index=%d",
				ErrTrajectoryIDInvalid,
				index,
			)
		}

		parsed, err := uuid.Parse(trimmed)
		if err != nil {
			return nil, fmt.Errorf(
				"%w: index=%d",
				ErrTrajectoryIDInvalid,
				index,
			)
		}

		canonical := parsed.String()
		if _, exists := seen[canonical]; exists {
			continue
		}

		seen[canonical] = struct{}{}
		result = append(
			result,
			canonical,
		)
	}

	if len(result) == 0 {
		return nil, ErrTrajectoryIDsMissing
	}

	return result, nil
}
