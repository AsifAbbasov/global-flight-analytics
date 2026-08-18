package livecollector

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	aviationconstraints "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/constraints"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	livetraffic "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/live"
)

const (
	defaultJitterRatio = 0.10
	maxJitterRatio     = 0.50
	maxPointRadiusNM   = 250
)

var (
	ErrContextRequired = errors.New("live collector context is required")
	ErrLoaderRequired  = errors.New("live collector provider loader is required")
	ErrStoreRequired   = errors.New("live collector store is required")
	ErrTargetsRequired = errors.New("live collector requires at least one target")
	ErrAlreadyRunning  = errors.New("live collector is already running")
	ErrInvalidConfig   = errors.New("live collector configuration is invalid")
)

type Loader interface {
	SourceName() string

	LoadByPoint(
		ctx context.Context,
		latitude float64,
		longitude float64,
		radius int,
	) ([]flightstate.FlightState, error)
}

type Store interface {
	UpsertFlightStates(
		states []flightstate.FlightState,
		receivedAt time.Time,
	) livetraffic.UpsertResult
}

type Target struct {
	Name      string
	Latitude  float64
	Longitude float64
	RadiusNM  int
}

func (target Target) Validate() error {
	if strings.TrimSpace(target.Name) == "" {
		return fmt.Errorf("%w: target name is required", ErrInvalidConfig)
	}
	if !aviationconstraints.IsLatitude(target.Latitude) {
		return fmt.Errorf("%w: target latitude is invalid", ErrInvalidConfig)
	}
	if !aviationconstraints.IsLongitude(target.Longitude) {
		return fmt.Errorf("%w: target longitude is invalid", ErrInvalidConfig)
	}
	if target.RadiusNM <= 0 || target.RadiusNM > maxPointRadiusNM {
		return fmt.Errorf(
			"%w: target radius must be between 1 and %d nautical miles",
			ErrInvalidConfig,
			maxPointRadiusNM,
		)
	}
	return nil
}

type Config struct {
	Loader Loader
	Store  Store

	Targets []Target

	PollInterval   time.Duration
	RequestTimeout time.Duration
	RequestSpacing time.Duration
	MaxBackoff     time.Duration
	JitterRatio    float64

	Logger *slog.Logger

	Now           func() time.Time
	RandomFloat64 func() float64
}

type Status struct {
	Running bool
	Source  string

	StartedAt     time.Time
	LastCycleAt   time.Time
	LastSuccessAt time.Time
	LastFailureAt time.Time

	Cycles           uint64
	SuccessfulCycles uint64
	FailedCycles     uint64

	StatesLoaded uint64
	Accepted     uint64
	Ignored      uint64
	Rejected     uint64
	Evicted      uint64

	ConsecutiveFailures int
	LastError           string
}

type CycleResult struct {
	Source string

	TargetsAttempted int
	TargetsSucceeded int
	StatesLoaded     int

	Upsert livetraffic.UpsertResult

	CompletedAt time.Time
}

type Collector struct {
	loader Loader
	store  Store

	targets []Target

	pollInterval   time.Duration
	requestTimeout time.Duration
	requestSpacing time.Duration
	maxBackoff     time.Duration
	jitterRatio    float64

	logger *slog.Logger

	now           func() time.Time
	randomFloat64 func() float64

	running atomic.Bool

	statusMu sync.RWMutex
	status   Status
}

func New(config Config) (*Collector, error) {
	if config.Loader == nil {
		return nil, ErrLoaderRequired
	}
	if config.Store == nil {
		return nil, ErrStoreRequired
	}
	if len(config.Targets) == 0 {
		return nil, ErrTargetsRequired
	}
	if config.PollInterval <= 0 {
		return nil, fmt.Errorf("%w: poll interval must be greater than zero", ErrInvalidConfig)
	}
	if config.RequestTimeout <= 0 {
		return nil, fmt.Errorf("%w: request timeout must be greater than zero", ErrInvalidConfig)
	}
	if config.RequestSpacing < 0 {
		return nil, fmt.Errorf("%w: request spacing must not be negative", ErrInvalidConfig)
	}
	if config.MaxBackoff <= 0 || config.MaxBackoff < config.PollInterval {
		return nil, fmt.Errorf(
			"%w: max backoff must be at least the poll interval",
			ErrInvalidConfig,
		)
	}

	jitterRatio := config.JitterRatio
	if jitterRatio == 0 {
		jitterRatio = defaultJitterRatio
	}
	if jitterRatio < 0 || jitterRatio > maxJitterRatio {
		return nil, fmt.Errorf(
			"%w: jitter ratio must be between zero and %.2f",
			ErrInvalidConfig,
			maxJitterRatio,
		)
	}

	targets := make([]Target, 0, len(config.Targets))
	seenNames := make(map[string]struct{}, len(config.Targets))
	for _, target := range config.Targets {
		if err := target.Validate(); err != nil {
			return nil, err
		}
		target.Name = strings.TrimSpace(target.Name)
		key := strings.ToLower(target.Name)
		if _, exists := seenNames[key]; exists {
			return nil, fmt.Errorf(
				"%w: duplicate target name %q",
				ErrInvalidConfig,
				target.Name,
			)
		}
		seenNames[key] = struct{}{}
		targets = append(targets, target)
	}

	now := config.Now
	if now == nil {
		now = time.Now
	}
	randomFloat64 := config.RandomFloat64
	if randomFloat64 == nil {
		randomFloat64 = rand.Float64
	}
	logger := config.Logger
	if logger == nil {
		logger = slog.Default()
	}

	return &Collector{
		loader:         config.Loader,
		store:          config.Store,
		targets:        targets,
		pollInterval:   config.PollInterval,
		requestTimeout: config.RequestTimeout,
		requestSpacing: config.RequestSpacing,
		maxBackoff:     config.MaxBackoff,
		jitterRatio:    jitterRatio,
		logger:         logger,
		now:            now,
		randomFloat64:  randomFloat64,
		status: Status{
			Source: strings.TrimSpace(config.Loader.SourceName()),
		},
	}, nil
}

func (collector *Collector) Run(ctx context.Context) error {
	if ctx == nil {
		return ErrContextRequired
	}
	if !collector.running.CompareAndSwap(false, true) {
		return ErrAlreadyRunning
	}
	defer collector.running.Store(false)

	collector.markStarted()
	consecutiveFailures := 0

	for {
		_, err := collector.CollectOnce(ctx)
		if ctx.Err() != nil {
			collector.markStopped()
			return nil
		}

		var delay time.Duration
		if err != nil {
			consecutiveFailures++
			delay = collector.failureDelay(err, consecutiveFailures)
		} else {
			consecutiveFailures = 0
			delay = collector.successDelay()
		}

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			collector.markStopped()
			return nil
		case <-timer.C:
		}
	}
}

func (collector *Collector) CollectOnce(ctx context.Context) (CycleResult, error) {
	if ctx == nil {
		return CycleResult{}, ErrContextRequired
	}

	result := CycleResult{
		Source: strings.TrimSpace(collector.loader.SourceName()),
	}

	for index, target := range collector.targets {
		if index > 0 && collector.requestSpacing > 0 {
			timer := time.NewTimer(collector.requestSpacing)
			select {
			case <-ctx.Done():
				if !timer.Stop() {
					<-timer.C
				}
				return collector.finishFailure(result, ctx.Err())
			case <-timer.C:
			}
		}

		result.TargetsAttempted++

		requestContext, cancel := context.WithTimeout(ctx, collector.requestTimeout)
		states, err := collector.loader.LoadByPoint(
			requestContext,
			target.Latitude,
			target.Longitude,
			target.RadiusNM,
		)
		cancel()
		if err != nil {
			return collector.finishFailure(
				result,
				fmt.Errorf("collect target %s: %w", target.Name, err),
			)
		}

		receivedAt := collector.now().UTC()
		upsert := collector.store.UpsertFlightStates(states, receivedAt)

		result.TargetsSucceeded++
		result.StatesLoaded += len(states)
		result.Upsert.Accepted += upsert.Accepted
		result.Upsert.Ignored += upsert.Ignored
		result.Upsert.Rejected += upsert.Rejected
		result.Upsert.Evicted += upsert.Evicted
		result.Upsert.Sequence = upsert.Sequence
	}

	result.CompletedAt = collector.now().UTC()
	collector.finishSuccess(result)
	return result, nil
}

func (collector *Collector) SnapshotStatus() Status {
	collector.statusMu.RLock()
	defer collector.statusMu.RUnlock()
	return collector.status
}

func (collector *Collector) successDelay() time.Duration {
	return collector.addPositiveJitter(collector.pollInterval)
}

func (collector *Collector) failureDelay(err error, consecutiveFailures int) time.Duration {
	delay := collector.pollInterval
	for attempt := 1; attempt < consecutiveFailures; attempt++ {
		if delay >= collector.maxBackoff/2 {
			delay = collector.maxBackoff
			break
		}
		delay *= 2
	}
	if delay > collector.maxBackoff {
		delay = collector.maxBackoff
	}

	delay = collector.addPositiveJitter(delay)

	retryAt := retryAtFromError(err)
	if !retryAt.IsZero() {
		untilRetry := retryAt.Sub(collector.now().UTC())
		if untilRetry > delay {
			delay = untilRetry
		}
	}

	return delay
}

func (collector *Collector) addPositiveJitter(delay time.Duration) time.Duration {
	if collector.jitterRatio <= 0 || delay <= 0 {
		return delay
	}
	random := collector.randomFloat64()
	if random < 0 {
		random = 0
	}
	if random > 1 {
		random = 1
	}
	extra := time.Duration(float64(delay) * collector.jitterRatio * random)
	return delay + extra
}

func (collector *Collector) markStarted() {
	now := collector.now().UTC()
	collector.statusMu.Lock()
	defer collector.statusMu.Unlock()
	collector.status.Running = true
	if collector.status.StartedAt.IsZero() {
		collector.status.StartedAt = now
	}
}

func (collector *Collector) markStopped() {
	collector.statusMu.Lock()
	defer collector.statusMu.Unlock()
	collector.status.Running = false
}

func (collector *Collector) finishSuccess(result CycleResult) {
	collector.statusMu.Lock()
	defer collector.statusMu.Unlock()

	collector.status.LastCycleAt = result.CompletedAt
	collector.status.LastSuccessAt = result.CompletedAt
	collector.status.Cycles++
	collector.status.SuccessfulCycles++
	collector.status.StatesLoaded += uint64(result.StatesLoaded)
	collector.status.Accepted += uint64(result.Upsert.Accepted)
	collector.status.Ignored += uint64(result.Upsert.Ignored)
	collector.status.Rejected += uint64(result.Upsert.Rejected)
	collector.status.Evicted += uint64(result.Upsert.Evicted)
	collector.status.ConsecutiveFailures = 0
	collector.status.LastError = ""
}

func (collector *Collector) finishFailure(result CycleResult, err error) (CycleResult, error) {
	result.CompletedAt = collector.now().UTC()

	collector.statusMu.Lock()
	collector.status.LastCycleAt = result.CompletedAt
	collector.status.LastFailureAt = result.CompletedAt
	collector.status.Cycles++
	collector.status.FailedCycles++
	collector.status.StatesLoaded += uint64(result.StatesLoaded)
	collector.status.Accepted += uint64(result.Upsert.Accepted)
	collector.status.Ignored += uint64(result.Upsert.Ignored)
	collector.status.Rejected += uint64(result.Upsert.Rejected)
	collector.status.Evicted += uint64(result.Upsert.Evicted)
	collector.status.ConsecutiveFailures++
	collector.status.LastError = err.Error()
	collector.statusMu.Unlock()

	collector.logger.Warn(
		"live traffic collector cycle failed",
		"source",
		result.Source,
		"targets_attempted",
		result.TargetsAttempted,
		"targets_succeeded",
		result.TargetsSucceeded,
		"error_type",
		fmt.Sprintf("%T", err),
	)

	return result, err
}

func retryAtFromError(err error) time.Time {
	var evidence interface {
		RetryAtTime() time.Time
	}
	if errors.As(err, &evidence) {
		return evidence.RetryAtTime().UTC()
	}
	return time.Time{}
}
