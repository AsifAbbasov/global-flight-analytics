package passport

import "time"

type Passport struct {
	Identity    Identity
	Location    Location
	Operations  Operations
	DataQuality DataQuality
	Description string
	GeneratedAt time.Time
}

type Identity struct {
	ICAOCode string
	IATACode string
	Name     string
}

type Location struct {
	City       string
	Country    string
	Latitude   float64
	Longitude  float64
	ElevationM float64
	Timezone   string
}

type Operations struct {
	Arrivals       int
	Departures     int
	Activity       int
	ActiveAircraft int
}

type DataQuality struct {
	FreshnessScore float64
	CoverageScore  float64
	ObservedAt     time.Time
}

type AnalyticsInput struct {
	Arrivals       int
	Departures     int
	ActiveAircraft int
	FreshnessScore float64
	CoverageScore  float64
	ObservedAt     time.Time
}
