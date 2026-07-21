package postgres

import (
	"context"
	"testing"
	"time"
)

type rollbackContextProbe struct {
	called      bool
	contextErr  error
	hasDeadline bool
}

func (probe *rollbackContextProbe) Rollback(ctx context.Context) error {
	probe.called = true
	probe.contextErr = ctx.Err()
	_, probe.hasDeadline = ctx.Deadline()
	return nil
}

func TestRollbackRepositoryTransactionUsesIndependentBoundedContext(
	t *testing.T,
) {
	t.Parallel()

	callerContext, cancelCaller := context.WithCancel(context.Background())
	cancelCaller()
	if callerContext.Err() == nil {
		t.Fatal("caller context was not cancelled")
	}

	probe := &rollbackContextProbe{}
	rollbackRepositoryTransaction(probe)

	if !probe.called {
		t.Fatal("rollback was not called")
	}
	if probe.contextErr != nil {
		t.Fatalf("rollback context was already cancelled: %v", probe.contextErr)
	}
	if !probe.hasDeadline {
		t.Fatal("rollback context has no deadline")
	}
}

func TestRollbackRepositoryTransactionAcceptsNil(t *testing.T) {
	t.Parallel()

	start := time.Now()
	rollbackRepositoryTransaction(nil)
	if time.Since(start) > time.Second {
		t.Fatal("nil rollback unexpectedly blocked")
	}
}
