package providerbudget

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

func TestPublicationLifecycleReleasesFailedReservationForRetry(t *testing.T) {
	manager, err := New(nil)
	if err != nil {
		t.Fatalf("create provider budget manager: %v", err)
	}

	reservation, err := manager.ReservePublication(
		context.Background(),
		providerpolicy.ProviderOurAirports,
		"publication-a",
	)
	if err != nil {
		t.Fatalf("reserve publication: %v", err)
	}
	if !reservation.Decision.Allowed {
		t.Fatal("expected first publication reservation to be allowed")
	}

	if err := manager.ReleasePublication(
		context.Background(),
		reservation,
	); err != nil {
		t.Fatalf("release publication: %v", err)
	}

	retryReservation, err := manager.ReservePublication(
		context.Background(),
		providerpolicy.ProviderOurAirports,
		"publication-a",
	)
	if err != nil {
		t.Fatalf("reserve publication retry: %v", err)
	}
	if !retryReservation.Decision.Allowed {
		t.Fatal("expected released publication to be retryable")
	}
	if retryReservation.Token == reservation.Token {
		t.Fatal("expected retry to receive a new reservation token")
	}
}

func TestPublicationLifecycleCommitsSuccessfulPublication(t *testing.T) {
	manager, err := New(nil)
	if err != nil {
		t.Fatalf("create provider budget manager: %v", err)
	}

	reservation, err := manager.ReservePublication(
		context.Background(),
		providerpolicy.ProviderOurAirports,
		"publication-a",
	)
	if err != nil {
		t.Fatalf("reserve publication: %v", err)
	}
	if err := manager.CommitPublication(
		context.Background(),
		reservation,
	); err != nil {
		t.Fatalf("commit publication: %v", err)
	}

	duplicate, err := manager.ReservePublication(
		context.Background(),
		providerpolicy.ProviderOurAirports,
		"publication-a",
	)
	if err != nil {
		t.Fatalf("reserve committed publication: %v", err)
	}
	if duplicate.Decision.Allowed {
		t.Fatal("expected committed publication to be denied")
	}
	if duplicate.Decision.Reason != DecisionReasonPublicationAlreadyProcessed {
		t.Fatalf("unexpected duplicate reason: %s", duplicate.Decision.Reason)
	}
}

func TestPublicationLifecycleAllowsOnlyOneConcurrentReservation(t *testing.T) {
	manager, err := New(nil)
	if err != nil {
		t.Fatalf("create provider budget manager: %v", err)
	}

	const workers = 32
	var waitGroup sync.WaitGroup
	waitGroup.Add(workers)

	decisions := make(chan Decision, workers)
	errorsChannel := make(chan error, workers)
	for index := 0; index < workers; index++ {
		go func() {
			defer waitGroup.Done()
			reservation, reserveErr := manager.ReservePublication(
				context.Background(),
				providerpolicy.ProviderOurAirports,
				"publication-concurrent",
			)
			if reserveErr != nil {
				errorsChannel <- reserveErr
				return
			}
			decisions <- reservation.Decision
		}()
	}
	waitGroup.Wait()
	close(decisions)
	close(errorsChannel)

	for reserveErr := range errorsChannel {
		t.Fatalf("concurrent reserve publication: %v", reserveErr)
	}

	allowed := 0
	inProgress := 0
	for decision := range decisions {
		if decision.Allowed {
			allowed++
			continue
		}
		if decision.Reason == DecisionReasonPublicationInProgress {
			inProgress++
		}
	}
	if allowed != 1 {
		t.Fatalf("allowed reservations = %d, want 1", allowed)
	}
	if inProgress != workers-1 {
		t.Fatalf("in-progress decisions = %d, want %d", inProgress, workers-1)
	}
}

func TestPublicationLifecycleRejectsForeignReservationToken(t *testing.T) {
	manager, err := New(nil)
	if err != nil {
		t.Fatalf("create provider budget manager: %v", err)
	}

	reservation, err := manager.ReservePublication(
		context.Background(),
		providerpolicy.ProviderOurAirports,
		"publication-a",
	)
	if err != nil {
		t.Fatalf("reserve publication: %v", err)
	}
	reservation.Token = "foreign-token"

	err = manager.CommitPublication(context.Background(), reservation)
	if !errors.Is(err, ErrPublicationReservationMismatch) {
		t.Fatalf("expected reservation mismatch, got %v", err)
	}
}
