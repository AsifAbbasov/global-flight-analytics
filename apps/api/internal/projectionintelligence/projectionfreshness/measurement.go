package projectionfreshness

import (
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
)

type freshnessMetrics struct {
	ages        []time.Duration
	selectedIDs []string
	recentCount int
	newestAge   time.Duration
	meanAge     time.Duration
	oldestAge   time.Duration
}

func measureFreshness(
	selection projectionneighbors.Result,
	config Config,
) freshnessMetrics {
	metrics := freshnessMetrics{
		ages:        make([]time.Duration, 0, len(selection.Neighbors)),
		selectedIDs: make([]string, 0, len(selection.Neighbors)),
	}
	asOfTime := selection.AsOfTime.UTC()
	for _, neighbor := range selection.Neighbors {
		age := asOfTime.Sub(neighbor.CandidateEndTime.UTC())
		metrics.ages = append(metrics.ages, age)
		metrics.selectedIDs = append(
			metrics.selectedIDs,
			strings.TrimSpace(neighbor.TrajectoryID),
		)
		if age <= config.RecentNeighborAgeLimit {
			metrics.recentCount++
		}
	}
	sort.Strings(metrics.selectedIDs)
	sort.Slice(metrics.ages, func(left int, right int) bool {
		return metrics.ages[left] < metrics.ages[right]
	})
	if len(metrics.ages) > 0 {
		metrics.newestAge = metrics.ages[0]
		metrics.meanAge = meanDuration(metrics.ages)
		metrics.oldestAge = metrics.ages[len(metrics.ages)-1]
	}
	return metrics
}

func meanDuration(values []time.Duration) time.Duration {
	if len(values) == 0 {
		return 0
	}
	count := int64(len(values))
	var quotient int64
	var remainder int64
	for _, value := range values {
		nanoseconds := int64(value)
		quotient += nanoseconds / count
		remainder += nanoseconds % count
	}
	return time.Duration(quotient + remainder/count)
}
