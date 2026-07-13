package calculator

import (
	"errors"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/registry"
)

var ErrMetricNotRegistered = errors.New("metric is not registered")

type Calculator struct {
	registry *registry.Registry
}

func New(
	reg *registry.Registry,
) *Calculator {
	return &Calculator{
		registry: reg,
	}
}

func (c *Calculator) Registry() *registry.Registry {
	return c.registry
}
