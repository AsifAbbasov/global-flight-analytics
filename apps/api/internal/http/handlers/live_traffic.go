package handlers

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/dto"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/response"
	livetraffic "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/live"
	"github.com/gofiber/fiber/v2"
)

type liveTrafficClock func() time.Time

type LiveTrafficHandler struct {
	store *livetraffic.Store
	now   liveTrafficClock
}

func NewLiveTrafficHandler(store *livetraffic.Store) *LiveTrafficHandler {
	return newLiveTrafficHandlerWithClock(store, time.Now)
}

func newLiveTrafficHandlerWithClock(
	store *livetraffic.Store,
	now liveTrafficClock,
) *LiveTrafficHandler {
	if store == nil {
		panic("live traffic store is required")
	}
	if now == nil {
		panic("live traffic clock is required")
	}
	return &LiveTrafficHandler{store: store, now: now}
}

func (h *LiveTrafficHandler) GetSnapshot(c *fiber.Ctx) error {
	query, err := parseLiveTrafficQuery(c)
	if err != nil {
		return response.Error(
			c,
			fiber.StatusBadRequest,
			"LIVE_TRAFFIC_QUERY_INVALID",
			"Invalid live traffic query",
		)
	}

	snapshot, err := h.store.Snapshot(h.now().UTC(), query)
	if err != nil {
		if errors.Is(err, livetraffic.ErrInvalidBounds) ||
			errors.Is(err, livetraffic.ErrInvalidLimit) ||
			errors.Is(err, livetraffic.ErrTooManySelected) ||
			strings.Contains(err.Error(), "ICAO24") {
			return response.Error(
				c,
				fiber.StatusBadRequest,
				"LIVE_TRAFFIC_QUERY_INVALID",
				"Invalid live traffic query",
			)
		}
		return response.Error(
			c,
			fiber.StatusInternalServerError,
			"LIVE_TRAFFIC_SNAPSHOT_FAILED",
			"Failed to build live traffic snapshot",
		)
	}

	return response.OK(c, toLiveTrafficSnapshot(snapshot))
}

func parseLiveTrafficQuery(c *fiber.Ctx) (livetraffic.SnapshotQuery, error) {
	query := livetraffic.SnapshotQuery{}
	boundsNames := []string{"min_lat", "min_lon", "max_lat", "max_lon"}
	boundsValues := make([]string, len(boundsNames))
	provided := 0
	for index, name := range boundsNames {
		boundsValues[index] = strings.TrimSpace(c.Query(name))
		if boundsValues[index] != "" {
			provided++
		}
	}
	if provided != 0 && provided != len(boundsNames) {
		return livetraffic.SnapshotQuery{}, fmt.Errorf("all bounding box values are required together")
	}
	if provided == len(boundsNames) {
		parsed := make([]float64, len(boundsValues))
		for index, value := range boundsValues {
			number, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return livetraffic.SnapshotQuery{}, err
			}
			parsed[index] = number
		}
		bounds := livetraffic.Bounds{
			MinLatitude:  parsed[0],
			MinLongitude: parsed[1],
			MaxLatitude:  parsed[2],
			MaxLongitude: parsed[3],
		}
		if err := bounds.Validate(); err != nil {
			return livetraffic.SnapshotQuery{}, err
		}
		query.Bounds = &bounds
	}

	if selected := strings.TrimSpace(c.Query("selected")); selected != "" {
		for _, value := range strings.Split(selected, ",") {
			normalized := strings.TrimSpace(value)
			if normalized != "" {
				query.SelectedICAO24 = append(query.SelectedICAO24, normalized)
			}
		}
	}

	if rawLimit := strings.TrimSpace(c.Query("limit")); rawLimit != "" {
		limit, err := strconv.Atoi(rawLimit)
		if err != nil {
			return livetraffic.SnapshotQuery{}, err
		}
		query.Limit = limit
	}
	return query, nil
}

func toLiveTrafficSnapshot(snapshot livetraffic.Snapshot) dto.LiveTrafficSnapshot {
	items := make([]dto.LiveTrafficItem, 0, len(snapshot.Aircraft))
	for _, item := range snapshot.Aircraft {
		freshness := snapshot.ServerTime.Sub(item.ObservedAt).Milliseconds()
		if freshness < 0 {
			freshness = 0
		}
		items = append(items, dto.LiveTrafficItem{
			ICAO24:          item.ICAO24,
			Callsign:        item.Callsign,
			Latitude:        item.Latitude,
			Longitude:       item.Longitude,
			AltitudeM:       item.AltitudeM,
			VelocityMPS:     item.VelocityMPS,
			HeadingDegrees:  item.HeadingDegrees,
			VerticalRateMPS: item.VerticalRateMPS,
			OnGround:        item.OnGround,
			ObservedAt:      item.ObservedAt,
			ReceivedAt:      item.ReceivedAt,
			Source:          item.Source,
			FreshnessMS:     freshness,
		})
	}
	return dto.LiveTrafficSnapshot{
		ServerTime:  snapshot.ServerTime,
		Sequence:    snapshot.Sequence,
		Aircraft:    items,
		TotalActive: snapshot.TotalActive,
		Matching:    snapshot.Matching,
		Truncated:   snapshot.Truncated,
	}
}
