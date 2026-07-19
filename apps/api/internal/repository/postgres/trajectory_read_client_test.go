package postgres

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

type trajectoryReadClientStub struct{}

func (
	trajectoryReadClientStub,
) QueryRow(
	context.Context,
	string,
	...any,
) pgx.Row {
	return nil
}

func (
	trajectoryReadClientStub,
) Query(
	context.Context,
	string,
	...any,
) (pgx.Rows, error) {
	return nil, nil
}

func TestNewTrajectoryReadRepositoryUsesSuppliedClient(
	t *testing.T,
) {
	client := trajectoryReadClientStub{}
	repository := NewTrajectoryReadRepository(client)

	if repository == nil {
		t.Fatal("repository is nil")
	}
	if repository.db != nil {
		t.Fatal("read-only repository unexpectedly owns a pool")
	}
	if repository.trajectoryReadExecutor() == nil {
		t.Fatal("read-only repository lost the supplied client")
	}
}
