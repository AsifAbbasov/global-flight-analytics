package weatherprovider

import domainweather "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/weather"

func CurrentWeatherRequestKey(
	request domainweather.CurrentWeatherRequest,
) string {
	return currentWeatherRequestKey(
		request,
	)
}
