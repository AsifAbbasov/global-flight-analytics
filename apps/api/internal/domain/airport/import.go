package airport

import "time"

type ImportRecord struct {
	SourceIdent       string
	ICAOCode          string
	IATACode          string
	Name              string
	City              string
	SourceCountryCode string
	Latitude          float64
	Longitude         float64
	ElevationFT       *int
	SourceName        string
	LastSyncedAt      time.Time
}
