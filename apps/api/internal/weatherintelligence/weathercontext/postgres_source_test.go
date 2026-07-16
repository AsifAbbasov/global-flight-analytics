package weathercontext

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	domainweather "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/weather"
	"github.com/jackc/pgx/v5"
)

func TestPostgresSnapshotReaderReturnsBoundedSnapshot(
	t *testing.T,
) {
	t.Parallel()

	asOfTime := time.Date(
		2026,
		time.July,
		16,
		12,
		0,
		0,
		0,
		time.UTC,
	)
	expected := domainweather.CurrentSnapshot{
		Provider:                 domainweather.ProviderOpenMeteo,
		Latitude:                 40.4675,
		Longitude:                50.0467,
		ObservedAt:               asOfTime.Add(-5 * time.Minute),
		RetrievedAt:              asOfTime.Add(-4 * time.Minute),
		TemperatureCelsius:       27.5,
		RelativeHumidityPercent:  54,
		PrecipitationMillimeters: 0.2,
		RainMillimeters:          0.1,
		WeatherCode:              2,
		CloudCoverPercent:        30,
		SurfacePressureHPA:       1011.4,
		WindSpeedMetersPerSecond: 5.5,
		WindDirectionDegrees:     190,
		WindGustsMetersPerSecond: 8.2,
	}
	queryer := &fakeWeatherSnapshotQueryer{
		row: fakeWeatherSnapshotRow{
			snapshot: expected,
		},
	}
	reader, err := newPostgresSnapshotReader(
		queryer,
		DefaultPostgresSnapshotPolicy(),
	)
	if err != nil {
		t.Fatalf(
			"newPostgresSnapshotReader() error = %v",
			err,
		)
	}

	result, err := reader.GetLatestSnapshot(
		context.Background(),
		WeatherSnapshotRequest{
			Latitude:  40.47,
			Longitude: 50.05,
			AsOfTime:  asOfTime,
		},
	)
	if err != nil {
		t.Fatalf("GetLatestSnapshot() error = %v", err)
	}
	if result != expected {
		t.Fatalf(
			"GetLatestSnapshot() = %#v, want %#v",
			result,
			expected,
		)
	}
	if !strings.Contains(
		queryer.query,
		"retrieved_at <= $2",
	) || !strings.Contains(
		queryer.query,
		"observed_at <= $2",
	) {
		t.Fatalf(
			"query does not enforce as-of boundaries: %s",
			queryer.query,
		)
	}
	if len(queryer.args) != 5 {
		t.Fatalf(
			"query argument count = %d, want 5",
			len(queryer.args),
		)
	}
	if queryer.args[0] != domainweather.ProviderOpenMeteo ||
		!queryer.args[1].(time.Time).Equal(asOfTime) ||
		queryer.args[2] != 40.47 ||
		queryer.args[3] != 50.05 ||
		queryer.args[4] != 1.0 {
		t.Fatalf(
			"query arguments = %#v",
			queryer.args,
		)
	}
}

func TestPostgresSnapshotReaderMapsNoRows(
	t *testing.T,
) {
	t.Parallel()

	reader, err := newPostgresSnapshotReader(
		&fakeWeatherSnapshotQueryer{
			row: fakeWeatherSnapshotRow{
				err: pgx.ErrNoRows,
			},
		},
		DefaultPostgresSnapshotPolicy(),
	)
	if err != nil {
		t.Fatalf(
			"newPostgresSnapshotReader() error = %v",
			err,
		)
	}

	_, err = reader.GetLatestSnapshot(
		context.Background(),
		WeatherSnapshotRequest{
			Latitude:  40.47,
			Longitude: 50.05,
			AsOfTime:  time.Now().UTC(),
		},
	)
	if !errors.Is(err, ErrWeatherNotFound) {
		t.Fatalf(
			"GetLatestSnapshot() error = %v, want ErrWeatherNotFound",
			err,
		)
	}
}

type fakeWeatherSnapshotQueryer struct {
	query string
	args  []any
	row   pgx.Row
}

func (
	queryer *fakeWeatherSnapshotQueryer,
) QueryRow(
	_ context.Context,
	query string,
	args ...any,
) pgx.Row {
	queryer.query = query
	queryer.args = append([]any(nil), args...)
	return queryer.row
}

type fakeWeatherSnapshotRow struct {
	snapshot domainweather.CurrentSnapshot
	err      error
}

func (row fakeWeatherSnapshotRow) Scan(
	destinations ...any,
) error {
	if row.err != nil {
		return row.err
	}
	if len(destinations) != 15 {
		return errors.New(
			"unexpected Weather Context scan destination count",
		)
	}

	*destinations[0].(*string) = row.snapshot.Provider
	*destinations[1].(*float64) = row.snapshot.Latitude
	*destinations[2].(*float64) = row.snapshot.Longitude
	*destinations[3].(*time.Time) = row.snapshot.ObservedAt
	*destinations[4].(*time.Time) = row.snapshot.RetrievedAt
	*destinations[5].(*float64) = row.snapshot.TemperatureCelsius
	*destinations[6].(*int) = row.snapshot.RelativeHumidityPercent
	*destinations[7].(*float64) = row.snapshot.PrecipitationMillimeters
	*destinations[8].(*float64) = row.snapshot.RainMillimeters
	*destinations[9].(*int) = row.snapshot.WeatherCode
	*destinations[10].(*int) = row.snapshot.CloudCoverPercent
	*destinations[11].(*float64) = row.snapshot.SurfacePressureHPA
	*destinations[12].(*float64) = row.snapshot.WindSpeedMetersPerSecond
	*destinations[13].(*int) = row.snapshot.WindDirectionDegrees
	*destinations[14].(*float64) = row.snapshot.WindGustsMetersPerSecond
	return nil
}
