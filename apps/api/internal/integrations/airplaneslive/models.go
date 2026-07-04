package airplaneslive

type StateResponse struct {
	Now      float64        `json:"now"`
	Messages int            `json:"messages"`
	Total    int            `json:"total"`
	Aircraft []AircraftItem `json:"ac"`
}

type AircraftItem struct {
	Hex          string   `json:"hex"`
	Flight       string   `json:"flight"`
	Latitude     float64  `json:"lat"`
	Longitude    float64  `json:"lon"`
	AltBaro      any      `json:"alt_baro"`
	AltGeom      *float64 `json:"alt_geom"`
	GroundSpeed  float64  `json:"gs"`
	Track        float64  `json:"track"`
	BaroRate     float64  `json:"baro_rate"`
	Seen         float64  `json:"seen"`
	Type         string   `json:"type"`
	Registration string   `json:"r"`
	AircraftType string   `json:"t"`
}
