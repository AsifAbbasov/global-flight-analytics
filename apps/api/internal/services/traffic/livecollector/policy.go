package livecollector

import (
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

func MinimumRequestSpacing(
	provider providerpolicy.Provider,
) (time.Duration, error) {
	policy, err := providerpolicy.Get(provider)
	if err != nil {
		return 0, err
	}
	if policy.BudgetMode != providerpolicy.BudgetModeFixedWindow {
		return 0, fmt.Errorf(
			"provider %s does not expose a fixed-window request policy",
			provider,
		)
	}

	var minimum time.Duration
	for _, limit := range policy.RequestLimits {
		window, err := policyWindowDuration(limit.Window)
		if err != nil {
			return 0, err
		}
		candidate := window / time.Duration(limit.MaxRequests)
		if candidate > minimum {
			minimum = candidate
		}
	}

	if minimum <= 0 {
		return 0, fmt.Errorf(
			"provider %s produced no positive request spacing",
			provider,
		)
	}
	return minimum, nil
}

func MinimumPollInterval(
	provider providerpolicy.Provider,
	targetCount int,
) (time.Duration, error) {
	if targetCount <= 0 {
		return 0, fmt.Errorf("target count must be greater than zero")
	}

	spacing, err := MinimumRequestSpacing(provider)
	if err != nil {
		return 0, err
	}
	return spacing * time.Duration(targetCount), nil
}

func policyWindowDuration(window providerpolicy.Window) (time.Duration, error) {
	switch window {
	case providerpolicy.WindowSecond:
		return time.Second, nil
	case providerpolicy.WindowMinute:
		return time.Minute, nil
	case providerpolicy.WindowHour:
		return time.Hour, nil
	case providerpolicy.WindowDay:
		return 24 * time.Hour, nil
	case providerpolicy.WindowMonth:
		return 31 * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("unsupported provider-policy window %q", window)
	}
}
