package weather

import (
	"errors"
	"testing"
	"time"
)

func validWeatherSnapshot() CurrentSnapshot {
	now := time.Now().UTC()
	return CurrentSnapshot{Provider: ProviderOpenMeteo, Latitude: 40, Longitude: 49, ObservedAt: now, RetrievedAt: now, TemperatureCelsius: 20, RelativeHumidityPercent: 50, CloudCoverPercent: 20, SurfacePressureHPA: 1013}
}

func TestWeatherValidateRejectsImpossiblePercentagesAndNegativePrecipitation(t *testing.T) {
	value := validWeatherSnapshot()
	value.RelativeHumidityPercent = -1
	if err := value.Validate(); !errors.Is(err, ErrWeatherHumidityInvalid) {
		t.Fatalf("humidity error = %v", err)
	}
	value = validWeatherSnapshot()
	value.PrecipitationMillimeters = -0.1
	if err := value.Validate(); !errors.Is(err, ErrWeatherPrecipitationInvalid) {
		t.Fatalf("precipitation error = %v", err)
	}
}
