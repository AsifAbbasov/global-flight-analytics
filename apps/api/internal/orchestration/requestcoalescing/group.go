package requestcoalescing

import (
	"context"
	"errors"
	"strings"
	"sync"
)

var (
	ErrKeyRequired = errors.New(
		"request coalescing key is required",
	)

	ErrFunctionRequired = errors.New(
		"request coalescing function is required",
	)
)

type Function[T any] func(
	ctx context.Context,
) (T, error)

type Result[T any] struct {
	Value  T
	Shared bool
}

type Group[T any] struct {
	mu    sync.Mutex
	calls map[string]*call[T]
}

type call[T any] struct {
	done chan struct{}

	value T
	err   error

	waiters  int
	finished bool
	cancel   context.CancelFunc
}

func New[T any]() *Group[T] {
	return &Group[T]{
		calls: make(map[string]*call[T]),
	}
}

func (group *Group[T]) Do(
	ctx context.Context,
	key string,
	function Function[T],
) (Result[T], error) {
	var zero Result[T]

	normalizedKey := strings.TrimSpace(key)
	if normalizedKey == "" {
		return zero, ErrKeyRequired
	}

	if function == nil {
		return zero, ErrFunctionRequired
	}

	if ctx == nil {
		ctx = context.Background()
	}

	group.mu.Lock()

	if group.calls == nil {
		group.calls = make(
			map[string]*call[T],
		)
	}

	if existingCall, exists := group.calls[normalizedKey]; exists {
		existingCall.waiters++

		group.mu.Unlock()

		return group.wait(
			ctx,
			existingCall,
			true,
		)
	}

	operationContext, cancel := context.WithCancel(
		context.WithoutCancel(ctx),
	)

	newCall := &call[T]{
		done:    make(chan struct{}),
		waiters: 1,
		cancel:  cancel,
	}

	group.calls[normalizedKey] = newCall

	group.mu.Unlock()

	go group.run(
		normalizedKey,
		newCall,
		operationContext,
		function,
	)

	return group.wait(
		ctx,
		newCall,
		false,
	)
}

func (group *Group[T]) run(
	key string,
	currentCall *call[T],
	ctx context.Context,
	function Function[T],
) {
	value, err := function(ctx)

	group.mu.Lock()

	currentCall.value = value
	currentCall.err = err
	currentCall.finished = true

	delete(
		group.calls,
		key,
	)

	close(currentCall.done)
	currentCall.cancel()

	group.mu.Unlock()
}

func (group *Group[T]) wait(
	ctx context.Context,
	currentCall *call[T],
	shared bool,
) (Result[T], error) {
	select {
	case <-currentCall.done:
		return Result[T]{
			Value:  currentCall.value,
			Shared: shared,
		}, currentCall.err

	case <-ctx.Done():
		group.releaseWaiter(
			currentCall,
		)

		var zero Result[T]

		return zero, ctx.Err()
	}
}

func (group *Group[T]) releaseWaiter(
	currentCall *call[T],
) {
	group.mu.Lock()
	defer group.mu.Unlock()

	if currentCall.finished {
		return
	}

	if currentCall.waiters > 0 {
		currentCall.waiters--
	}

	if currentCall.waiters == 0 {
		currentCall.cancel()
	}
}
