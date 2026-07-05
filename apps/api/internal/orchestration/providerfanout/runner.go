package providerfanout

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/ingestionorchestrator"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
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

type Executor interface {
	Execute(
		ctx context.Context,
		provider providerpolicy.Provider,
		requestKey string,
		function ingestionorchestrator.Function,
	) (ingestionorchestrator.ExecuteResult, error)
}

type Task struct {
	ID         string
	Provider   providerpolicy.Provider
	RequestKey string
	Function   ingestionorchestrator.Function
}

type Result struct {
	TaskID     string
	Provider   providerpolicy.Provider
	RequestKey string
	Value      any
	Shared     bool
	Err        error
}

type Runner struct {
	executor Executor
}

func New(
	executor Executor,
) (*Runner, error) {
	if executor == nil {
		return nil, ErrExecutorRequired
	}

	return &Runner{
		executor: executor,
	}, nil
}

func (runner *Runner) Run(
	ctx context.Context,
	tasks []Task,
) ([]Result, error) {
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
		[]Result,
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
			currentTask Task,
			normalizedTaskID string,
		) {
			defer waitGroup.Done()

			executionResult, executionErr := runner.executor.Execute(
				ctx,
				currentTask.Provider,
				currentTask.RequestKey,
				currentTask.Function,
			)

			results[resultIndex] = Result{
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

func validateTaskIDs(
	tasks []Task,
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
