package historicalroute

import (
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalread"
)

func routeSourceNames(items []routeEvidence) []string {
	unique := map[string]struct{}{
		historicalread.DatasetRoutes: {},
	}
	for _, item := range items {
		for _, sourceName := range item.result.Provenance.SourceNames {
			normalized := strings.TrimSpace(sourceName)
			if normalized != "" {
				unique[normalized] = struct{}{}
			}
		}
	}

	result := make([]string, 0, len(unique))
	for sourceName := range unique {
		result = append(result, sourceName)
	}
	sort.Strings(result)
	return result
}

func latestRouteUpdate(items []routeEvidence) time.Time {
	latest := time.Time{}
	for _, item := range items {
		latest = laterTime(latest, item.record.StoredAt)
		latest = laterTime(
			latest,
			item.result.Provenance.TrajectoryUpdatedAt,
		)
	}
	return latest
}

func laterTime(current time.Time, candidate time.Time) time.Time {
	if candidate.IsZero() {
		return current
	}
	normalized := candidate.UTC()
	if current.IsZero() || normalized.After(current) {
		return normalized
	}
	return current
}
