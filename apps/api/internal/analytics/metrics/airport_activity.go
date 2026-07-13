package metrics

type AirportActivity struct{}

func (AirportActivity) Name() string {
	return "traffic.airport_activity"
}

func (AirportActivity) Calculate(arrivals, departures int) int {
	if arrivals < 0 {
		arrivals = 0
	}

	if departures < 0 {
		departures = 0
	}

	return arrivals + departures
}
