package metricexecution

import (
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/analyticalresult"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func prepareUniqueAircraftContributors(
	messageFormat string,
) trajectoryMetricPreparation {
	return func(
		items []trajectory.FlightTrajectory,
	) (
		[]trajectory.FlightTrajectory,
		[]analyticalresult.Notice,
	) {
		unique, removed := uniqueAircraftTrajectories(items)
		if removed == 0 {
			return unique, nil
		}

		return unique, []analyticalresult.Notice{
			{
				Code: NoticeCodeDuplicateTrajectoriesRemoved,
				Message: fmt.Sprintf(
					messageFormat,
					removed,
				),
			},
		}
	}
}
