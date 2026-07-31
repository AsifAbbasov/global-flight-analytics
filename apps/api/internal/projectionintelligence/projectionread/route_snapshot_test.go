package projectionread

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestLoadRouteRejectsPersistedRowIdentityMismatch(
	t *testing.T,
) {
	asOfTime := projectionReadTestAsOfTime()
	current := projectionReadTrajectory(
		"73aa02ab-7061-4e9e-a238-d32710371ee3",
		asOfTime,
	)
	route := projectionReadCompleteRoute(
		current,
		asOfTime,
	)
	payload, err := json.Marshal(route)
	if err != nil {
		t.Fatal(err)
	}

	client := &scriptedClient{
		rowQueue: []scriptedRow{
			{
				values: []any{
					"83aa02ab-7061-4e9e-a238-d32710371ee3",
					string(routecontract.SchemaVersionV1),
					asOfTime,
					asOfTime.UnixNano(),
					route.Provenance.InputFingerprint,
					string(route.Status),
					payload,
				},
			},
		},
	}
	source := newProjectionReadTestSource(
		t,
		client,
		&trajectoryRepositoryStub{},
	)

	_, err = source.LoadRoute(
		context.Background(),
		current.ID,
		asOfTime,
	)
	if !errors.Is(err, ErrRouteSnapshotInvalid) {
		t.Fatalf(
			"error = %v, want route snapshot sentinel",
			err,
		)
	}
}

func TestLoadRouteRejectsPersistedFingerprintMismatch(
	t *testing.T,
) {
	asOfTime := projectionReadTestAsOfTime()
	current := projectionReadTrajectory(
		"73aa02ab-7061-4e9e-a238-d32710371ee3",
		asOfTime,
	)
	route := projectionReadCompleteRoute(
		current,
		asOfTime,
	)
	payload, err := json.Marshal(route)
	if err != nil {
		t.Fatal(err)
	}

	client := &scriptedClient{
		rowQueue: []scriptedRow{
			{
				values: []any{
					current.ID,
					string(routecontract.SchemaVersionV1),
					asOfTime,
					asOfTime.UnixNano(),
					"sha256:" + strings.Repeat("f", 64),
					string(route.Status),
					payload,
				},
			},
		},
	}
	source := newProjectionReadTestSource(
		t,
		client,
		&trajectoryRepositoryStub{},
	)

	_, err = source.LoadRoute(
		context.Background(),
		current.ID,
		asOfTime,
	)
	if !errors.Is(err, ErrRouteSnapshotInvalid) {
		t.Fatalf(
			"error = %v, want route snapshot sentinel",
			err,
		)
	}
}

func TestLoadRouteRejectsInvalidPayloadContract(
	t *testing.T,
) {
	asOfTime := projectionReadTestAsOfTime()
	current := projectionReadTrajectory(
		"73aa02ab-7061-4e9e-a238-d32710371ee3",
		asOfTime,
	)
	route := projectionReadCompleteRoute(
		current,
		asOfTime,
	)
	route.Destination = nil
	payload, err := json.Marshal(route)
	if err != nil {
		t.Fatal(err)
	}

	client := &scriptedClient{
		rowQueue: []scriptedRow{
			{
				values: []any{
					current.ID,
					string(routecontract.SchemaVersionV1),
					asOfTime,
					asOfTime.UnixNano(),
					route.Provenance.InputFingerprint,
					string(route.Status),
					payload,
				},
			},
		},
	}
	source := newProjectionReadTestSource(
		t,
		client,
		&trajectoryRepositoryStub{},
	)

	_, err = source.LoadRoute(
		context.Background(),
		current.ID,
		asOfTime,
	)
	if !errors.Is(err, ErrRouteSnapshotInvalid) {
		t.Fatalf(
			"error = %v, want route snapshot sentinel",
			err,
		)
	}
}

func TestLoadCurrentTrajectoryRepairsUpdatedAtBeforeLatestPoint(
	t *testing.T,
) {
	asOfTime := projectionReadTestAsOfTime()
	item := projectionReadTrajectory(
		"73aa02ab-7061-4e9e-a238-d32710371ee3",
		asOfTime,
	)
	item.Points = nil
	item.UpdatedAt = item.StartTime

	client := &scriptedClient{
		rowsQueue: []*scriptedRows{
			{
				values: [][]any{
					projectionReadPointRow(
						"state-a",
						item.StartTime,
						40.40,
						49.80,
					),
					projectionReadPointRow(
						"state-b",
						item.EndTime,
						40.50,
						50.00,
					),
				},
			},
		},
	}
	source := newProjectionReadTestSource(
		t,
		client,
		&trajectoryRepositoryStub{
			items: map[string]trajectory.FlightTrajectory{
				item.ID: item,
			},
			errs: map[string]error{},
		},
	)

	result, err := source.LoadCurrentTrajectory(
		context.Background(),
		item.ID,
		asOfTime,
	)
	if err != nil {
		t.Fatalf(
			"LoadCurrentTrajectory() error = %v",
			err,
		)
	}
	if !result.UpdatedAt.Equal(result.EndTime) {
		t.Fatalf(
			"UpdatedAt=%s, want latest point %s",
			result.UpdatedAt,
			result.EndTime,
		)
	}
}

func TestRequiredTelemetryRejectsMissingMeasurements(
	t *testing.T,
) {
	validFloat := pgtype.Float8{
		Float64: 1,
		Valid:   true,
	}
	validBool := pgtype.Bool{
		Bool:  true,
		Valid: true,
	}
	missingFloat := pgtype.Float8{}
	missingBool := pgtype.Bool{}

	if completeRequiredTelemetry(
		missingFloat,
		validFloat,
		validFloat,
		validFloat,
		validFloat,
		validBool,
	) ||
		completeRequiredTelemetry(
			validFloat,
			validFloat,
			missingFloat,
			validFloat,
			validFloat,
			validBool,
		) ||
		completeRequiredTelemetry(
			validFloat,
			validFloat,
			validFloat,
			validFloat,
			validFloat,
			missingBool,
		) {
		t.Fatal(
			"missing telemetry was converted into a usable zero value",
		)
	}
}
