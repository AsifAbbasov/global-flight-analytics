package extractor

import (
	"errors"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

var (
	ErrTemporalBuilderRequired     = errors.New("temporal feature builder is required")
	ErrGeographicalBuilderRequired = errors.New("geographical feature builder is required")
	ErrOperationalBuilderRequired  = errors.New("operational feature builder is required")
	ErrTrajectoryBuilderRequired   = errors.New("trajectory feature builder is required")

	ErrTrajectoryIDRequired        = errors.New("trajectory id is required")
	ErrIdentityKeyRequired         = errors.New("trajectory identity key is required")
	ErrInvalidICAO24               = errors.New("invalid trajectory icao24")
	ErrTrajectoryStartTimeRequired = errors.New("trajectory start time is required")
	ErrTrajectoryEndTimeRequired   = errors.New("trajectory end time is required")
	ErrInvalidTrajectoryWindow     = errors.New("trajectory end time is before start time")
	ErrAsOfTimeRequired            = errors.New("feature as-of time is required")
	ErrAsOfBeforeTrajectoryEnd     = errors.New("feature as-of time is before trajectory end time")
	ErrTrajectoryEvidenceRequired  = errors.New("trajectory points or segments are required")
)

type GroupBuildError struct {
	Group flightfeatures.FeatureGroup
	Err   error
}

func (err *GroupBuildError) Error() string {
	if err == nil {
		return "feature group build failed"
	}

	return fmt.Sprintf(
		"build %s feature group: %v",
		err.Group,
		err.Err,
	)
}

func (err *GroupBuildError) Unwrap() error {
	if err == nil {
		return nil
	}

	return err.Err
}
