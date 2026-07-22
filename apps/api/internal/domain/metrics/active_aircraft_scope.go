package metrics

import (
	"errors"
	"fmt"
	"math"
)

var (
	ErrActiveAircraftQueryScopeInvalid = errors.New(
		"active aircraft query scope is invalid",
	)
	ErrActiveAircraftBoundsInvalid = errors.New(
		"active aircraft bounds are invalid",
	)
)

type ActiveAircraftQueryScopeType string

const (
	ActiveAircraftQueryScopeGlobal ActiveAircraftQueryScopeType = "global"
	ActiveAircraftQueryScopeBounds ActiveAircraftQueryScopeType = "bounds"
)

type ActiveAircraftQueryScope struct {
	Type   ActiveAircraftQueryScopeType
	Bounds Bounds
}

func NewGlobalActiveAircraftQueryScope() ActiveAircraftQueryScope {
	return ActiveAircraftQueryScope{Type: ActiveAircraftQueryScopeGlobal}
}

func NewBoundedActiveAircraftQueryScope(
	bounds Bounds,
) (ActiveAircraftQueryScope, error) {
	if err := bounds.Validate(); err != nil {
		return ActiveAircraftQueryScope{}, err
	}
	return ActiveAircraftQueryScope{
		Type:   ActiveAircraftQueryScopeBounds,
		Bounds: bounds,
	}, nil
}

func (scope ActiveAircraftQueryScope) Validate() error {
	switch scope.Type {
	case "", ActiveAircraftQueryScopeGlobal:
		return nil
	case ActiveAircraftQueryScopeBounds:
		return scope.Bounds.Validate()
	default:
		return fmt.Errorf(
			"%w: %q",
			ErrActiveAircraftQueryScopeInvalid,
			scope.Type,
		)
	}
}

func (scope ActiveAircraftQueryScope) IsBounded() bool {
	return scope.Type == ActiveAircraftQueryScopeBounds
}

func (bounds Bounds) Validate() error {
	values := []float64{
		bounds.MinLatitude,
		bounds.MaxLatitude,
		bounds.MinLongitude,
		bounds.MaxLongitude,
	}
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return ErrActiveAircraftBoundsInvalid
		}
	}
	if bounds.MinLatitude < -90 ||
		bounds.MaxLatitude > 90 ||
		bounds.MinLatitude > bounds.MaxLatitude ||
		bounds.MinLongitude < -180 ||
		bounds.MaxLongitude > 180 ||
		bounds.MinLongitude > bounds.MaxLongitude {
		return ErrActiveAircraftBoundsInvalid
	}
	return nil
}

func (query ActiveAircraftQuery) Validate() error {
	return query.Scope.Validate()
}
