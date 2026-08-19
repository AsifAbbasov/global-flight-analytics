package adsblol

import (
	"context"
	"errors"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/providerbatch"
)

var ErrClientRequired = errors.New("adsb.lol client is required")

type Provider struct {
	client *Client
}

func NewProvider(client *Client) *Provider {
	if client == nil {
		return nil
	}
	return &Provider{client: client}
}

func (provider *Provider) SourceName() string {
	return sourceName
}

func (provider *Provider) LoadByPoint(
	ctx context.Context,
	latitude float64,
	longitude float64,
	radius int,
) ([]flightstate.FlightState, error) {
	states, _, err := provider.LoadByPointWithBatchEvidence(
		ctx,
		latitude,
		longitude,
		radius,
	)
	return states, err
}

func (provider *Provider) LoadByPointWithBatchEvidence(
	ctx context.Context,
	latitude float64,
	longitude float64,
	radius int,
) ([]flightstate.FlightState, providerbatch.Evidence, error) {
	if provider == nil || provider.client == nil {
		return nil, providerbatch.Evidence{}, ErrClientRequired
	}
	response, err := provider.client.GetByPoint(ctx, latitude, longitude, radius)
	if err != nil {
		return nil, providerbatch.Evidence{}, fmt.Errorf(
			"load adsb.lol traffic by point: %w",
			err,
		)
	}
	return MapStateResponseWithEvidence(response)
}
