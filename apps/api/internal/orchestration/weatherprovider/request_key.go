package weatherprovider

import "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/openmeteo"

func CurrentWeatherRequestKey(
	request openmeteo.CurrentWeatherRequest,
) string {
	return currentWeatherRequestKey(
		request,
	)
}
