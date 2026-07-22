package flightstate

type AircraftCategory struct {
	value     int
	available bool
}

func NewAircraftCategory(value int) (AircraftCategory, error) {
	category := AircraftCategory{
		value:     value,
		available: true,
	}
	if err := category.Validate(); err != nil {
		return AircraftCategory{}, err
	}
	return category, nil
}

func UnavailableAircraftCategory() AircraftCategory {
	return AircraftCategory{}
}

func (value AircraftCategory) Value() int {
	return value.value
}

func (value AircraftCategory) Available() bool {
	return value.available
}

func (value AircraftCategory) Validate() error {
	if !value.available {
		if value.value != 0 {
			return ErrAircraftCategoryInvalid
		}
		return nil
	}
	if value.value < MinimumAircraftCategory ||
		value.value > MaximumAircraftCategory {
		return ErrAircraftCategoryInvalid
	}
	return nil
}

func (state FlightState) ResolveAircraftCategory() (AircraftCategory, error) {
	if !state.AircraftCategoryAvailable {
		if state.AircraftCategory != 0 {
			return AircraftCategory{}, ErrAircraftCategoryInvalid
		}
		return UnavailableAircraftCategory(), nil
	}
	return NewAircraftCategory(state.AircraftCategory)
}
