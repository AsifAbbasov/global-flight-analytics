package opensky

import (
	"errors"
	"fmt"
	"math"
	"net/url"
	"strconv"
)

var (
	ErrBoundingBoxLatitudeInvalid  = errors.New("OpenSky bounding box latitude is invalid")
	ErrBoundingBoxLongitudeInvalid = errors.New("OpenSky bounding box longitude is invalid")
	ErrBoundingBoxOrderInvalid     = errors.New("OpenSky bounding box minimum must be less than maximum")
)

type BoundingBox struct {
	MinimumLatitude  float64
	MaximumLatitude  float64
	MinimumLongitude float64
	MaximumLongitude float64
}

func (box BoundingBox) Validate() error {
	if !finiteRange(box.MinimumLatitude, -90, 90) ||
		!finiteRange(box.MaximumLatitude, -90, 90) {
		return ErrBoundingBoxLatitudeInvalid
	}
	if !finiteRange(box.MinimumLongitude, -180, 180) ||
		!finiteRange(box.MaximumLongitude, -180, 180) {
		return ErrBoundingBoxLongitudeInvalid
	}
	if box.MinimumLatitude >= box.MaximumLatitude ||
		box.MinimumLongitude >= box.MaximumLongitude {
		return ErrBoundingBoxOrderInvalid
	}
	return nil
}

func (box BoundingBox) AreaSquareDegrees() (float64, error) {
	if err := box.Validate(); err != nil {
		return 0, err
	}
	return (box.MaximumLatitude - box.MinimumLatitude) *
		(box.MaximumLongitude - box.MinimumLongitude), nil
}

func (box BoundingBox) AddTo(values url.Values) error {
	if err := box.Validate(); err != nil {
		return err
	}
	values.Set("lamin", strconv.FormatFloat(box.MinimumLatitude, 'f', -1, 64))
	values.Set("lamax", strconv.FormatFloat(box.MaximumLatitude, 'f', -1, 64))
	values.Set("lomin", strconv.FormatFloat(box.MinimumLongitude, 'f', -1, 64))
	values.Set("lomax", strconv.FormatFloat(box.MaximumLongitude, 'f', -1, 64))
	return nil
}

func (box BoundingBox) EstimatedStateCreditCost() (int, error) {
	area, err := box.AreaSquareDegrees()
	if err != nil {
		return 0, err
	}
	switch {
	case area <= 25:
		return 1, nil
	case area <= 100:
		return 2, nil
	case area <= 400:
		return 3, nil
	default:
		return 4, nil
	}
}

func finiteRange(value float64, minimum float64, maximum float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) &&
		value >= minimum && value <= maximum
}

func (box BoundingBox) String() string {
	return fmt.Sprintf(
		"[%g,%g]x[%g,%g]",
		box.MinimumLatitude,
		box.MaximumLatitude,
		box.MinimumLongitude,
		box.MaximumLongitude,
	)
}
