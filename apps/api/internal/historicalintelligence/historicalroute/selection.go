package historicalroute

import (
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalread"
)

func latestRouteRecords(
	records []historicalread.RouteRecord,
	asOfTime time.Time,
) ([]historicalread.RouteRecord, error) {
	latest := make(map[string]historicalread.RouteRecord)
	cutoff := asOfTime.UTC()

	for _, record := range records {
		trajectoryID := strings.TrimSpace(record.TrajectoryID)
		if strings.TrimSpace(record.ID) == "" || trajectoryID == "" {
			return nil, ErrRouteRecordIdentityRequired
		}
		if record.AsOfTime.IsZero() || record.AsOfTime.After(cutoff) {
			return nil, &EvidenceError{
				RecordID: record.ID,
				Err:      ErrRouteRecordTimeInvalid,
			}
		}

		current, exists := latest[trajectoryID]
		if !exists || newerRouteRecord(record, current) {
			latest[trajectoryID] = record
		}
	}

	result := make([]historicalread.RouteRecord, 0, len(latest))
	for _, record := range latest {
		result = append(result, record)
	}
	sort.SliceStable(result, func(left int, right int) bool {
		if !result[left].AsOfTime.Equal(result[right].AsOfTime) {
			return result[left].AsOfTime.Before(result[right].AsOfTime)
		}
		if !result[left].StoredAt.Equal(result[right].StoredAt) {
			return result[left].StoredAt.Before(result[right].StoredAt)
		}
		return result[left].ID < result[right].ID
	})
	return result, nil
}

func newerRouteRecord(
	candidate historicalread.RouteRecord,
	current historicalread.RouteRecord,
) bool {
	if !candidate.AsOfTime.Equal(current.AsOfTime) {
		return candidate.AsOfTime.After(current.AsOfTime)
	}
	if !candidate.StoredAt.Equal(current.StoredAt) {
		return candidate.StoredAt.After(current.StoredAt)
	}
	return candidate.ID < current.ID
}
