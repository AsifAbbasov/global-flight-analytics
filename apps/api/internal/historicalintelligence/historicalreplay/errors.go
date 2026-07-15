package historicalreplay

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrMaterializerRequired = errors.New(
		"historical replay materializer is required",
	)
	ErrMaximumWindowCountInvalid = errors.New(
		"historical replay maximum window count is invalid",
	)
	ErrNoReplayWindow = errors.New(
		"historical replay contains no complete window",
	)
)

type WindowCountExceededError struct {
	Count   int
	Maximum int
}

func (err *WindowCountExceededError) Error() string {
	if err == nil {
		return "historical replay window count exceeded"
	}

	return fmt.Sprintf(
		"historical replay window count %d exceeds maximum %d",
		err.Count,
		err.Maximum,
	)
}

type WindowError struct {
	Sequence  int
	StartTime time.Time
	EndTime   time.Time
	Err       error
}

func (err *WindowError) Error() string {
	if err == nil {
		return "historical replay window failed"
	}

	return fmt.Sprintf(
		"historical replay window %d [%s,%s) failed: %v",
		err.Sequence,
		err.StartTime.UTC().Format(time.RFC3339),
		err.EndTime.UTC().Format(time.RFC3339),
		err.Err,
	)
}

func (err *WindowError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Err
}
