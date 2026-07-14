package datasetprofiler

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

type frequencyAccumulator struct {
	occurrences     map[string]int
	affectedRecords map[string]int
}

func newFrequencyAccumulator() frequencyAccumulator {
	return frequencyAccumulator{
		occurrences:     make(map[string]int),
		affectedRecords: make(map[string]int),
	}
}

func (accumulator frequencyAccumulator) addRecordValues(
	values []string,
) {
	seen := make(map[string]struct{}, len(values))

	for _, rawValue := range values {
		value := strings.TrimSpace(rawValue)
		if value == "" {
			continue
		}

		accumulator.occurrences[value]++
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		accumulator.affectedRecords[value]++
	}
}

func (accumulator frequencyAccumulator) profiles() []FrequencyProfile {
	result := make(
		[]FrequencyProfile,
		0,
		len(accumulator.occurrences),
	)
	for value, count := range accumulator.occurrences {
		result = append(
			result,
			FrequencyProfile{
				Value:               value,
				OccurrenceCount:     count,
				AffectedRecordCount: accumulator.affectedRecords[value],
			},
		)
	}

	sort.SliceStable(
		result,
		func(left int, right int) bool {
			if result[left].AffectedRecordCount !=
				result[right].AffectedRecordCount {
				return result[left].AffectedRecordCount >
					result[right].AffectedRecordCount
			}
			if result[left].OccurrenceCount !=
				result[right].OccurrenceCount {
				return result[left].OccurrenceCount >
					result[right].OccurrenceCount
			}

			return result[left].Value <
				result[right].Value
		},
	)

	return result
}

func (accumulator frequencyAccumulator) limitationProfiles() []LimitationProfile {
	frequencies := accumulator.profiles()
	result := make(
		[]LimitationProfile,
		0,
		len(frequencies),
	)
	for _, frequency := range frequencies {
		result = append(
			result,
			LimitationProfile{
				Code:                frequency.Value,
				OccurrenceCount:     frequency.OccurrenceCount,
				AffectedRecordCount: frequency.AffectedRecordCount,
			},
		)
	}

	return result
}

type timeAccumulator struct {
	earliestWindowStart time.Time
	latestWindowEnd     time.Time
	earliestAsOfTime    time.Time
	latestAsOfTime      time.Time
	earliestExtractedAt time.Time
	latestExtractedAt   time.Time
}

func (accumulator *timeAccumulator) add(
	features flightfeatures.FlightFeatures,
) {
	accumulator.earliestWindowStart =
		earlierNonZero(
			accumulator.earliestWindowStart,
			features.Window.StartTime,
		)
	accumulator.latestWindowEnd =
		laterNonZero(
			accumulator.latestWindowEnd,
			features.Window.EndTime,
		)
	accumulator.earliestAsOfTime =
		earlierNonZero(
			accumulator.earliestAsOfTime,
			features.Window.AsOfTime,
		)
	accumulator.latestAsOfTime =
		laterNonZero(
			accumulator.latestAsOfTime,
			features.Window.AsOfTime,
		)
	accumulator.earliestExtractedAt =
		earlierNonZero(
			accumulator.earliestExtractedAt,
			features.ExtractedAt,
		)
	accumulator.latestExtractedAt =
		laterNonZero(
			accumulator.latestExtractedAt,
			features.ExtractedAt,
		)
}

func (accumulator timeAccumulator) profile() TimeProfile {
	return TimeProfile{
		EarliestWindowStart: accumulator.earliestWindowStart,
		LatestWindowEnd:     accumulator.latestWindowEnd,
		EarliestAsOfTime:    accumulator.earliestAsOfTime,
		LatestAsOfTime:      accumulator.latestAsOfTime,
		EarliestExtractedAt: accumulator.earliestExtractedAt,
		LatestExtractedAt:   accumulator.latestExtractedAt,
	}
}

func earlierNonZero(
	current time.Time,
	candidate time.Time,
) time.Time {
	if candidate.IsZero() {
		return current
	}

	candidate = candidate.UTC()
	if current.IsZero() ||
		candidate.Before(current) {
		return candidate
	}

	return current
}

func laterNonZero(
	current time.Time,
	candidate time.Time,
) time.Time {
	if candidate.IsZero() {
		return current
	}

	candidate = candidate.UTC()
	if current.IsZero() ||
		candidate.After(current) {
		return candidate
	}

	return current
}

func snapshotKey(
	features flightfeatures.FlightFeatures,
) string {
	return fmt.Sprintf(
		"%s|%s|%s",
		strings.TrimSpace(features.TrajectoryID),
		features.SchemaVersion,
		features.Window.AsOfTime.UTC().
			Format(time.RFC3339Nano),
	)
}
