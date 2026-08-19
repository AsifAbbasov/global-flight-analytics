package dto

import "time"

type LiveTrafficItem struct {
	ICAO24          string    `json:"icao24"`
	Callsign        string    `json:"callsign"`
	Latitude        float64   `json:"latitude"`
	Longitude       float64   `json:"longitude"`
	AltitudeM       *float64  `json:"altitude_m"`
	VelocityMPS     *float64  `json:"velocity_mps"`
	HeadingDegrees  *float64  `json:"heading_degrees"`
	VerticalRateMPS *float64  `json:"vertical_rate_mps"`
	OnGround        *bool     `json:"on_ground"`
	ObservedAt      time.Time `json:"observed_at"`
	ReceivedAt      time.Time `json:"received_at"`
	Source          string    `json:"source"`
	FreshnessMS     int64     `json:"freshness_ms"`
}

type LiveTrafficSnapshot struct {
	ServerTime  time.Time         `json:"server_time"`
	Sequence    uint64            `json:"sequence"`
	Aircraft    []LiveTrafficItem `json:"aircraft"`
	TotalActive int               `json:"total_active"`
	Matching    int               `json:"matching"`
	Truncated   bool              `json:"truncated"`
}
