package airspaceregionanalytics

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/localtrafficscene"
)

type placement struct {
	bucketID          string
	bucketStart       time.Time
	bucketEnd         time.Time
	sectorID          string
	cellID            string
	latitudeIndex     int
	longitudeIndex    int
	altitudeBandIndex int
	altitudeKnown     bool
	aircraft          localtrafficscene.Aircraft
}

type occupancyContext struct {
	placements         []placement
	placementsBySector map[string][]placement
	nodeSectorByBucket map[string]map[string]string
	snapshotsByBucket  map[string][]SnapshotInput
	meanDataQuality    float64
	latestObservedAt   time.Time
	sourceNames        []string
}

func buildOccupancyIndex(
	request Request,
	policy Policy,
) (TemporalOccupancyIndex, occupancyContext) {
	selected := make(map[string]placement)
	snapshotsByBucket := make(map[string][]SnapshotInput)
	sourceSet := make(map[string]struct{})
	latestObservedAt := time.Time{}

	for _, snapshot := range request.Snapshots {
		bucketStart := bucketStartFor(snapshot.Scene.AsOfTime, policy.TimeBucketDuration)
		bucketEnd := bucketStart.Add(policy.TimeBucketDuration)
		bucketID := bucketIDFor(bucketStart)
		snapshotsByBucket[bucketID] = append(snapshotsByBucket[bucketID], snapshot)
		for _, aircraft := range snapshot.Scene.Aircraft {
			latitudeIndex := coordinateIndex(aircraft.Latitude, -90, policy.LatitudeCellDegrees)
			longitudeIndex := coordinateIndex(aircraft.Longitude, -180, policy.LongitudeCellDegrees)
			altitudeBandIndex, altitudeKnown := altitudeBandFor(
				aircraft.AltitudeMeters,
				policy.AltitudeBandMeters,
			)
			sectorID := sectorIDFor(bucketID, latitudeIndex, longitudeIndex)
			cellID := cellIDFor(sectorID, altitudeBandIndex, altitudeKnown)
			candidate := placement{
				bucketID:          bucketID,
				bucketStart:       bucketStart,
				bucketEnd:         bucketEnd,
				sectorID:          sectorID,
				cellID:            cellID,
				latitudeIndex:     latitudeIndex,
				longitudeIndex:    longitudeIndex,
				altitudeBandIndex: altitudeBandIndex,
				altitudeKnown:     altitudeKnown,
				aircraft:          aircraft,
			}
			key := bucketID + "|" + aircraft.NodeID
			current, exists := selected[key]
			if !exists || preferPlacement(candidate, current) {
				selected[key] = candidate
			}
			if aircraft.ObservedAt.After(latestObservedAt) {
				latestObservedAt = aircraft.ObservedAt
			}
			if aircraft.SourceName != "" {
				sourceSet[aircraft.SourceName] = struct{}{}
			}
		}
	}

	placements := make([]placement, 0, len(selected))
	for _, item := range selected {
		placements = append(placements, item)
	}
	sort.Slice(placements, func(left int, right int) bool {
		if placements[left].bucketStart.Equal(placements[right].bucketStart) {
			return placements[left].aircraft.NodeID < placements[right].aircraft.NodeID
		}
		return placements[left].bucketStart.Before(placements[right].bucketStart)
	})

	cellsByBucket := make(map[string]map[string][]placement)
	placementsBySector := make(map[string][]placement)
	nodeSectorByBucket := make(map[string]map[string]string)
	uniqueAircraft := make(map[string]struct{})
	qualityTotal := 0.0
	unknownAltitudeCount := 0
	for _, item := range placements {
		if cellsByBucket[item.bucketID] == nil {
			cellsByBucket[item.bucketID] = make(map[string][]placement)
		}
		cellsByBucket[item.bucketID][item.cellID] = append(
			cellsByBucket[item.bucketID][item.cellID],
			item,
		)
		placementsBySector[item.sectorID] = append(
			placementsBySector[item.sectorID],
			item,
		)
		if nodeSectorByBucket[item.bucketID] == nil {
			nodeSectorByBucket[item.bucketID] = make(map[string]string)
		}
		nodeSectorByBucket[item.bucketID][item.aircraft.NodeID] = item.sectorID
		uniqueAircraft[item.aircraft.NodeID] = struct{}{}
		qualityTotal += item.aircraft.QualityScore
		if !item.altitudeKnown {
			unknownAltitudeCount++
		}
	}

	bucketIDs := make([]string, 0, len(cellsByBucket))
	for bucketID := range cellsByBucket {
		bucketIDs = append(bucketIDs, bucketID)
	}
	sort.Strings(bucketIDs)

	buckets := make([]OccupancyBucket, 0, len(bucketIDs))
	peakAircraftPerBucket := 0
	peakOccupiedCells := 0
	totalOccupiedCells := 0
	for _, bucketID := range bucketIDs {
		cellMap := cellsByBucket[bucketID]
		cellIDs := make([]string, 0, len(cellMap))
		for cellID := range cellMap {
			cellIDs = append(cellIDs, cellID)
		}
		sort.Strings(cellIDs)
		cells := make([]OccupancyCell, 0, len(cellIDs))
		bucketAircraftCount := 0
		bucketUnknownAltitudeCount := 0
		bucketQualityTotal := 0.0
		var startTime time.Time
		var endTime time.Time
		for _, cellID := range cellIDs {
			items := cellMap[cellID]
			sort.Slice(items, func(left int, right int) bool {
				return items[left].aircraft.NodeID < items[right].aircraft.NodeID
			})
			nodeIDs := make([]string, 0, len(items))
			cellQualityTotal := 0.0
			for _, item := range items {
				nodeIDs = append(nodeIDs, item.aircraft.NodeID)
				cellQualityTotal += item.aircraft.QualityScore
				bucketQualityTotal += item.aircraft.QualityScore
				if !item.altitudeKnown {
					bucketUnknownAltitudeCount++
				}
			}
			first := items[0]
			startTime = first.bucketStart
			endTime = first.bucketEnd
			cells = append(cells, OccupancyCell{
				ID:                cellID,
				BucketID:          bucketID,
				BucketStart:       first.bucketStart,
				BucketEnd:         first.bucketEnd,
				LatitudeIndex:     first.latitudeIndex,
				LongitudeIndex:    first.longitudeIndex,
				AltitudeBandIndex: first.altitudeBandIndex,
				AltitudeKnown:     first.altitudeKnown,
				AircraftNodeIDs:   nodeIDs,
				AircraftCount:     len(items),
				MeanQualityScore:  cellQualityTotal / float64(len(items)),
			})
			bucketAircraftCount += len(items)
		}
		meanQuality := 0.0
		if bucketAircraftCount > 0 {
			meanQuality = bucketQualityTotal / float64(bucketAircraftCount)
		}
		buckets = append(buckets, OccupancyBucket{
			ID:        bucketID,
			StartTime: startTime,
			EndTime:   endTime,
			Cells:     cells,
			Metrics: OccupancyBucketMetrics{
				AircraftCount:        bucketAircraftCount,
				OccupiedCellCount:    len(cells),
				UnknownAltitudeCount: bucketUnknownAltitudeCount,
				MeanQualityScore:     meanQuality,
			},
		})
		peakAircraftPerBucket = maxInt(peakAircraftPerBucket, bucketAircraftCount)
		peakOccupiedCells = maxInt(peakOccupiedCells, len(cells))
		totalOccupiedCells += len(cells)
	}

	expectedBucketCount := expectedBuckets(
		request.WindowStart,
		request.WindowEnd,
		policy.TimeBucketDuration,
	)
	temporalCoverage := 0.0
	if expectedBucketCount > 0 {
		temporalCoverage = clampUnit(float64(len(buckets)) / float64(expectedBucketCount))
	}
	meanAircraftPerBucket := 0.0
	if len(buckets) > 0 {
		meanAircraftPerBucket = float64(len(placements)) / float64(len(buckets))
	}
	meanDataQuality := 0.0
	if len(placements) > 0 {
		meanDataQuality = qualityTotal / float64(len(placements))
	}

	sourceNames := make([]string, 0, len(sourceSet))
	for sourceName := range sourceSet {
		sourceNames = append(sourceNames, sourceName)
	}
	sort.Strings(sourceNames)

	index := TemporalOccupancyIndex{
		BucketDuration:       policy.TimeBucketDuration,
		LatitudeCellDegrees:  policy.LatitudeCellDegrees,
		LongitudeCellDegrees: policy.LongitudeCellDegrees,
		AltitudeBandMeters:   policy.AltitudeBandMeters,
		Buckets:              buckets,
		Metrics: OccupancyIndexMetrics{
			BucketCount:              len(buckets),
			ExpectedBucketCount:      expectedBucketCount,
			OccupiedCellCount:        totalOccupiedCells,
			AircraftObservationCount: len(placements),
			UniqueAircraftCount:      len(uniqueAircraft),
			UnknownAltitudeCount:     unknownAltitudeCount,
			PeakAircraftPerBucket:    peakAircraftPerBucket,
			PeakOccupiedCells:        peakOccupiedCells,
			MeanAircraftPerBucket:    meanAircraftPerBucket,
			TemporalCoverage:         temporalCoverage,
		},
	}
	return index, occupancyContext{
		placements:         placements,
		placementsBySector: placementsBySector,
		nodeSectorByBucket: nodeSectorByBucket,
		snapshotsByBucket:  snapshotsByBucket,
		meanDataQuality:    meanDataQuality,
		latestObservedAt:   latestObservedAt,
		sourceNames:        sourceNames,
	}
}

func preferPlacement(left, right placement) bool {
	if !left.aircraft.ObservedAt.Equal(right.aircraft.ObservedAt) {
		return left.aircraft.ObservedAt.After(right.aircraft.ObservedAt)
	}
	if left.aircraft.QualityScore != right.aircraft.QualityScore {
		return left.aircraft.QualityScore > right.aircraft.QualityScore
	}
	return left.aircraft.SourceName < right.aircraft.SourceName
}

func bucketStartFor(value time.Time, duration time.Duration) time.Time {
	return value.UTC().Truncate(duration)
}

func bucketIDFor(start time.Time) string {
	return start.UTC().Format("20060102T150405.000000000Z")
}

func coordinateIndex(value float64, origin float64, size float64) int {
	return int(math.Floor((value - origin) / size))
}

func altitudeBandFor(value *float64, bandSize float64) (int, bool) {
	if value == nil {
		return -1, false
	}
	return int(math.Floor(*value / bandSize)), true
}

func sectorIDFor(bucketID string, latitudeIndex, longitudeIndex int) string {
	return fmt.Sprintf("%s|lat:%d|lon:%d", bucketID, latitudeIndex, longitudeIndex)
}

func cellIDFor(sectorID string, altitudeBandIndex int, altitudeKnown bool) string {
	if !altitudeKnown {
		return sectorID + "|alt:unknown"
	}
	return fmt.Sprintf("%s|alt:%d", sectorID, altitudeBandIndex)
}

func expectedBuckets(start, end time.Time, duration time.Duration) int {
	window := end.Sub(start)
	if window <= 0 {
		return 0
	}
	return int(math.Ceil(float64(window) / float64(duration)))
}
