package statistics

import "time"

type Input struct {
	ICAOCode            string
	WindowStart         time.Time
	WindowEnd           time.Time
	Arrivals            int
	Departures          int
	ActiveAircraft      int
	ActiveRoutes        int
	ObservedSamples     int
	ExpectedSamples     int
	LatestObservationAt time.Time
	GeneratedAt         time.Time
}

type Statistics struct {
	ICAOCode            string
	WindowStart         time.Time
	WindowEnd           time.Time
	Arrivals            int
	Departures          int
	TotalMovements      int
	ArrivalShare        float64
	DepartureShare      float64
	MovementsPerHour    float64
	ActiveAircraft      int
	ActiveRoutes        int
	ObservedSamples     int
	ExpectedSamples     int
	CoverageScore       float64
	FreshnessScore      float64
	LatestObservationAt time.Time
	GeneratedAt         time.Time
}
