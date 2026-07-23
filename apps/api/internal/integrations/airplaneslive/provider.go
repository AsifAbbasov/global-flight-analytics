package airplaneslive

import (
	"context"
	"errors"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

var ErrClientRequired = errors.New(
	"airplanes.live client is required",
)

type Provider struct {
	client *Client
}

func NewProvider(client *Client) *Provider {
	if client == nil {
		return nil
	}
	return &Provider{
		client: client,
	}
}

func (p *Provider) requireClient() (*Client, error) {
	if p == nil || p.client == nil {
		return nil, ErrClientRequired
	}
	return p.client, nil
}

func (p *Provider) SourceName() string {
	return sourceName
}

func (p *Provider) LoadByCallsign(
	ctx context.Context,
	callsign string,
) ([]flightstate.FlightState, error) {
	client, err := p.requireClient()
	if err != nil {
		return nil, err
	}
	result, err := client.GetByCallsign(ctx, callsign)
	if err != nil {
		return nil, fmt.Errorf(
			"load airplanes live traffic by callsign: %w",
			err,
		)
	}

	return MapStateResponse(result), nil
}

func (p *Provider) LoadByPoint(
	ctx context.Context,
	latitude float64,
	longitude float64,
	radius int,
) ([]flightstate.FlightState, error) {
	client, err := p.requireClient()
	if err != nil {
		return nil, err
	}
	result, err := client.GetByPoint(
		ctx,
		latitude,
		longitude,
		radius,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"load airplanes live traffic by point: %w",
			err,
		)
	}

	return MapStateResponse(result), nil
}
