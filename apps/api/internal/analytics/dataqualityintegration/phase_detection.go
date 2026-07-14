package dataqualityintegration

import (
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/dataqualitycontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/flightphase"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

const (
	PermissionReasonPhaseDetectionNoClassifiedPoints = "phase_detection_no_classified_points"
	PermissionReasonPhaseDetectionOnlyUnknownPoints  = "phase_detection_only_unknown_points"

	LimitationCodePhaseDetectionPartial     = "phase_detection_partial_evidence"
	LimitationCodePhaseDetectionUnavailable = "phase_detection_unavailable"
)

func evaluatePhaseDetection(
	items []trajectory.FlightTrajectory,
) (
	dataqualitycontract.Permission,
	[]dataqualitycontract.Notice,
	error,
) {
	detector := flightphase.NewDefault()

	classifiedPointCount := 0
	knownPhasePointCount := 0
	unknownPhasePointCount := 0
	excludedPointCount := 0
	limitedTrajectoryCount := 0

	for index, item := range items {
		result, err := detector.Detect(item)
		if err != nil {
			return dataqualitycontract.Permission{},
				nil,
				fmt.Errorf(
					"detect trajectory phase at index %d: %w",
					index,
					err,
				)
		}

		classifiedPointCount += result.ClassifiedPointCount
		excludedPointCount += result.ExcludedPointCount
		if len(result.Limitations) > 0 {
			limitedTrajectoryCount++
		}

		for _, point := range result.Points {
			if point.Phase == flightphase.PhaseUnknown {
				unknownPhasePointCount++
				continue
			}

			knownPhasePointCount++
		}
	}

	switch {
	case classifiedPointCount == 0:
		permission, err := dataqualitycontract.DeniedPermission(
			PermissionReasonPhaseDetectionNoClassifiedPoints,
		)
		if err != nil {
			return dataqualitycontract.Permission{},
				nil,
				fmt.Errorf(
					"build unavailable phase-detection permission: %w",
					err,
				)
		}

		return permission,
			[]dataqualitycontract.Notice{
				{
					Code:    LimitationCodePhaseDetectionUnavailable,
					Message: "Basic flight-phase detection could not classify any retained trajectory point.",
				},
			},
			nil

	case knownPhasePointCount == 0:
		permission, err := dataqualitycontract.DeniedPermission(
			PermissionReasonPhaseDetectionOnlyUnknownPoints,
		)
		if err != nil {
			return dataqualitycontract.Permission{},
				nil,
				fmt.Errorf(
					"build unknown-only phase-detection permission: %w",
					err,
				)
		}

		return permission,
			[]dataqualitycontract.Notice{
				{
					Code: LimitationCodePhaseDetectionUnavailable,
					Message: fmt.Sprintf(
						"Basic flight-phase detection classified %d retained points, but every point remained unknown because the required operational signals were insufficient.",
						classifiedPointCount,
					),
				},
			},
			nil
	}

	permission := dataqualitycontract.AllowedPermission()
	limitations := make(
		[]dataqualitycontract.Notice,
		0,
		1,
	)

	if unknownPhasePointCount > 0 ||
		excludedPointCount > 0 ||
		limitedTrajectoryCount > 0 {
		limitations = append(
			limitations,
			dataqualitycontract.Notice{
				Code: LimitationCodePhaseDetectionPartial,
				Message: fmt.Sprintf(
					"Basic flight-phase detection produced %d known-phase points and %d unknown-phase points; %d points were excluded and %d trajectories reported detector limitations.",
					knownPhasePointCount,
					unknownPhasePointCount,
					excludedPointCount,
					limitedTrajectoryCount,
				),
			},
		)
	}

	return permission, limitations, nil
}
