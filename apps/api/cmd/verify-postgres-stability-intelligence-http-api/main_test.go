package main

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/stabilityintelligence/stabilityproduction"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestConnectPostgreSQLWithRetryUsingRecoversAfterTransientFailures(t *testing.T) {
	attempts := 0
	var timeouts []time.Duration
	var delays []time.Duration
	var log bytes.Buffer

	pool, err := connectPostgreSQLWithRetryUsing(
		"postgres://verification",
		time.Second,
		&log,
		func(_ string, timeout time.Duration) (*pgxpool.Pool, error) {
			attempts++
			timeouts = append(timeouts, timeout)
			if attempts < 3 {
				return nil, errors.New("temporary connection failure")
			}
			return &pgxpool.Pool{}, nil
		},
		func(delay time.Duration) {
			delays = append(delays, delay)
		},
	)
	if err != nil {
		t.Fatalf("connect with retry: %v", err)
	}
	if pool == nil {
		t.Fatal("connection pool is nil")
	}
	if attempts != 3 {
		t.Fatalf("attempts = %d, want 3", attempts)
	}
	for index, timeout := range timeouts {
		if timeout != verificationMinimumDatabaseConnectTimeout {
			t.Fatalf("timeout[%d] = %s, want %s", index, timeout, verificationMinimumDatabaseConnectTimeout)
		}
	}
	wantDelays := []time.Duration{2 * time.Second, 4 * time.Second}
	if len(delays) != len(wantDelays) {
		t.Fatalf("delays = %#v, want %#v", delays, wantDelays)
	}
	for index := range wantDelays {
		if delays[index] != wantDelays[index] {
			t.Fatalf("delay[%d] = %s, want %s", index, delays[index], wantDelays[index])
		}
	}
	if !strings.Contains(log.String(), "established on attempt 3 of 4") {
		t.Fatalf("connection log = %q", log.String())
	}
}

func TestConnectPostgreSQLWithRetryUsingReturnsLastFailure(t *testing.T) {
	attempts := 0
	var delays []time.Duration
	lastFailure := errors.New("database remains unavailable")

	pool, err := connectPostgreSQLWithRetryUsing(
		"postgres://verification",
		45*time.Second,
		&bytes.Buffer{},
		func(_ string, timeout time.Duration) (*pgxpool.Pool, error) {
			attempts++
			if timeout != 45*time.Second {
				t.Fatalf("timeout = %s, want 45s", timeout)
			}
			return nil, lastFailure
		},
		func(delay time.Duration) {
			delays = append(delays, delay)
		},
	)
	if pool != nil {
		t.Fatal("unexpected connection pool")
	}
	if err == nil || !errors.Is(err, lastFailure) {
		t.Fatalf("error = %v, want wrapped last failure", err)
	}
	if attempts != verificationDatabaseConnectAttempts {
		t.Fatalf("attempts = %d, want %d", attempts, verificationDatabaseConnectAttempts)
	}
	wantDelays := []time.Duration{2 * time.Second, 4 * time.Second, 6 * time.Second}
	if len(delays) != len(wantDelays) {
		t.Fatalf("delays = %#v, want %#v", delays, wantDelays)
	}
	for index := range wantDelays {
		if delays[index] != wantDelays[index] {
			t.Fatalf("delay[%d] = %s, want %s", index, delays[index], wantDelays[index])
		}
	}
}

func TestRuntimeVerificationTimeoutUsesMinimum(t *testing.T) {
	if got := runtimeVerificationTimeout(time.Second); got != verificationMinimumRuntimeTimeout {
		t.Fatalf("minimum runtime timeout = %s, want %s", got, verificationMinimumRuntimeTimeout)
	}
	configured := 10 * time.Minute
	if got := runtimeVerificationTimeout(configured); got != configured {
		t.Fatalf("configured runtime timeout = %s, want %s", got, configured)
	}
}

func TestVerificationIdentityKeyMatchesFlightTrajectoryContract(t *testing.T) {
	pattern := regexp.MustCompile(`^flight-identity-[0-9a-f]{64}$`)
	if !pattern.MatchString(verificationIdentityKey) {
		t.Fatalf(
			"verification identity key %q does not match canonical flight identity format",
			verificationIdentityKey,
		)
	}
}

func TestInsertFixtureWithExecutorCoversCompleteFixture(t *testing.T) {
	schedule, err := buildVerificationSchedule(
		time.Date(2026, time.July, 17, 12, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("build verification schedule: %v", err)
	}

	callCount := 0
	trajectoryIdentity := ""
	err = insertFixtureWithExecutor(
		context.Background(),
		schedule,
		func(
			_ context.Context,
			_ string,
			arguments ...any,
		) error {
			callCount++
			if callCount == 1 {
				if len(arguments) < 2 {
					t.Fatalf("trajectory insert arguments = %d, want at least 2", len(arguments))
				}
				value, ok := arguments[1].(string)
				if !ok {
					t.Fatalf("trajectory identity argument type = %T, want string", arguments[1])
				}
				trajectoryIdentity = value
			}
			return nil
		},
	)
	if err != nil {
		t.Fatalf("insert fixture through test executor: %v", err)
	}
	if callCount != verificationStateCount+1 {
		t.Fatalf(
			"fixture insert calls = %d, want %d",
			callCount,
			verificationStateCount+1,
		)
	}
	if trajectoryIdentity != verificationIdentityKey {
		t.Fatalf(
			"trajectory identity = %q, want %q",
			trajectoryIdentity,
			verificationIdentityKey,
		)
	}
}

type stubStabilityIntelligenceService struct {
	get func(
		context.Context,
		stabilityproduction.Request,
	) (stabilityproduction.Result, error)
}

func (
	service stubStabilityIntelligenceService,
) Get(
	ctx context.Context,
	request stabilityproduction.Request,
) (stabilityproduction.Result, error) {
	if service.get == nil {
		return stabilityproduction.Result{}, nil
	}
	return service.get(ctx, request)
}

func TestRuntimeStabilityReaderAppliesBoundedContext(t *testing.T) {
	reader := runtimeStabilityReader{
		service: stubStabilityIntelligenceService{
			get: func(
				ctx context.Context,
				_ stabilityproduction.Request,
			) (stabilityproduction.Result, error) {
				deadline, ok := ctx.Deadline()
				if !ok {
					t.Fatal("runtime Stability Intelligence context has no deadline")
				}
				remaining := time.Until(deadline)
				if remaining <= 0 || remaining > 100*time.Millisecond {
					t.Fatalf("runtime deadline remaining = %s, want within 100ms", remaining)
				}
				<-ctx.Done()
				return stabilityproduction.Result{}, ctx.Err()
			},
		},
		timeout: 25 * time.Millisecond,
	}

	startedAt := time.Now()
	_, err := reader.Get(
		context.Background(),
		stabilityproduction.Request{},
	)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("runtime reader error = %v, want context deadline exceeded", err)
	}
	if elapsed := time.Since(startedAt); elapsed > time.Second {
		t.Fatalf("runtime reader elapsed = %s, want bounded execution", elapsed)
	}
}

func TestRuntimeStabilityReaderUsesDefaultHTTPServiceTimeout(t *testing.T) {
	reader := runtimeStabilityReader{
		service: stubStabilityIntelligenceService{
			get: func(
				ctx context.Context,
				_ stabilityproduction.Request,
			) (stabilityproduction.Result, error) {
				deadline, ok := ctx.Deadline()
				if !ok {
					t.Fatal("runtime Stability Intelligence context has no deadline")
				}
				remaining := time.Until(deadline)
				if remaining < verificationHTTPServiceTimeout-time.Second ||
					remaining > verificationHTTPServiceTimeout {
					t.Fatalf(
						"default runtime timeout remaining = %s, want approximately %s",
						remaining,
						verificationHTTPServiceTimeout,
					)
				}
				return stabilityproduction.Result{}, nil
			},
		},
	}

	if _, err := reader.Get(
		context.Background(),
		stabilityproduction.Request{},
	); err != nil {
		t.Fatalf("runtime reader with default timeout: %v", err)
	}
}

func TestExecuteFiberRequestAllowsDatabaseBackedRequestBeyondOneSecond(t *testing.T) {
	app := fiber.New()
	app.Get(
		"/slow-runtime-verification",
		func(ctx *fiber.Ctx) error {
			time.Sleep(1100 * time.Millisecond)
			return ctx.SendStatus(fiber.StatusNoContent)
		},
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/slow-runtime-verification",
		nil,
	)
	response, err := executeFiberRequest(app, request)
	if err != nil {
		t.Fatalf("execute bounded Fiber request: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != fiber.StatusNoContent {
		t.Fatalf(
			"slow Fiber response status = %d, want %d",
			response.StatusCode,
			fiber.StatusNoContent,
		)
	}
}

func TestExecuteFiberRequestRejectsMissingInputs(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	if _, err := executeFiberRequest(nil, request); err == nil {
		t.Fatal("missing Fiber application was accepted")
	}
	if _, err := executeFiberRequest(fiber.New(), nil); err == nil {
		t.Fatal("missing HTTP request was accepted")
	}
}
