package regionalprovider

func PointRequestKey(
	latitude float64,
	longitude float64,
	radius int,
) string {
	return regionalRequestKey(
		latitude,
		longitude,
		radius,
	)
}
