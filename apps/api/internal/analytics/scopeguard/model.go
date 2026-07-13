package scopeguard

import (
	"fmt"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

type Decision struct {
	Capability  trajectoryeligibility.Capability
	Allowed     bool
	Reasons     []trajectoryeligibility.ReasonCode
	EvaluatedAt time.Time
}

func (decision Decision) HasReason(reason trajectoryeligibility.ReasonCode) bool {
	for _, current := range decision.Reasons {
		if current == reason {
			return true
		}
	}
	return false
}

type DeniedTrajectory struct {
	Trajectory trajectory.FlightTrajectory
	Decision   Decision
}

type FilterResult struct {
	Allowed     []trajectory.FlightTrajectory
	Denied      []DeniedTrajectory
	EvaluatedAt time.Time
}

func (result FilterResult) AllowedCount() int { return len(result.Allowed) }
func (result FilterResult) DeniedCount() int  { return len(result.Denied) }

type DeniedError struct {
	Capability  trajectoryeligibility.Capability
	Reasons     []trajectoryeligibility.ReasonCode
	IdentityKey string
	ICAO24      string
	EvaluatedAt time.Time
}

func (err *DeniedError) Error() string {
	if err == nil {
		return ErrDenied.Error()
	}

	reasons := make([]string, 0, len(err.Reasons))
	for _, reason := range err.Reasons {
		reasons = append(reasons, string(reason))
	}

	reasonText := "unspecified"
	if len(reasons) > 0 {
		reasonText = strings.Join(reasons, ",")
	}

	return fmt.Sprintf(
		"%s: capability=%s identity_key=%q icao24=%q reasons=%s evaluated_at=%s",
		ErrDenied.Error(),
		err.Capability,
		err.IdentityKey,
		err.ICAO24,
		reasonText,
		err.EvaluatedAt.UTC().Format(time.RFC3339Nano),
	)
}

func (err *DeniedError) Unwrap() error { return ErrDenied }

func newDeniedError(item trajectory.FlightTrajectory, decision Decision) *DeniedError {
	return &DeniedError{
		Capability: decision.Capability,
		Reasons: append(
			[]trajectoryeligibility.ReasonCode(nil),
			decision.Reasons...,
		),
		IdentityKey: item.IdentityKey,
		ICAO24:      item.ICAO24,
		EvaluatedAt: decision.EvaluatedAt,
	}
}
