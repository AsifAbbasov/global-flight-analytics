package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/dataquality"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

func TestDataQualityRepositoryResolvesDatabaseGeneratedLiveParentByObservationIdentity(
	t *testing.T,
) {
	fixture := newQualityAssociationFixture(t)

	mustExecQualitySQL(
		t,
		fixture.pool,
		`
			CREATE TABLE flight_states (
				id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
				source_name text NOT NULL,
				icao24 text NOT NULL,
				observed_at timestamptz NOT NULL,
				UNIQUE (source_name, icao24, observed_at)
			);

			CREATE TABLE data_quality_reports (
				id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
				state_id uuid NOT NULL
					REFERENCES flight_states(id)
					ON DELETE CASCADE,
				flight_state_id uuid NOT NULL
					REFERENCES flight_states(id)
					ON DELETE CASCADE,
				validation_status text NOT NULL,
				completeness text NOT NULL,
				confidence text NOT NULL,
				score numeric NOT NULL DEFAULT 0,
				missing_fields text[] NOT NULL DEFAULT '{}',
				warnings_json jsonb NOT NULL DEFAULT '[]'::jsonb,
				calculated_at timestamptz NOT NULL DEFAULT now(),
				created_at timestamptz NOT NULL DEFAULT now(),
				CHECK (state_id = flight_state_id)
			);
		`,
	)

	observedAt := time.Date(
		2026,
		time.August,
		3,
		16,
		18,
		4,
		0,
		time.UTC,
	)
	sourceName := "airplanes.live"
	icao24 := "424382"

	var persistedID string
	err := fixture.pool.QueryRow(
		context.Background(),
		`
			INSERT INTO flight_states (
				source_name,
				icao24,
				observed_at
			)
			VALUES ($1, $2, $3)
			RETURNING id::text
		`,
		sourceName,
		icao24,
		observedAt,
	).Scan(&persistedID)
	if err != nil {
		t.Fatalf("insert live parent with database-generated id: %v", err)
	}

	quality := dataquality.DataQuality{
		ValidationStatus: dataquality.ValidationStatusValid,
		Completeness:     dataquality.CompletenessLevelComplete,
		Confidence:       dataquality.ConfidenceLevelHigh,
		Score:            1,
		MissingFields:    []string{},
		Warnings:         []dataquality.Warning{},
	}

	err = fixture.repository.SaveFlightStateQuality(
		context.Background(),
		flightstate.FlightState{
			ICAO24:     icao24,
			ObservedAt: observedAt,
			SourceName: sourceName,
		},
		quality,
	)
	if err != nil {
		t.Fatalf(
			"save quality for live state without application-assigned id: %v",
			err,
		)
	}

	var stateID string
	var flightStateID string
	err = fixture.pool.QueryRow(
		context.Background(),
		`
			SELECT
				state_id::text,
				flight_state_id::text
			FROM data_quality_reports
			LIMIT 1
		`,
	).Scan(
		&stateID,
		&flightStateID,
	)
	if err != nil {
		t.Fatalf("load persisted live quality association: %v", err)
	}

	if stateID != persistedID {
		t.Fatalf(
			"expected state id %s, got %s",
			persistedID,
			stateID,
		)
	}
	if flightStateID != persistedID {
		t.Fatalf(
			"expected flight state id %s, got %s",
			persistedID,
			flightStateID,
		)
	}
}
