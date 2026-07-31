package projectionread

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

func scanRouteSnapshotResult(
	scanner rowScanner,
	requestedTrajectoryID string,
	requestedAsOfTime time.Time,
) (routecontract.Result, error) {
	var (
		rowTrajectoryID     string
		rowSchemaVersion    string
		rowAsOfTime         time.Time
		rowAsOfTimeUnixNano int64
		rowInputFingerprint string
		rowRouteStatus      string
		payload             []byte
	)

	if err := scanner.Scan(
		&rowTrajectoryID,
		&rowSchemaVersion,
		&rowAsOfTime,
		&rowAsOfTimeUnixNano,
		&rowInputFingerprint,
		&rowRouteStatus,
		&payload,
	); err != nil {
		return routecontract.Result{}, err
	}

	exactRowAsOfTime := time.Unix(
		0,
		rowAsOfTimeUnixNano,
	).UTC()
	requestedTrajectoryID = strings.TrimSpace(
		requestedTrajectoryID,
	)
	requestedAsOfTime = requestedAsOfTime.UTC()

	if strings.TrimSpace(rowTrajectoryID) !=
		requestedTrajectoryID ||
		rowSchemaVersion !=
			string(routecontract.SchemaVersionV1) ||
		!rowAsOfTime.UTC().Equal(exactRowAsOfTime) ||
		exactRowAsOfTime.After(requestedAsOfTime) {
		return routecontract.Result{},
			fmt.Errorf(
				"%w: persisted route row identity or time mirror is invalid",
				ErrRouteSnapshotInvalid,
			)
	}

	var result routecontract.Result
	if err := json.Unmarshal(
		payload,
		&result,
	); err != nil {
		return routecontract.Result{},
			fmt.Errorf(
				"%w: decode route payload: %w",
				ErrRouteSnapshotInvalid,
				err,
			)
	}
	report := routecontract.Validate(result)
	if report.Status !=
		routecontract.ValidationStatusValid {
		return routecontract.Result{},
			fmt.Errorf(
				"%w: route payload contract is invalid: %#v",
				ErrRouteSnapshotInvalid,
				report.Issues,
			)
	}

	if strings.TrimSpace(result.TrajectoryID) !=
		requestedTrajectoryID ||
		string(result.SchemaVersion) !=
			rowSchemaVersion ||
		!result.Window.AsOfTime.UTC().Equal(
			exactRowAsOfTime,
		) ||
		result.Provenance.InputFingerprint !=
			strings.TrimSpace(rowInputFingerprint) ||
		string(result.Status) !=
			strings.TrimSpace(rowRouteStatus) {
		return routecontract.Result{},
			fmt.Errorf(
				"%w: route payload does not match persisted row metadata",
				ErrRouteSnapshotInvalid,
			)
	}

	return result.Clone(), nil
}
