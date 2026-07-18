package main

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/extractor"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/featurepipeline"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/featurestore"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

func TestParseCommandOptionsAcceptsTrajectoryID(t *testing.T) {
	options, err := parseCommandOptions(
		[]string{
			"--trajectory-id",
			"8f42d9a8-5ad6-4a90-a38c-5f2c6348a318",
		},
		&bytes.Buffer{},
	)
	if err != nil {
		t.Fatal(err)
	}
	if options.TrajectoryID !=
		"8f42d9a8-5ad6-4a90-a38c-5f2c6348a318" {
		t.Fatalf("trajectory ID = %q", options.TrajectoryID)
	}
}

func TestParseCommandOptionsRejectsAmbiguousSelector(t *testing.T) {
	_, err := parseCommandOptions(
		[]string{
			"--trajectory-id",
			"8f42d9a8-5ad6-4a90-a38c-5f2c6348a318",
			"--icao24",
			"ABC123",
		},
		&bytes.Buffer{},
	)
	if err == nil {
		t.Fatal("expected selector error")
	}
}

func TestOperationDefaultsAsOfTimeToTrajectoryEnd(t *testing.T) {
	endTime := time.Date(2026, time.July, 18, 10, 0, 0, 0, time.UTC)
	item := trajectory.FlightTrajectory{
		ID:          "8f42d9a8-5ad6-4a90-a38c-5f2c6348a318",
		IdentityKey: "identity-key",
		ICAO24:      "ABC123",
		StartTime:   endTime.Add(-time.Hour),
		EndTime:     endTime,
	}
	reader := &fakeTrajectoryReader{byID: item}
	processor := &fakeFeatureProcessor{}
	operation, err := newMaterializationOperation(reader, processor)
	if err != nil {
		t.Fatal(err)
	}

	report, err := operation.Execute(
		context.Background(),
		commandOptions{TrajectoryID: item.ID},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !processor.request.AsOfTime.Equal(endTime) {
		t.Fatalf(
			"as-of time = %s, want %s",
			processor.request.AsOfTime,
			endTime,
		)
	}
	if report.SnapshotID != "feature-record-test" {
		t.Fatalf("snapshot ID = %q", report.SnapshotID)
	}
}

func TestOperationRejectsAsOfBeforeTrajectoryEnd(t *testing.T) {
	endTime := time.Date(2026, time.July, 18, 10, 0, 0, 0, time.UTC)
	operation, err := newMaterializationOperation(
		&fakeTrajectoryReader{
			byID: trajectory.FlightTrajectory{
				ID:          "8f42d9a8-5ad6-4a90-a38c-5f2c6348a318",
				IdentityKey: "identity-key",
				ICAO24:      "ABC123",
				StartTime:   endTime.Add(-time.Hour),
				EndTime:     endTime,
			},
		},
		&fakeFeatureProcessor{},
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = operation.Execute(
		context.Background(),
		commandOptions{
			TrajectoryID: "8f42d9a8-5ad6-4a90-a38c-5f2c6348a318",
			AsOfTime:     endTime.Add(-time.Second),
		},
	)
	if !errors.Is(err, errAsOfBeforeTrajectoryEnd) {
		t.Fatalf("error = %v, want %v", err, errAsOfBeforeTrajectoryEnd)
	}
}

type fakeTrajectoryReader struct {
	byID   trajectory.FlightTrajectory
	latest trajectory.FlightTrajectory
	err    error
}

func (reader *fakeTrajectoryReader) GetTrajectoryByID(
	context.Context,
	string,
) (trajectory.FlightTrajectory, error) {
	if reader.err != nil {
		return trajectory.FlightTrajectory{}, reader.err
	}
	return reader.byID, nil
}

func (reader *fakeTrajectoryReader) GetLatestTrajectoryByICAO24(
	context.Context,
	string,
) (trajectory.FlightTrajectory, error) {
	if reader.err != nil {
		return trajectory.FlightTrajectory{}, reader.err
	}
	return reader.latest, nil
}

type fakeFeatureProcessor struct {
	request extractor.Request
	err     error
}

func (processor *fakeFeatureProcessor) Process(
	_ context.Context,
	request extractor.Request,
) (featurepipeline.Result, error) {
	processor.request = request
	if processor.err != nil {
		return featurepipeline.Result{}, processor.err
	}

	features := flightfeatures.FlightFeatures{
		SchemaVersion: flightfeatures.SchemaVersionV1,
		TrajectoryID:  request.Trajectory.ID,
		IdentityKey:   request.Trajectory.IdentityKey,
		ICAO24:        request.Trajectory.ICAO24,
		Window: flightfeatures.FeatureWindow{
			StartTime: request.Trajectory.StartTime,
			EndTime:   request.Trajectory.EndTime,
			AsOfTime:  request.AsOfTime,
		},
		ExtractedAt: request.AsOfTime,
		Quality: flightfeatures.FeatureQuality{
			Status: flightfeatures.ValidationStatusLimited,
		},
		Provenance: flightfeatures.FeatureProvenance{
			InputFingerprint: "sha256:test",
		},
	}

	return featurepipeline.Result{
		PipelineVersion: featurepipeline.Version,
		Features:        features,
		Record: featurestore.Record{
			ID:               "feature-record-test",
			InputFingerprint: "sha256:test",
			Features:         features,
			StoredAt:         request.AsOfTime,
		},
	}, nil
}
