package featurestore

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

const recordIDPrefix = "feature-record-"

type MemoryStore struct {
	mutex            sync.RWMutex
	now              func() time.Time
	records          map[string]Record
	keysByTrajectory map[string][]string
}

func NewMemory(config MemoryConfig) *MemoryStore {
	now := config.Now
	if now == nil {
		now = time.Now
	}

	return &MemoryStore{
		now:              now,
		records:          make(map[string]Record),
		keysByTrajectory: make(map[string][]string),
	}
}

func (store *MemoryStore) Put(
	ctx context.Context,
	features flightfeatures.FlightFeatures,
) (Record, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return Record{}, err
	}
	if err := validateStorableFeatures(features); err != nil {
		return Record{}, err
	}

	normalized := normalizeFeatures(features)
	key := snapshotKey(normalized)
	compositeKey := encodeSnapshotKey(key)
	recordID := makeRecordID(
		compositeKey,
		normalized.Provenance.InputFingerprint,
	)

	store.mutex.Lock()
	defer store.mutex.Unlock()

	if existing, exists := store.records[compositeKey]; exists {
		if existing.InputFingerprint !=
			normalized.Provenance.InputFingerprint {
			return Record{}, ErrSnapshotConflict
		}

		return existing.Clone(), nil
	}

	record := Record{
		ID:               recordID,
		Key:              key,
		InputFingerprint: normalized.Provenance.InputFingerprint,
		Features:         normalized.Clone(),
		StoredAt:         store.now().UTC(),
	}

	store.records[compositeKey] = record.Clone()
	trajectoryIndexKey := encodeTrajectoryIndexKey(
		key.TrajectoryID,
		key.SchemaVersion,
	)
	store.keysByTrajectory[trajectoryIndexKey] = append(
		store.keysByTrajectory[trajectoryIndexKey],
		compositeKey,
	)
	store.sortTrajectoryIndexLocked(trajectoryIndexKey)

	return record.Clone(), nil
}

func (store *MemoryStore) Get(
	ctx context.Context,
	key SnapshotKey,
) (Record, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return Record{}, err
	}

	normalizedKey, err := normalizeSnapshotKey(key)
	if err != nil {
		return Record{}, err
	}
	compositeKey := encodeSnapshotKey(normalizedKey)

	store.mutex.RLock()
	record, exists := store.records[compositeKey]
	store.mutex.RUnlock()

	if !exists {
		return Record{}, ErrSnapshotNotFound
	}

	return record.Clone(), nil
}

func (store *MemoryStore) GetLatest(
	ctx context.Context,
	trajectoryID string,
	schemaVersion flightfeatures.SchemaVersion,
) (Record, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return Record{}, err
	}

	normalizedTrajectoryID, err := normalizeTrajectoryID(
		trajectoryID,
	)
	if err != nil {
		return Record{}, err
	}
	if schemaVersion != flightfeatures.SchemaVersionV1 {
		return Record{}, ErrUnsupportedSchemaVersion
	}

	indexKey := encodeTrajectoryIndexKey(
		normalizedTrajectoryID,
		schemaVersion,
	)

	store.mutex.RLock()
	keys := append(
		[]string(nil),
		store.keysByTrajectory[indexKey]...,
	)
	if len(keys) == 0 {
		store.mutex.RUnlock()
		return Record{}, ErrSnapshotNotFound
	}
	record := store.records[keys[0]]
	store.mutex.RUnlock()

	return record.Clone(), nil
}

func (store *MemoryStore) List(
	ctx context.Context,
	query ListQuery,
) (Page, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return Page{}, err
	}

	normalizedQuery, err := normalizeListQuery(query)
	if err != nil {
		return Page{}, err
	}
	indexKey := encodeTrajectoryIndexKey(
		normalizedQuery.TrajectoryID,
		normalizedQuery.SchemaVersion,
	)

	store.mutex.RLock()
	keys := append(
		[]string(nil),
		store.keysByTrajectory[indexKey]...,
	)
	result := make(
		[]Record,
		0,
		normalizedQuery.Limit+1,
	)
	for _, compositeKey := range keys {
		record := store.records[compositeKey]
		if !normalizedQuery.BeforeAsOfTime.IsZero() &&
			!record.Key.AsOfTime.Before(
				normalizedQuery.BeforeAsOfTime,
			) {
			continue
		}

		result = append(result, record.Clone())
		if len(result) == normalizedQuery.Limit+1 {
			break
		}
	}
	store.mutex.RUnlock()

	hasMore := len(result) > normalizedQuery.Limit
	if hasMore {
		result = result[:normalizedQuery.Limit]
	}

	return Page{
		Records: result,
		HasMore: hasMore,
	}.Clone(), nil
}

func (store *MemoryStore) sortTrajectoryIndexLocked(
	indexKey string,
) {
	keys := store.keysByTrajectory[indexKey]

	sort.SliceStable(keys, func(left int, right int) bool {
		leftRecord := store.records[keys[left]]
		rightRecord := store.records[keys[right]]

		if leftRecord.Key.AsOfTime.Equal(
			rightRecord.Key.AsOfTime,
		) {
			return leftRecord.ID < rightRecord.ID
		}

		return leftRecord.Key.AsOfTime.After(
			rightRecord.Key.AsOfTime,
		)
	})

	store.keysByTrajectory[indexKey] = keys
}

func validateStorableFeatures(
	features flightfeatures.FlightFeatures,
) error {
	if _, err := normalizeTrajectoryID(
		features.TrajectoryID,
	); err != nil {
		return err
	}
	if features.SchemaVersion !=
		flightfeatures.SchemaVersionV1 {
		return ErrUnsupportedSchemaVersion
	}
	if features.Window.AsOfTime.IsZero() {
		return ErrAsOfTimeRequired
	}
	if strings.TrimSpace(
		features.Provenance.InputFingerprint,
	) == "" {
		return ErrInputFingerprintRequired
	}

	switch features.Quality.Status {
	case flightfeatures.ValidationStatusValid,
		flightfeatures.ValidationStatusLimited:
		return nil
	case flightfeatures.ValidationStatusInvalid:
		return ErrFeaturesInvalid
	default:
		return ErrFeaturesUnvalidated
	}
}

func normalizeFeatures(
	features flightfeatures.FlightFeatures,
) flightfeatures.FlightFeatures {
	normalized := features.Clone()
	normalized.TrajectoryID =
		strings.TrimSpace(normalized.TrajectoryID)
	normalized.IdentityKey =
		strings.TrimSpace(normalized.IdentityKey)
	normalized.FlightID =
		strings.TrimSpace(normalized.FlightID)
	normalized.AircraftID =
		strings.TrimSpace(normalized.AircraftID)
	normalized.ICAO24 = strings.ToUpper(
		strings.TrimSpace(normalized.ICAO24),
	)
	normalized.Callsign =
		strings.TrimSpace(normalized.Callsign)
	normalized.Window.StartTime =
		normalized.Window.StartTime.UTC()
	normalized.Window.EndTime =
		normalized.Window.EndTime.UTC()
	normalized.Window.AsOfTime =
		normalized.Window.AsOfTime.UTC()
	normalized.ExtractedAt =
		normalized.ExtractedAt.UTC()
	normalized.Provenance.ExtractorVersion =
		strings.TrimSpace(
			normalized.Provenance.ExtractorVersion,
		)
	normalized.Provenance.InputFingerprint =
		strings.TrimSpace(
			normalized.Provenance.InputFingerprint,
		)
	normalized.Provenance.TrajectoryUpdatedAt =
		normalized.Provenance.TrajectoryUpdatedAt.UTC()

	return normalized
}

func snapshotKey(
	features flightfeatures.FlightFeatures,
) SnapshotKey {
	return SnapshotKey{
		TrajectoryID:  features.TrajectoryID,
		SchemaVersion: features.SchemaVersion,
		AsOfTime:      features.Window.AsOfTime,
	}
}

func normalizeSnapshotKey(
	key SnapshotKey,
) (SnapshotKey, error) {
	trajectoryID, err := normalizeTrajectoryID(
		key.TrajectoryID,
	)
	if err != nil {
		return SnapshotKey{}, err
	}
	if key.SchemaVersion !=
		flightfeatures.SchemaVersionV1 {
		return SnapshotKey{},
			ErrUnsupportedSchemaVersion
	}
	if key.AsOfTime.IsZero() {
		return SnapshotKey{}, ErrAsOfTimeRequired
	}

	return SnapshotKey{
		TrajectoryID:  trajectoryID,
		SchemaVersion: key.SchemaVersion,
		AsOfTime:      key.AsOfTime.UTC(),
	}, nil
}

func normalizeListQuery(
	query ListQuery,
) (ListQuery, error) {
	trajectoryID, err := normalizeTrajectoryID(
		query.TrajectoryID,
	)
	if err != nil {
		return ListQuery{}, err
	}
	if query.SchemaVersion !=
		flightfeatures.SchemaVersionV1 {
		return ListQuery{},
			ErrUnsupportedSchemaVersion
	}

	limit := query.Limit
	if limit == 0 {
		limit = DefaultListLimit
	}
	if limit < 1 || limit > MaximumListLimit {
		return ListQuery{}, ErrInvalidListLimit
	}

	beforeAsOfTime := query.BeforeAsOfTime
	if !beforeAsOfTime.IsZero() {
		beforeAsOfTime = beforeAsOfTime.UTC()
	}

	return ListQuery{
		TrajectoryID:   trajectoryID,
		SchemaVersion:  query.SchemaVersion,
		BeforeAsOfTime: beforeAsOfTime,
		Limit:          limit,
	}, nil
}

func normalizeTrajectoryID(
	trajectoryID string,
) (string, error) {
	normalized := strings.TrimSpace(trajectoryID)
	if normalized == "" {
		return "", ErrTrajectoryIDRequired
	}

	return normalized, nil
}

func encodeSnapshotKey(key SnapshotKey) string {
	return fmt.Sprintf(
		"%s\x00%s\x00%s",
		key.TrajectoryID,
		key.SchemaVersion,
		key.AsOfTime.UTC().Format(time.RFC3339Nano),
	)
}

func encodeTrajectoryIndexKey(
	trajectoryID string,
	schemaVersion flightfeatures.SchemaVersion,
) string {
	return fmt.Sprintf(
		"%s\x00%s",
		trajectoryID,
		schemaVersion,
	)
}

func makeRecordID(
	compositeKey string,
	fingerprint string,
) string {
	sum := sha256.Sum256(
		[]byte(compositeKey + "\x00" + fingerprint),
	)

	return recordIDPrefix + hex.EncodeToString(sum[:])
}
