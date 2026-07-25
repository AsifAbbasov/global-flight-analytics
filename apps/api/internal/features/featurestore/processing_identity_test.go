package featurestore

import (
	"context"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

func TestMemoryStoreSeparatesProcessingVersions(t *testing.T) {
	store := NewMemory(MemoryConfig{})
	asOf := time.Date(2026, time.July, 25, 12, 0, 0, 0, time.UTC)

	first := validStoredFeatures("trajectory-one", asOf, "a")
	first.Provenance.ProcessingVersion =
		flightfeatures.CurrentProcessingVersion
	second := first.Clone()
	second.Provenance.ProcessingVersion =
		"flight-feature-processing-pipeline-v2"

	firstRecord, err := store.Put(context.Background(), first)
	if err != nil {
		t.Fatalf("first Put() error = %v", err)
	}
	secondRecord, err := store.Put(context.Background(), second)
	if err != nil {
		t.Fatalf("second Put() error = %v", err)
	}
	if firstRecord.ID == secondRecord.ID {
		t.Fatal("processing versions produced the same record identifier")
	}

	firstLoaded, err := store.Get(context.Background(), firstRecord.Key)
	if err != nil {
		t.Fatalf("first Get() error = %v", err)
	}
	secondLoaded, err := store.Get(context.Background(), secondRecord.Key)
	if err != nil {
		t.Fatalf("second Get() error = %v", err)
	}
	if firstLoaded.ID != firstRecord.ID ||
		secondLoaded.ID != secondRecord.ID {
		t.Fatal("version-aware Get returned the wrong record")
	}

	latestSecond, err := store.GetLatest(
		context.Background(),
		"trajectory-one",
		flightfeatures.SchemaVersionV1,
		second.Provenance.ProcessingVersion,
	)
	if err != nil {
		t.Fatalf("GetLatest(v2) error = %v", err)
	}
	if latestSecond.ID != secondRecord.ID {
		t.Fatal("version-aware GetLatest returned the wrong record")
	}

	page, err := store.List(
		context.Background(),
		ListQuery{
			TrajectoryID:      "trajectory-one",
			SchemaVersion:     flightfeatures.SchemaVersionV1,
			ProcessingVersion: second.Provenance.ProcessingVersion,
		},
	)
	if err != nil {
		t.Fatalf("List(v2) error = %v", err)
	}
	if len(page.Records) != 1 ||
		page.Records[0].ID != secondRecord.ID {
		t.Fatalf("unexpected version-aware page: %#v", page)
	}
}

func TestMemoryStoreDefaultsBlankProcessingVersion(t *testing.T) {
	store := NewMemory(MemoryConfig{})
	features := validStoredFeatures(
		"trajectory-one",
		time.Date(2026, time.July, 25, 12, 0, 0, 0, time.UTC),
		"a",
	)
	features.Provenance.ProcessingVersion = ""

	record, err := store.Put(context.Background(), features)
	if err != nil {
		t.Fatalf("Put() error = %v", err)
	}
	if record.Key.ProcessingVersion !=
		flightfeatures.CurrentProcessingVersion {
		t.Fatalf(
			"processing version = %q, want %q",
			record.Key.ProcessingVersion,
			flightfeatures.CurrentProcessingVersion,
		)
	}
}

func TestLegacyProcessingSnapshotAcceptsOriginalIdentifier(t *testing.T) {
	asOf := time.Date(2026, time.July, 25, 12, 0, 0, 0, time.UTC)
	features := validStoredFeatures("trajectory-one", asOf, "a")
	features = normalizeFeatures(features)
	features.Provenance.ProcessingVersion =
		flightfeatures.LegacyProcessingVersion

	key := snapshotKey(features)
	record := Record{
		ID: makeLegacyRecordID(
			key,
			features.Provenance.InputFingerprint,
		),
		Key:              key,
		InputFingerprint: features.Provenance.InputFingerprint,
		Features:         features,
		StoredAt:         asOf.Add(time.Minute),
	}

	if err := validateDecodedRecord(
		record,
		flightfeatures.ValidationStatusValid,
	); err != nil {
		t.Fatalf("validateDecodedRecord() error = %v", err)
	}
}
