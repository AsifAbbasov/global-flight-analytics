package region

type Bounds struct {
	MinLatitude  float64
	MaxLatitude  float64
	MinLongitude float64
	MaxLongitude float64
}

type Region struct {
	Code        string
	Name        string
	Description string
	Bounds      Bounds
}
