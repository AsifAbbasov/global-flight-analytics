package airplaneslive

import (
	"context"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

type Provider struct {
	client *Client
}

func NewProvider(client *Client) *Provider {
	return &Provider{
		client: client,
	}
}

func (p *Provider) LoadByCallsign(
	ctx context.Context,
	callsign string,
) ([]flightstate.FlightState, error) {
	result, err := p.client.GetByCallsign(ctx, callsign)
	if err != nil {
		return nil, fmt.Errorf("load airplanes live traffic by callsign: %w", err)
	}

	return MapStateResponse(result), nil
}
