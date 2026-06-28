package traffic

import (
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/airplaneslive"
	integrationcommon "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/common"
)

type Service struct {
	airplanesLive *airplaneslive.Client
}

func NewService() *Service {
	config := integrationcommon.DefaultHTTPClientConfig(
		airplaneslive.BaseURL,
	)

	return &Service{
		airplanesLive: airplaneslive.NewClient(config),
	}
}
