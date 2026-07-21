package postgres

import (
	"context"
	"time"
)

const repositoryRollbackTimeout = 5 * time.Second

type repositoryTransactionRollbacker interface {
	Rollback(context.Context) error
}

func rollbackRepositoryTransaction(
	transaction repositoryTransactionRollbacker,
) {
	if transaction == nil {
		return
	}

	rollbackContext, cancel := context.WithTimeout(
		context.Background(),
		repositoryRollbackTimeout,
	)
	defer cancel()

	_ = transaction.Rollback(rollbackContext)
}
