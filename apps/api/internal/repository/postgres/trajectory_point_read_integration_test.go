package postgres

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var trajectoryPointSchemaCounter uint64

func TestTrajectoryPointReadPreservesNullableOperationalTelemetry(t *testing.T) {
	databaseURL := strings.TrimSpace(os.Getenv("TEST_DATABASE_URL"))
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set; skipping PostgreSQL integration test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	bootstrap, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect PostgreSQL: %v", err)
	}
	schemaName := fmt.Sprintf(
		"trajectory_point_read_%d_%d",
		time.Now().UnixNano(),
		atomic.AddUint64(&trajectoryPointSchemaCounter, 1),
	)
	quotedSchema := pgx.Identifier{schemaName}.Sanitize()
	if _, err := bootstrap.Exec(ctx, "CREATE SCHEMA "+quotedSchema); err != nil {
		_ = bootstrap.Close(ctx)
		t.Fatalf("create schema: %v", err)
	}

	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		_ = bootstrap.Close(ctx)
		t.Fatalf("parse pool config: %v", err)
	}
	if poolConfig.ConnConfig.RuntimeParams == nil {
		poolConfig.ConnConfig.RuntimeParams = make(map[string]string)
	}
	poolConfig.ConnConfig.RuntimeParams["search_path"] = schemaName
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		_ = bootstrap.Close(ctx)
		t.Fatalf("create pool: %v", err)
	}
	t.Cleanup(func() {
		pool.Close()
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		_, _ = bootstrap.Exec(cleanupCtx, "DROP SCHEMA IF EXISTS "+quotedSchema+" CASCADE")
		_ = bootstrap.Close(cleanupCtx)
	})

	if _, err := pool.Exec(ctx, `
		CREATE TABLE flight_states (
			id uuid PRIMARY KEY,
			flight_id uuid,
			aircraft_id uuid,
			icao24 varchar(6) NOT NULL,
			callsign text,
			latitude double precision,
			longitude double precision,
			barometric_altitude_m double precision,
			barometric_altitude_status text NOT NULL,
			geometric_altitude_m double precision,
			geometric_altitude_status text NOT NULL,
			velocity_mps double precision,
			heading_degrees double precision,
			vertical_rate_mps double precision,
			on_ground boolean,
			origin_country text,
			observed_at timestamptz NOT NULL,
			source_name text NOT NULL
		)
	`); err != nil {
		t.Fatalf("create flight_states: %v", err)
	}

	flightID := "11111111-1111-4111-8111-111111111111"
	start := time.Date(2026, time.July, 27, 8, 0, 0, 0, time.UTC)
	insert := `INSERT INTO flight_states (
		id, flight_id, icao24, latitude, longitude,
		barometric_altitude_m, barometric_altitude_status,
		geometric_altitude_m, geometric_altitude_status,
		velocity_mps, heading_degrees, vertical_rate_mps, on_ground,
		observed_at, source_name
	) VALUES ($1::uuid, $2::uuid, 'ABC123', 40, 49, NULL, 'unavailable', NULL, 'unavailable', $3, $4, $5, $6, $7, 'test')`
	rows := []struct {
		id         string
		velocity   any
		heading    any
		vertical   any
		onGround   any
		observedAt time.Time
	}{
		{"00000000-0000-4000-8000-000000000001", nil, 0.0, nil, nil, start},
		{"00000000-0000-4000-8000-000000000002", 0.0, nil, 0.0, false, start.Add(time.Second)},
		{"00000000-0000-4000-8000-000000000003", 50.0, 90.0, 1.0, false, start.Add(-time.Second)},
		{"00000000-0000-4000-8000-000000000004", 50.0, 90.0, 1.0, false, start.Add(2 * time.Second)},
	}
	for _, row := range rows {
		if _, err := pool.Exec(
			ctx,
			insert,
			row.id,
			flightID,
			row.velocity,
			row.heading,
			row.vertical,
			row.onGround,
			row.observedAt,
		); err != nil {
			t.Fatalf("insert point %s: %v", row.id, err)
		}
	}

	repository := NewTrajectoryReadRepository(pool)
	points, err := repository.listTrajectoryPoints(ctx, trajectory.FlightTrajectory{
		FlightID:  flightID,
		ICAO24:    "ABC123",
		StartTime: start,
		EndTime:   start.Add(time.Second),
	})
	if err != nil {
		t.Fatalf("list trajectory points: %v", err)
	}
	if len(points) != 2 {
		t.Fatalf("point count = %d, want 2", len(points))
	}
	first := points[0]
	second := points[1]
	if !first.ObservedAt.Equal(start) || !second.ObservedAt.Equal(start.Add(time.Second)) {
		t.Fatalf("points are not ordered: %#v", points)
	}
	if !first.TelemetryAvailabilityKnown || first.HasVelocity() || !first.HasHeading() || first.HasVerticalRate() || first.HasOnGroundState() {
		t.Fatalf("first nullable telemetry changed: %#v", first)
	}
	if !second.TelemetryAvailabilityKnown || !second.HasVelocity() || second.HasHeading() || !second.HasVerticalRate() || !second.HasOnGroundState() {
		t.Fatalf("second nullable telemetry changed: %#v", second)
	}
	if second.VelocityMPS != 0 || second.VerticalRateMPS != 0 || second.OnGround {
		t.Fatalf("legitimate zero telemetry changed: %#v", second)
	}
}
