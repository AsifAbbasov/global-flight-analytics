package projectionread

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

type scriptedRow struct {
	values []any
	err    error
}

func (
	row scriptedRow,
) Scan(
	destinations ...any,
) error {
	if row.err != nil {
		return row.err
	}

	return assignScanValues(
		row.values,
		destinations,
	)
}

type scriptedRows struct {
	values [][]any
	index  int
	err    error
	closed bool
}

func (
	rows *scriptedRows,
) Next() bool {
	if rows.closed ||
		rows.index >= len(rows.values) {
		return false
	}

	return true
}

func (
	rows *scriptedRows,
) Scan(
	destinations ...any,
) error {
	if rows.closed ||
		rows.index >= len(rows.values) {
		return fmt.Errorf(
			"no scripted row is available",
		)
	}

	values := rows.values[rows.index]
	rows.index++

	return assignScanValues(
		values,
		destinations,
	)
}

func (
	rows *scriptedRows,
) Err() error {
	return rows.err
}

func (
	rows *scriptedRows,
) Close() {
	rows.closed = true
}

type queryCall struct {
	query string
	args  []any
}

type scriptedClient struct {
	mu sync.Mutex

	rowQueue  []scriptedRow
	rowsQueue []*scriptedRows

	queryRowCalls []queryCall
	queryCalls    []queryCall
}

func (
	client *scriptedClient,
) QueryRow(
	_ context.Context,
	query string,
	args ...any,
) rowScanner {
	client.mu.Lock()
	defer client.mu.Unlock()

	client.queryRowCalls = append(
		client.queryRowCalls,
		queryCall{
			query: query,
			args: append(
				[]any(nil),
				args...,
			),
		},
	)
	if len(client.rowQueue) == 0 {
		return scriptedRow{
			err: fmt.Errorf(
				"unexpected QueryRow call",
			),
		}
	}

	row := client.rowQueue[0]
	client.rowQueue =
		client.rowQueue[1:]

	return row
}

func (
	client *scriptedClient,
) Query(
	_ context.Context,
	query string,
	args ...any,
) (rowIterator, error) {
	client.mu.Lock()
	defer client.mu.Unlock()

	client.queryCalls = append(
		client.queryCalls,
		queryCall{
			query: query,
			args: append(
				[]any(nil),
				args...,
			),
		},
	)
	if len(client.rowsQueue) == 0 {
		return nil,
			fmt.Errorf(
				"unexpected Query call",
			)
	}

	rows := client.rowsQueue[0]
	client.rowsQueue =
		client.rowsQueue[1:]

	return rows,
		nil
}

type trajectoryRepositoryStub struct {
	items map[string]trajectory.FlightTrajectory
	errs  map[string]error
	calls []string
}

func (
	repository *trajectoryRepositoryStub,
) GetTrajectoryByID(
	_ context.Context,
	trajectoryID string,
) (trajectory.FlightTrajectory, error) {
	repository.calls = append(
		repository.calls,
		trajectoryID,
	)
	if err := repository.errs[trajectoryID]; err != nil {
		return trajectory.FlightTrajectory{},
			err
	}
	item, exists :=
		repository.items[trajectoryID]
	if !exists {
		return trajectory.FlightTrajectory{},
			ErrTrajectoryNotFound
	}

	return item,
		nil
}

func assignScanValues(
	values []any,
	destinations []any,
) error {
	if len(values) != len(destinations) {
		return fmt.Errorf(
			"scan value count %d does not match destination count %d",
			len(values),
			len(destinations),
		)
	}

	for index := range values {
		destinationValue :=
			reflect.ValueOf(
				destinations[index],
			)
		if destinationValue.Kind() !=
			reflect.Pointer ||
			destinationValue.IsNil() {
			return fmt.Errorf(
				"scan destination %d is not a non-nil pointer",
				index,
			)
		}

		target :=
			destinationValue.Elem()
		value := reflect.ValueOf(
			values[index],
		)
		if !value.IsValid() {
			target.Set(
				reflect.Zero(
					target.Type(),
				),
			)
			continue
		}
		if value.Type().AssignableTo(
			target.Type(),
		) {
			target.Set(value)
			continue
		}
		if value.Type().ConvertibleTo(
			target.Type(),
		) {
			target.Set(
				value.Convert(
					target.Type(),
				),
			)
			continue
		}

		return fmt.Errorf(
			"scan value %d type %s cannot be assigned to %s",
			index,
			value.Type(),
			target.Type(),
		)
	}

	return nil
}
