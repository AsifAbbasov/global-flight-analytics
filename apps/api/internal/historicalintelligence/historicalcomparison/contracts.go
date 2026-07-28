package historicalcomparison

const Version = "historical-period-comparison-v2"

const coverageEqualityTolerance = 1e-12

type periodValues struct {
	current  float64
	previous float64
}
