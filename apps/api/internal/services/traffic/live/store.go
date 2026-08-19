package live

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	defaultTTL              = 60 * time.Second
	defaultCapacity         = 25000
	defaultSnapshotLimit    = 1500
	defaultMaxSnapshotLimit = 5000
	defaultMaxSelected      = 100
)

var ErrInvalidStoreConfig = errors.New("live traffic store configuration is invalid")

type Config struct {
	TTL                  time.Duration
	Capacity             int
	DefaultSnapshotLimit int
	MaxSnapshotLimit     int
	MaxSelected          int
	SourcePriority       map[string]int
}

func DefaultConfig() Config {
	return Config{
		TTL:                  defaultTTL,
		Capacity:             defaultCapacity,
		DefaultSnapshotLimit: defaultSnapshotLimit,
		MaxSnapshotLimit:     defaultMaxSnapshotLimit,
		MaxSelected:          defaultMaxSelected,
	}
}

type UpsertResult struct {
	Accepted int
	Ignored  int
	Rejected int
	Evicted  int
	Sequence uint64
}

type Store struct {
	mu       sync.RWMutex
	config   Config
	states   map[string]Aircraft
	sequence uint64
}

func NewStore(config Config) (*Store, error) {
	if config.TTL == 0 {
		config.TTL = defaultTTL
	}
	if config.Capacity == 0 {
		config.Capacity = defaultCapacity
	}
	if config.DefaultSnapshotLimit == 0 {
		config.DefaultSnapshotLimit = defaultSnapshotLimit
	}
	if config.MaxSnapshotLimit == 0 {
		config.MaxSnapshotLimit = defaultMaxSnapshotLimit
	}
	if config.MaxSelected == 0 {
		config.MaxSelected = defaultMaxSelected
	}
	if config.TTL <= 0 || config.Capacity <= 0 ||
		config.DefaultSnapshotLimit <= 0 || config.MaxSnapshotLimit <= 0 ||
		config.MaxSelected <= 0 ||
		config.DefaultSnapshotLimit > config.MaxSnapshotLimit ||
		config.MaxSelected > config.MaxSnapshotLimit {
		return nil, ErrInvalidStoreConfig
	}

	priorities := make(map[string]int, len(config.SourcePriority))
	for source, priority := range config.SourcePriority {
		normalized := strings.ToLower(strings.TrimSpace(source))
		if normalized == "" {
			return nil, fmt.Errorf("%w: empty source priority key", ErrInvalidStoreConfig)
		}
		priorities[normalized] = priority
	}
	config.SourcePriority = priorities

	return &Store{
		config: config,
		states: make(map[string]Aircraft, config.Capacity),
	}, nil
}

func (s *Store) UpsertBatch(candidates []Aircraft) UpsertResult {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := UpsertResult{}
	changed := false
	for _, candidate := range candidates {
		normalized, ok := normalizeAircraft(candidate)
		if !ok {
			result.Rejected++
			continue
		}

		current, exists := s.states[normalized.ICAO24]
		if exists && !s.shouldReplace(current, normalized) {
			result.Ignored++
			continue
		}

		s.states[normalized.ICAO24] = normalized
		result.Accepted++
		changed = true
	}

	if len(s.states) > s.config.Capacity {
		overflow := len(s.states) - s.config.Capacity
		keys := make([]string, 0, len(s.states))
		for key := range s.states {
			keys = append(keys, key)
		}
		sort.Slice(keys, func(i, j int) bool {
			left := s.states[keys[i]]
			right := s.states[keys[j]]
			if !left.ObservedAt.Equal(right.ObservedAt) {
				return left.ObservedAt.Before(right.ObservedAt)
			}
			if !left.ReceivedAt.Equal(right.ReceivedAt) {
				return left.ReceivedAt.Before(right.ReceivedAt)
			}
			return keys[i] < keys[j]
		})
		for _, key := range keys[:overflow] {
			delete(s.states, key)
			result.Evicted++
		}
		changed = true
	}

	if changed {
		s.sequence++
	}
	result.Sequence = s.sequence
	return result
}

func (s *Store) Snapshot(now time.Time, query SnapshotQuery) (Snapshot, error) {
	if now.IsZero() {
		return Snapshot{}, fmt.Errorf("snapshot server time is required")
	}
	if query.Bounds != nil {
		if err := query.Bounds.Validate(); err != nil {
			return Snapshot{}, err
		}
	}

	limit := query.Limit
	if limit == 0 {
		limit = s.config.DefaultSnapshotLimit
	}
	if limit <= 0 || limit > s.config.MaxSnapshotLimit {
		return Snapshot{}, ErrInvalidLimit
	}

	selected := make(map[string]struct{}, len(query.SelectedICAO24))
	for _, raw := range query.SelectedICAO24 {
		normalized, err := NormalizeICAO24(raw)
		if err != nil {
			return Snapshot{}, err
		}
		selected[normalized] = struct{}{}
	}
	if len(selected) > s.config.MaxSelected {
		return Snapshot{}, ErrTooManySelected
	}
	if len(selected) > limit {
		return Snapshot{}, fmt.Errorf("%w: limit must include every selected aircraft", ErrInvalidLimit)
	}

	s.mu.Lock()
	pruned := s.pruneLocked(now)
	if pruned > 0 {
		s.sequence++
	}

	type candidate struct {
		aircraft Aircraft
		selected bool
	}
	candidates := make([]candidate, 0, len(s.states))
	for key, state := range s.states {
		_, isSelected := selected[key]
		inBounds := query.Bounds == nil || query.Bounds.Contains(state.Latitude, state.Longitude)
		if !isSelected && !inBounds {
			continue
		}
		candidates = append(candidates, candidate{
			aircraft: cloneAircraft(state),
			selected: isSelected,
		})
	}
	totalActive := len(s.states)
	sequence := s.sequence
	s.mu.Unlock()

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].selected != candidates[j].selected {
			return candidates[i].selected
		}
		return candidates[i].aircraft.ICAO24 < candidates[j].aircraft.ICAO24
	})

	matching := len(candidates)
	truncated := matching > limit
	if truncated {
		candidates = candidates[:limit]
	}

	items := make([]Aircraft, 0, len(candidates))
	for _, item := range candidates {
		items = append(items, item.aircraft)
	}

	return Snapshot{
		ServerTime:  now.UTC(),
		Sequence:    sequence,
		Aircraft:    items,
		TotalActive: totalActive,
		Matching:    matching,
		Truncated:   truncated,
	}, nil
}

func (s *Store) shouldReplace(current, candidate Aircraft) bool {
	if !candidate.ObservedAt.Equal(current.ObservedAt) {
		return candidate.ObservedAt.After(current.ObservedAt)
	}

	candidatePriority := s.sourcePriority(candidate.Source)
	currentPriority := s.sourcePriority(current.Source)
	if candidatePriority != currentPriority {
		return candidatePriority > currentPriority
	}
	if !candidate.ReceivedAt.Equal(current.ReceivedAt) {
		return candidate.ReceivedAt.After(current.ReceivedAt)
	}
	return aircraftTieKey(candidate) > aircraftTieKey(current)
}

func (s *Store) sourcePriority(source string) int {
	return s.config.SourcePriority[strings.ToLower(strings.TrimSpace(source))]
}

func (s *Store) pruneLocked(now time.Time) int {
	pruned := 0
	for key, state := range s.states {
		if now.After(state.ObservedAt) && now.Sub(state.ObservedAt) > s.config.TTL {
			delete(s.states, key)
			pruned++
		}
	}
	return pruned
}

func normalizeAircraft(candidate Aircraft) (Aircraft, bool) {
	icao24, err := NormalizeICAO24(candidate.ICAO24)
	if err != nil || !finite(candidate.Latitude) || !finite(candidate.Longitude) ||
		candidate.Latitude < -90 || candidate.Latitude > 90 ||
		candidate.Longitude < -180 || candidate.Longitude > 180 ||
		candidate.ObservedAt.IsZero() || candidate.ReceivedAt.IsZero() ||
		strings.TrimSpace(candidate.Source) == "" ||
		!optionalFinite(candidate.AltitudeM) ||
		!optionalFinite(candidate.VelocityMPS) ||
		!optionalFinite(candidate.HeadingDegrees) ||
		!optionalFinite(candidate.VerticalRateMPS) {
		return Aircraft{}, false
	}

	candidate.ICAO24 = icao24
	candidate.Callsign = strings.TrimSpace(candidate.Callsign)
	candidate.Source = strings.ToLower(strings.TrimSpace(candidate.Source))
	candidate.ObservedAt = candidate.ObservedAt.UTC()
	candidate.ReceivedAt = candidate.ReceivedAt.UTC()
	return cloneAircraft(candidate), true
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func optionalFinite(value *float64) bool {
	return value == nil || finite(*value)
}

func cloneAircraft(value Aircraft) Aircraft {
	value.AltitudeM = cloneFloat(value.AltitudeM)
	value.VelocityMPS = cloneFloat(value.VelocityMPS)
	value.HeadingDegrees = cloneFloat(value.HeadingDegrees)
	value.VerticalRateMPS = cloneFloat(value.VerticalRateMPS)
	value.OnGround = cloneBool(value.OnGround)
	return value
}

func cloneFloat(value *float64) *float64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func cloneBool(value *bool) *bool {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func aircraftTieKey(value Aircraft) string {
	return fmt.Sprintf(
		"%s|%s|%.8f|%.8f|%s|%s|%s|%s|%s",
		value.Source,
		value.Callsign,
		value.Latitude,
		value.Longitude,
		optionalFloatKey(value.AltitudeM),
		optionalFloatKey(value.VelocityMPS),
		optionalFloatKey(value.HeadingDegrees),
		optionalFloatKey(value.VerticalRateMPS),
		optionalBoolKey(value.OnGround),
	)
}

func optionalFloatKey(value *float64) string {
	if value == nil {
		return "null"
	}
	return fmt.Sprintf("%.9f", *value)
}

func optionalBoolKey(value *bool) string {
	if value == nil {
		return "null"
	}
	return fmt.Sprintf("%t", *value)
}
