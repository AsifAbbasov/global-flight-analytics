package opensky

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var (
	ErrWaypointFieldCount = errors.New("OpenSky track waypoint has fewer than six fields")
)

type TrackResponse struct {
	ICAO24    string            `json:"icao24"`
	StartTime int64             `json:"startTime"`
	EndTime   int64             `json:"endTime"`
	Callsign  *string           `json:"callsign"`
	Path      []json.RawMessage `json:"path"`
}

type Waypoint struct {
	Time          time.Time
	Latitude      *float64
	Longitude     *float64
	BaroAltitudeM *float64
	TrueTrack     *float64
	OnGround      bool
}

func ParseWaypoint(raw json.RawMessage) (Waypoint, error) {
	var values []json.RawMessage
	if err := json.Unmarshal(raw, &values); err != nil {
		return Waypoint{}, fmt.Errorf("decode OpenSky waypoint array: %w", err)
	}
	if len(values) < 6 {
		return Waypoint{}, fmt.Errorf("%w: got %d", ErrWaypointFieldCount, len(values))
	}
	timestamp, err := requiredInt64(values[0], "waypoint_time")
	if err != nil {
		return Waypoint{}, err
	}
	latitude, err := optionalFloat64(values[1], "waypoint_latitude")
	if err != nil {
		return Waypoint{}, err
	}
	longitude, err := optionalFloat64(values[2], "waypoint_longitude")
	if err != nil {
		return Waypoint{}, err
	}
	altitude, err := optionalFloat64(values[3], "waypoint_baro_altitude")
	if err != nil {
		return Waypoint{}, err
	}
	track, err := optionalFloat64(values[4], "waypoint_true_track")
	if err != nil {
		return Waypoint{}, err
	}
	onGround, err := requiredBool(values[5], "waypoint_on_ground")
	if err != nil {
		return Waypoint{}, err
	}
	return Waypoint{
		Time:          time.Unix(timestamp, 0).UTC(),
		Latitude:      latitude,
		Longitude:     longitude,
		BaroAltitudeM: altitude,
		TrueTrack:     track,
		OnGround:      onGround,
	}, nil
}

func (response TrackResponse) ParsePath() ([]Waypoint, error) {
	path := make([]Waypoint, 0, len(response.Path))
	for index, raw := range response.Path {
		point, err := ParseWaypoint(raw)
		if err != nil {
			return nil, fmt.Errorf("parse OpenSky waypoint %d: %w", index, err)
		}
		path = append(path, point)
	}
	return path, nil
}

func ExperimentalTrackDisclosure() []string {
	return []string{
		"OpenSky documents the track endpoint as experimental.",
		"Provider tracks are secondary evidence and must not replace the project Track Builder.",
		"Tracks older than thirty days are unavailable through the REST endpoint.",
	}
}
