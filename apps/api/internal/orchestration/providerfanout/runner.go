package providerfanout

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/ingestionorchestrator"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/requestcoalescing"
)

var (
	ErrExecutorRequired = errors.New(
		"provider fan-out executor is required",
	)

	ErrTaskIDRequired = errors.New(
		"provider fan-out task identifier is required",
	)

	ErrDuplicateTaskID = errors.New(
		"duplicate provider fan-out task identifier",
	)
)

type Executor[T requestcoalescing.Value] interface {
	Execute(
		ctx context.Context,
		provider providerpolicy.Provider,
		requestKey string,
		function ingestionorchestrator.Function[T],
	) (ingestionorchestrator.ExecuteResult[T], error)
}

type Task[T requestcoalescing.Value] struct {
	ID         string
	Provider   providerpolicy.Provider
	RequestKey string
	Function   ingestionorchestrator.Function[T]
}

type Result[T requestcoalescing.Value] struct {
	TaskID     string
	Provider   providerpolicy.Provider
	RequestKey string
	Value      T
	Shared     bool
	Err        error
}

type Runner[T requestcoalescing.Value] struct {
	executor Executor[T]
}

func New[T requestcoalescing.Value](
	executor Executor[T],
) (*Runner[T], error) {
	if executor == nil {
		return nil, ErrExecutorRequired
	}

	return &Runner[T]{
		executor: executor,
	}, nil
}

func (
	runner *Runner[T],
) Run(
	ctx context.Context,
	tasks []Task[T],
) ([]Result[T], error) {
	if ctx == nil {
		ctx = context.Background()
	}

	normalizedTaskIDs, err := validateTaskIDs(
		tasks,
	)
	if err != nil {
		return nil, err
	}

	results := make(
		[]Result[T],
		len(tasks),
	)

	var waitGroup sync.WaitGroup

	waitGroup.Add(
		len(tasks),
	)

	for index := range tasks {
		task := tasks[index]
		taskID := normalizedTaskIDs[index]

		go func(
			resultIndex int,
			currentTask Task[T],
			normalizedTaskID string,
		) {
			defer waitGroup.Done()

			executionResult, executionErr := runner.executor.Execute(
				ctx,
				currentTask.Provider,
				currentTask.RequestKey,
				currentTask.Function,
			)

			results[resultIndex] = Result[T]{
				TaskID:     normalizedTaskID,
				Provider:   currentTask.Provider,
				RequestKey: currentTask.RequestKey,
				Value:      executionResult.Value,
				Shared:     executionResult.Shared,
				Err:        executionErr,
			}
		}(
			index,
			task,
			taskID,
		)
	}

	waitGroup.Wait()

	return results, nil
}

func validateTaskIDs[T requestcoalescing.Value](
	tasks []Task[T],
) ([]string, error) {
	normalizedTaskIDs := make(
		[]string,
		len(tasks),
	)

	seenTaskIDs := make(
		map[string]struct{},
		len(tasks),
	)

	for index, task := range tasks {
		taskID := strings.TrimSpace(
			task.ID,
		)

		if taskID == "" {
			return nil, fmt.Errorf(
				"%w: index=%d",
				ErrTaskIDRequired,
				index,
			)
		}

		if _, exists := seenTaskIDs[taskID]; exists {
			return nil, fmt.Errorf(
				"%w: %s",
				ErrDuplicateTaskID,
				taskID,
			)
		}

		seenTaskIDs[taskID] = struct{}{}
		normalizedTaskIDs[index] = taskID
	}

	return normalizedTaskIDs, nil
}
