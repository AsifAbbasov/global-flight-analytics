package metrics

const ActiveAircraftMetricID = "traffic.active_aircraft"

type ActiveAircraft struct{}

func (ActiveAircraft) ID() string {
	return ActiveAircraftMetricID
}

func (ActiveAircraft) Name() string {
	return "traffic.active_aircraft"
}

func (ActiveAircraft) Calculate(total int) int {
	if total < 0 {
		return 0
	}

	return total
}
