package airport

type Airport struct {
	ICAOCode           string
	IATACode           string
	Name               string
	City               string
	Country            string
	Latitude           float64
	Longitude          float64
	ElevationM         float64
	ElevationAvailable bool
	Timezone           string
	Description        string
}
