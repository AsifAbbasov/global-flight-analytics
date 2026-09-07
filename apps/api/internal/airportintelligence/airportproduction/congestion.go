package airportproduction

import (
	"context"
	"errors"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/congestion"
)

func (service *Service) GetCongestion(
	ctx context.Context,
	icaoCode string,
	request WindowRequest,
) (CongestionResult, error) {
	if service == nil {
		return CongestionResult{}, ErrInvalidConfiguration
	}

	historyResult, err := service.GetHistory(ctx, icaoCode, request)
	if err != nil {
		return CongestionResult{}, err
	}

	generatedAt := service.now().UTC()
	value, err := congestion.NewAnalyzer().Analyze(congestion.Input{
		History:     historyResult.History,
		WindowStart: historyResult.Window.StartTime,
		WindowEnd:   historyResult.Window.EndTime,
		GeneratedAt: generatedAt,
	})
	if err != nil {
		if errors.Is(err, congestion.ErrInsufficientHistory) {
			return CongestionResult{}, fmt.Errorf("%w: %v", ErrInsufficientHistory, err)
		}
		return CongestionResult{}, fmt.Errorf("analyze Airport Congestion Intelligence: %w", err)
	}

	limitations := append([]Limitation(nil), historyResult.Limitations...)
	limitations = append(limitations,
		Limitation{
			Code:    "RELATIVE_OBSERVED_ACTIVITY_PROXY",
			Message: "Airport Congestion Intelligence compares completed-day observed movement intensity only against the same airport's prior observed history.",
		},
		Limitation{
			Code:    "NO_AIRPORT_CAPACITY_OR_DELAY_MODEL",
			Message: "The result does not model declared airport capacity, runway occupancy, queues, slots, schedules, delays, or delay causes.",
		},
	)

	return CongestionResult{
		Version:     congestion.Version,
		Window:      historyResult.Window,
		Congestion:  value,
		Limitations: limitations,
		GeneratedAt: generatedAt,
	}, nil
}
