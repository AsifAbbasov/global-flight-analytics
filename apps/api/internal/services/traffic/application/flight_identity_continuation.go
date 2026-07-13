package application

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/flightcontinuation"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/processor"
)

type continuationCandidate struct {
	collectionKey string
	item          trajectory.FlightTrajectory
}

type continuationUpdate struct {
	oldCollectionKey string
	newCollectionKey string
	item             trajectory.FlightTrajectory
}

func (
	service *Service,
) applyFlightIdentityContinuations(
	ctx context.Context,
	result *processor.ProcessingResult,
) (int, error) {
	if service.trajectoryContinuationRepository == nil ||
		service.identityContinuationConfig.MaxGap <= 0 ||
		result == nil ||
		len(result.Trajectories) == 0 {
		return 0, nil
	}

	candidates := earliestContinuationCandidates(
		result.Trajectories,
	)
	icao24Values := make(
		[]string,
		0,
		len(candidates),
	)

	for icao24 := range candidates {
		icao24Values = append(
			icao24Values,
			icao24,
		)
	}

	sort.Strings(icao24Values)

	updates := make(
		[]continuationUpdate,
		0,
		len(icao24Values),
	)

	for _, icao24 := range icao24Values {
		candidate := candidates[icao24]

		previous, err :=
			service.trajectoryContinuationRepository.
				GetLatestTrajectoryByICAO24(
					ctx,
					icao24,
				)
		if errors.Is(
			err,
			trajectory.ErrNotFound,
		) {
			continue
		}
		if err != nil {
			return 0, fmt.Errorf(
				"load previous trajectory for flight identity continuation for icao24 %s: %w",
				icao24,
				err,
			)
		}

		continued, ok := flightcontinuation.Continue(
			previous,
			candidate.item,
			service.identityContinuationConfig,
		)
		if !ok {
			continue
		}

		newCollectionKey := candidate.collectionKey
		if candidate.collectionKey != icao24 {
			newCollectionKey = continued.IdentityKey
		}

		updates = append(
			updates,
			continuationUpdate{
				oldCollectionKey: candidate.collectionKey,
				newCollectionKey: newCollectionKey,
				item:             continued,
			},
		)
	}

	if len(updates) == 0 {
		return 0, nil
	}

	nextTrajectories := make(
		map[string]trajectory.FlightTrajectory,
		len(result.Trajectories),
	)

	for key, item := range result.Trajectories {
		nextTrajectories[key] = item
	}

	for _, update := range updates {
		if update.newCollectionKey != update.oldCollectionKey {
			if _, exists :=
				nextTrajectories[update.newCollectionKey]; exists {
				return 0, fmt.Errorf(
					"flight identity continuation collection key collision: %s",
					update.newCollectionKey,
				)
			}

			delete(
				nextTrajectories,
				update.oldCollectionKey,
			)
		}

		nextTrajectories[update.newCollectionKey] =
			update.item
	}

	result.Trajectories = nextTrajectories

	return len(updates), nil
}

func earliestContinuationCandidates(
	items map[string]trajectory.FlightTrajectory,
) map[string]continuationCandidate {
	result := make(
		map[string]continuationCandidate,
	)

	for collectionKey, item := range items {
		icao24 := strings.ToUpper(
			strings.TrimSpace(
				item.ICAO24,
			),
		)
		if icao24 == "" {
			continue
		}

		current, exists := result[icao24]
		if !exists ||
			isEarlierContinuationCandidate(
				collectionKey,
				item,
				current,
			) {
			result[icao24] = continuationCandidate{
				collectionKey: collectionKey,
				item:          item,
			}
		}
	}

	return result
}

func isEarlierContinuationCandidate(
	collectionKey string,
	item trajectory.FlightTrajectory,
	current continuationCandidate,
) bool {
	if item.StartTime.IsZero() {
		return false
	}

	if current.item.StartTime.IsZero() {
		return true
	}

	if item.StartTime.Before(
		current.item.StartTime,
	) {
		return true
	}

	if item.StartTime.After(
		current.item.StartTime,
	) {
		return false
	}

	return collectionKey <
		current.collectionKey
}
