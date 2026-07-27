package featurestore

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/validator"
	"github.com/jackc/pgx/v5"
)

const reviewTrajectoryID = "11111111-1111-4111-8111-111111111111"

func TestMemoryStoreRejectsSameInputWithDifferentOutput(t *testing.T) {
	store := NewMemory(MemoryConfig{})
	asOfTime := time.Date(2026, time.July, 27, 10, 0, 0, 0, time.UTC)
	first := validStoredFeatures(reviewTrajectoryID, asOfTime, "a")
	second := first.Clone()
	second.Aircraft.Model = "A321"

	if _, err := store.Put(context.Background(), first); err != nil {
		t.Fatalf("first Put() error = %v", err)
	}
	if _, err := store.Put(context.Background(), second); !errors.Is(err, ErrSnapshotConflict) {
		t.Fatalf("second Put() error = %v, want %v", err, ErrSnapshotConflict)
	}
}

func TestPostgresStoreRejectsSameInputWithDifferentOutput(t *testing.T) {
	asOfTime := time.Date(2026, time.July, 27, 10, 0, 0, 0, time.UTC)
	existingFeatures := validPostgresFeatures(testTrajectoryID, asOfTime, "a")
	incoming := existingFeatures.Clone()
	incoming.Aircraft.Model = "A321"
	existing := expectedRecord(existingFeatures, asOfTime.Add(time.Minute))
	call := 0
	client := &fakePostgresClient{
		queryRow: func(context.Context, string, ...any) rowScanner {
			call++
			if call == 1 {
				return errorRow{err: pgx.ErrNoRows}
			}
			return rowFromRecord(t, existing)
		},
	}

	store := newPostgresStore(client, time.Now)
	if _, err := store.Put(context.Background(), incoming); !errors.Is(err, ErrSnapshotConflict) {
		t.Fatalf("Put() error = %v, want %v", err, ErrSnapshotConflict)
	}
}

func TestOutputFingerprintIgnoresOperationalTimestamps(t *testing.T) {
	asOfTime := time.Date(2026, time.July, 27, 10, 0, 0, 0, time.UTC)
	first := validStoredFeatures(reviewTrajectoryID, asOfTime, "a")
	second := first.Clone()
	second.ExtractedAt = first.ExtractedAt.Add(time.Hour)
	second.ValidationReport.ValidatedAt = first.ValidationReport.ValidatedAt.Add(time.Hour)
	second.Provenance.AircraftMetadataRetrievedAt = first.Provenance.AircraftMetadataRetrievedAt.Add(time.Hour)

	firstFingerprint, err := fingerprintSnapshotOutput(first)
	if err != nil {
		t.Fatalf("first fingerprint error = %v", err)
	}
	secondFingerprint, err := fingerprintSnapshotOutput(second)
	if err != nil {
		t.Fatalf("second fingerprint error = %v", err)
	}
	if firstFingerprint != secondFingerprint {
		t.Fatalf("operational timestamps changed output identity: %q != %q", firstFingerprint, secondFingerprint)
	}
}

func TestVersionedPayloadRoundTripAndStrictDecoding(t *testing.T) {
	features := validStoredFeatures(
		reviewTrajectoryID,
		time.Date(2026, time.July, 27, 10, 0, 0, 0, time.UTC),
		"a",
	)
	encoded, fingerprint, err := encodeSnapshotPayload(features)
	if err != nil {
		t.Fatalf("encodeSnapshotPayload() error = %v", err)
	}
	if !strings.Contains(string(encoded), `"PayloadVersion":"`+snapshotPayloadVersionV1+`"`) ||
		!strings.Contains(string(encoded), `"OutputFingerprint":"`+fingerprint+`"`) {
		t.Fatalf("versioned payload markers are missing: %s", encoded)
	}

	decoded, decodedFingerprint, err := decodeSnapshotPayload(encoded)
	if err != nil {
		t.Fatalf("decodeSnapshotPayload() error = %v", err)
	}
	if decodedFingerprint != fingerprint || !reflect.DeepEqual(decoded, features) {
		t.Fatalf("round trip mismatch\nfeatures=%#v\ndecoded=%#v", features, decoded)
	}

	var object map[string]any
	if err := json.Unmarshal(encoded, &object); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	object["UnexpectedField"] = true
	malformed, err := json.Marshal(object)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if _, _, err := decodeSnapshotPayload(malformed); err == nil {
		t.Fatal("strict decoder accepted an unknown persistence field")
	}

	delete(object, "UnexpectedField")
	delete(object, "Aircraft")
	missingField, err := json.Marshal(object)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if _, _, err := decodeSnapshotPayload(missingField); !errors.Is(err, ErrCorruptSnapshot) {
		t.Fatalf("missing field error = %v, want %v", err, ErrCorruptSnapshot)
	}
}

func TestLegacyPayloadRemainsReadable(t *testing.T) {
	features := validStoredFeatures(
		reviewTrajectoryID,
		time.Date(2026, time.July, 27, 10, 0, 0, 0, time.UTC),
		"a",
	)
	legacy, err := json.Marshal(features)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	decoded, fingerprint, err := decodeSnapshotPayload(legacy)
	if err != nil {
		t.Fatalf("decodeSnapshotPayload() error = %v", err)
	}
	if fingerprint == "" || !reflect.DeepEqual(decoded, features) {
		t.Fatalf("legacy round trip mismatch: %#v", decoded)
	}
}

func TestStoresUseProductionIdentityContracts(t *testing.T) {
	stores := map[string]SnapshotWriter{
		"memory": NewMemory(MemoryConfig{}),
		"postgres": newPostgresStore(
			&fakePostgresClient{},
			time.Now,
		),
	}
	for name, store := range stores {
		t.Run(name, func(t *testing.T) {
			features := validStoredFeatures(
				reviewTrajectoryID,
				time.Date(2026, time.July, 27, 10, 0, 0, 0, time.UTC),
				"a",
			)
			features.TrajectoryID = "trajectory-not-a-uuid"
			if _, err := store.Put(context.Background(), features); !errors.Is(err, ErrInvalidTrajectoryID) {
				t.Fatalf("invalid trajectory Put() error = %v, want %v", err, ErrInvalidTrajectoryID)
			}

			features = validStoredFeatures(
				reviewTrajectoryID,
				time.Date(2026, time.July, 27, 10, 0, 0, 0, time.UTC),
				"a",
			)
			features.Provenance.InputFingerprint = "fingerprint-a"
			if _, err := store.Put(context.Background(), features); !errors.Is(err, ErrInvalidInputFingerprint) {
				t.Fatalf("invalid fingerprint Put() error = %v, want %v", err, ErrInvalidInputFingerprint)
			}
		})
	}
}

func TestStoresRequireCompleteValidationProof(t *testing.T) {
	stores := map[string]SnapshotWriter{
		"memory": NewMemory(MemoryConfig{}),
		"postgres": newPostgresStore(
			&fakePostgresClient{},
			time.Now,
		),
	}
	for name, store := range stores {
		t.Run(name, func(t *testing.T) {
			features := validStoredFeatures(
				reviewTrajectoryID,
				time.Date(2026, time.July, 27, 10, 0, 0, 0, time.UTC),
				"a",
			)
			features.ValidationReport = flightfeatures.ValidationReport{}
			if _, err := store.Put(context.Background(), features); !errors.Is(err, ErrValidationProofRequired) {
				t.Fatalf("Put() error = %v, want %v", err, ErrValidationProofRequired)
			}
		})
	}
}

func TestStoresRejectNilContextAndNonFiniteValues(t *testing.T) {
	stores := map[string]SnapshotWriter{
		"memory": NewMemory(MemoryConfig{}),
		"postgres": newPostgresStore(
			&fakePostgresClient{},
			time.Now,
		),
	}
	for name, store := range stores {
		t.Run(name, func(t *testing.T) {
			features := validStoredFeatures(
				reviewTrajectoryID,
				time.Date(2026, time.July, 27, 10, 0, 0, 0, time.UTC),
				"a",
			)
			if _, err := store.Put(nil, features); !errors.Is(err, ErrContextRequired) {
				t.Fatalf("nil context Put() error = %v, want %v", err, ErrContextRequired)
			}

			features.Geographical.GreatCircleDistanceKM = math.NaN()
			if _, err := store.Put(context.Background(), features); !errors.Is(err, ErrNonFiniteFeatureValue) {
				t.Fatalf("non-finite Put() error = %v, want %v", err, ErrNonFiniteFeatureValue)
			}
		})
	}
}

func TestMemoryStoreEnforcesCapacityWithoutEviction(t *testing.T) {
	store := NewMemory(MemoryConfig{MaximumRecords: 1})
	first := validStoredFeatures(
		reviewTrajectoryID,
		time.Date(2026, time.July, 27, 10, 0, 0, 0, time.UTC),
		"a",
	)
	second := validStoredFeatures(
		"22222222-2222-4222-8222-222222222222",
		time.Date(2026, time.July, 27, 11, 0, 0, 0, time.UTC),
		"b",
	)
	if _, err := store.Put(context.Background(), first); err != nil {
		t.Fatalf("first Put() error = %v", err)
	}
	if _, err := store.Put(context.Background(), second); !errors.Is(err, ErrMemoryCapacityExceeded) {
		t.Fatalf("second Put() error = %v, want %v", err, ErrMemoryCapacityExceeded)
	}
}

func TestStoreInterfacesRemainSegregated(t *testing.T) {
	var writer SnapshotWriter = NewMemory(MemoryConfig{})
	var reader SnapshotReader = NewMemory(MemoryConfig{})
	if writer == nil || reader == nil {
		t.Fatal("segregated store interfaces are not implemented")
	}
}

func testCompleteValidationReport(
	status flightfeatures.ValidationStatus,
	validatedAt time.Time,
) flightfeatures.ValidationReport {
	return flightfeatures.ValidationReport{
		AuditState:       flightfeatures.ValidationAuditStateComplete,
		ValidatorVersion: validator.Version,
		Status:           status,
		ValidatedAt:      validatedAt.UTC(),
	}
}
